package scan

import (
	"context"
	"reflect"

	"github.com/bang9ming9/bm-cli-tool/eventlogger/logger"
	"github.com/bang9ming9/bm-cli-tool/eventlogger/logtypes"
	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type IScanner interface {
	Scan(ctx context.Context, client logger.LoggerClient, fromBlock uint64, tx chan<- func(db *gorm.DB) error) error
}

type Scanner struct {
	address  common.Address
	abi      *abi.ABI
	types    map[common.Hash]reflect.Type // event.ID => EventType
	logentry *logrus.Entry
}

func (s *Scanner) Scan(ctx context.Context, client logger.LoggerClient, fromBlock uint64, tx chan<- func(db *gorm.DB) error) error {
	stream, err := client.Connect(ctx, &logger.ConnectReqMessage{
		FromBlock: fromBlock,
		Address:   s.address.Bytes(),
	})
	if err != nil {
		return err
	}
	go func() {
		for {
			recv, err := stream.Recv()
			if err != nil {
				s.logentry.WithField("recv", recv).Error(err.Error())
				tx <- func(_ *gorm.DB) error { return err }
			}
			if recv == nil {
				continue
			}
			log := logtypes.LogFromProtobuf(recv)
			logentry := s.logentry.WithField("log", log)

			out, err := parse(log, s.types[log.Topics[0]], s.abi)
			if err != nil {
				if errors.Is(err, ErrNonTargetedEvent) {
					logentry.Warn(err.Error())
				} else {
					logentry.Error(err.Error())
				}
			} else {
				tx <- out.Do(log)
			}
		}
	}()
	return nil
}

// /////////
// Common //
// /////////

var (
	ErrNoEventSignature       = errors.New("no event signature")
	ErrInvalidEventID         = errors.New("invalid eventID exists")
	ErrEventSignatureMismatch = errors.New("event signature mismatch")
	ErrNonTargetedEvent       = errors.New("non-targeted event")
)

func parse(log types.Log, outType reflect.Type, aBI *abi.ABI) (dbtypes.IRecord, error) {
	// Anonymous events are not supported.
	if len(log.Topics) == 0 {
		return nil, ErrNoEventSignature
	}
	event, err := aBI.EventByID(log.Topics[0])
	if err != nil {
		return nil, err
	}
	if outType == nil {
		return nil, errors.Wrap(ErrNonTargetedEvent, event.Name)
	}

	out := reflect.New(outType).Interface()
	if len(log.Data) > 0 {
		if err := aBI.UnpackIntoInterface(out, event.Name, log.Data); err != nil {
			return nil, errors.Wrap(err, event.Name)
		}
	}

	var indexed abi.Arguments
	for _, arg := range event.Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return out.(dbtypes.IRecord), abi.ParseTopics(out, indexed, log.Topics[1:])
}
