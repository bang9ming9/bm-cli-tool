package eventlogger

import (
	"context"
	"errors"
	"math"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/bang9ming9/bm-cli-tool/eventlogger/logger"
	"github.com/bang9ming9/bm-cli-tool/eventlogger/logtypes"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Backend interface {
	// utils.Backend
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
}

type LoggerServer struct {
	logger.UnimplementedLoggerServer
	logger.UnimplementedAdminServer
	logger *logrus.Entry

	client     Backend
	collection *mongo.Collection

	qlock   sync.RWMutex
	addrSet map[common.Address]struct{}
	query   ethereum.FilterQuery

	scanBlock uint64
	stopBlock uint64
	scanStop  chan struct{}

	slock   sync.Mutex
	streams map[common.Address]*struct {
		lock      sync.Mutex
		idCounter uint32
		clients   map[uint32]struct {
			sss grpc.ServerStreamingServer[logger.Log]
			err chan error
		}
	}
}

func NewLoggerServer(stopCh <-chan os.Signal, addr string, log *logrus.Logger, client Backend, collection *mongo.Collection, query *ethereum.FilterQuery) error {
	// 입력값 확인
	if stopCh == nil {
		return errors.New("stop channel is nil")
	}
	if addr == "" {
		return errors.New("addr is not set")
	} else {
		split := strings.Split(addr, ":")
		if len(split) != 2 {
			return errors.New("invalid addr require <ip:port>")
		} else if port, err := strconv.Atoi(split[1]); err != nil {
			return errors.New("invalid open port: is not number")
		} else if port < 1000 {
			return errors.New("invalid open port: require 'port >= 1000'")
		}
	}
	if log == nil {
		return errors.New("logrus is nil")
	}
	if client == nil {
		return errors.New("chain client is nil")
	}
	if collection == nil {
		return errors.New("mongo collection is nil")
	}
	// 입력값 확인 끝
	logentry := log.WithField("module", "LoggerServer")
	// gRPC 서버 Open
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	server := &LoggerServer{
		logger: logentry,

		client:     client,
		collection: collection,

		// qlock: sync.RWMutex{},
		addrSet: make(map[common.Address]struct{}),
		// query: ethereum.FilterQuery{},

		// scanBlock: 0,
		// stopBlock: 0,
		scanStop: make(chan struct{}),

		slock: sync.Mutex{},
		streams: make(map[common.Address]*struct {
			lock      sync.Mutex
			idCounter uint32
			clients   map[uint32]struct {
				sss grpc.ServerStreamingServer[logger.Log]
				err chan error
			}
		}),
	}

	if query == nil {
		logentry.Warn("filterquery is not set, waiting for scan start")
	} else {
		server.query = *query
		server.slock.Lock()
		for _, address := range server.query.Addresses {
			server.addrSet[address] = struct{}{}
		}
		server.slock.Unlock()
		if err := server.start(query.FromBlock.Uint64()); err != nil {
			return err
		}
	}

	logentry.Info("Starting gRPC server on ", addr)
	logger.RegisterLoggerServer(s, server)
	logger.RegisterAdminServer(s, server)

	go func() {
		<-stopCh
		logentry.Warn("Quit...")
		server.quit()
		s.Stop()
	}()

	return s.Serve(listener)
}

// ///////////////////
// Public Procedure //
// ///////////////////

func (s *LoggerServer) Info(context.Context, *emptypb.Empty) (*logger.InfoResMessage, error) {
	s.logger.Debug("Info")

	addresses := make([][]byte, 0, len(s.addrSet))
	for a := range s.addrSet {
		addresses = append(addresses, a.Bytes())
	}

	return &logger.InfoResMessage{Address: addresses}, nil
}

func (s *LoggerServer) Connect(req *logger.ConnectReqMessage, stream grpc.ServerStreamingServer[logger.Log]) error {
	s.logger.WithField("req", req).Trace("Connect")
	address := common.BytesToAddress(req.Address)
	logentry := s.logger.WithFields(logrus.Fields{
		"address": address.Hex(),
		"from":    req.FromBlock,
	})
	logentry.Debug("Connect")
	if _, ok := s.addrSet[address]; !ok {
		return status.Error(codes.InvalidArgument, "invalid address")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if req.FromBlock != 0 {
		fh, fl := logtypes.SplitUint64(req.FromBlock)
		var filter bson.D = bson.D{
			{Key: "address", Value: address},
			{Key: "raw.block_number_high", Value: bson.D{{Key: "$gte", Value: fh}}},
			{Key: "raw.block_number_low", Value: bson.D{{Key: "$gte", Value: fl}}},
		}
		cursor, err := s.collection.Find(ctx, filter)
		if err != nil {
			if !(errors.Is(err, mongo.ErrNoDocuments) || errors.Is(err, mongo.ErrNilDocument)) {
				return status.Error(codes.Unavailable, "fail to find logs")
			}
		} else {
			defer cursor.Close(ctx)
			for cursor.Next(ctx) {
				var result bson.M
				if err := cursor.Decode(&result); err != nil {
					return status.Error(codes.Internal, err.Error())
				} else {
					if err := stream.Send(logtypes.LogToProtobuf(logtypes.LogFromBsonM(result))); err != nil {
						logentry.WithField("message", err.Error()).Error("stream send error")
						return status.Error(codes.Internal, err.Error())
					}
				}
			}
		}
	}

	close, err := s.addClient(address, stream)
	defer close()

	select {
	case <-stream.Context().Done():
		return nil
	case e := <-err:
		return status.Error(codes.Unknown, e.Error())
	}
}

func (s *LoggerServer) addClient(address common.Address, client grpc.ServerStreamingServer[logger.Log]) (func(), <-chan error) {
	s.slock.Lock()
	if _, ok := s.streams[address]; !ok {
		s.streams[address] = &struct {
			lock      sync.Mutex
			idCounter uint32
			clients   map[uint32]struct {
				sss grpc.ServerStreamingServer[logger.Log]
				err chan error
			}
		}{sync.Mutex{}, 0, make(map[uint32]struct {
			sss grpc.ServerStreamingServer[logger.Log]
			err chan error
		})}
	}
	stream := s.streams[address]
	s.slock.Unlock()

	stream.lock.Lock()
	defer stream.lock.Unlock()
	stream.idCounter++
	id, err := stream.idCounter, make(chan error)
	stream.clients[id] = struct {
		sss grpc.ServerStreamingServer[logger.Log]
		err chan error
	}{client, err}

	return func() {
		close(err)

		stream.lock.Lock()
		defer stream.lock.Unlock()
		delete(stream.clients, id)
	}, err
}

// //////////////////
// Admin Procedure //
// //////////////////

func (s *LoggerServer) Add(ctx context.Context, req *logger.AddressReqMessage) (*logger.BlockNumberMessage, error) {
	s.logger.WithField("req", req).Trace("Add")
	address := common.BytesToAddress(req.Address)
	logentry := s.logger.WithFields(logrus.Fields{
		"address": address,
	})
	logentry.Debug("Add")

	s.qlock.Lock()
	defer s.qlock.Unlock()

	if _, ok := s.addrSet[address]; ok {
		return nil, status.Error(codes.AlreadyExists, address.Hex())
	} else {
		s.addrSet[address] = struct{}{}
		addresses := make([]common.Address, 0, len(s.addrSet))
		for a := range s.addrSet {
			addresses = append(addresses, a)
		}
		s.query.Addresses = addresses
	}

	return &logger.BlockNumberMessage{
		BlockNumber: s.scanBlock,
	}, nil
}

func (s *LoggerServer) Remove(ctx context.Context, req *logger.AddressReqMessage) (*logger.BlockNumberMessage, error) {
	s.logger.WithField("req", req).Trace("Remove")
	address := common.BytesToAddress(req.Address)
	logentry := s.logger.WithFields(logrus.Fields{
		"address": address,
	})
	logentry.Debug("Remove")

	s.qlock.Lock()
	defer s.qlock.Unlock()

	if _, ok := s.addrSet[address]; !ok {
		return nil, status.Error(codes.InvalidArgument, "unknown address")
	} else {
		delete(s.addrSet, address)
		s.addrSet[address] = struct{}{}
		addresses := make([]common.Address, 0, len(s.addrSet))
		for a := range s.addrSet {
			addresses = append(addresses, a)
		}
		s.query.Addresses = addresses
	}

	return &logger.BlockNumberMessage{
		BlockNumber: s.scanBlock,
	}, nil
}

func (s *LoggerServer) Start(ctx context.Context, req *logger.BlockNumberMessage) (*emptypb.Empty, error) {
	s.logger.WithField("req", req).Trace("Start")
	// s.scanBlock 는 New...() 또는 Stop() 에서 종료가 완료되면 0으로 설정된다.
	if s.scanBlock != 0 {
		err := status.Errorf(codes.Aborted, "already started %v ...", s.scanBlock)
		s.logger.WithField("message", err.Error()).Error("Start")
		return nil, err
	}
	s.logger.WithField("block-number", req.BlockNumber).Debug("Upsert")

	return nil, s.start(req.BlockNumber)
}

func (s *LoggerServer) Stop(ctx context.Context, _ *emptypb.Empty) (*logger.BlockNumberMessage, error) {
	s.stop()

	return &logger.BlockNumberMessage{
		BlockNumber: s.stopBlock,
	}, nil
}

// //////////////
// Code Reduce //
// //////////////

func (s *LoggerServer) filterQuery() ethereum.FilterQuery {
	s.qlock.RLock()
	defer s.qlock.RUnlock()

	return s.query
}

func (s *LoggerServer) start(startBlock uint64) error {
	s.stopBlock = math.MaxUint64

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	newHead := make(chan *types.Header)
	sub, err := s.client.SubscribeNewHead(ctx, newHead)
	if err != nil {
		return err
	}

	{
		rawBlock := struct {
			Raw struct {
				High int64 `bson:"block_number_high"`
				Low  int64 `bson:"block_number_low"`
			} `bson:"raw"`
		}{}
		err = s.collection.FindOne(ctx, bson.D{}, options.FindOne().SetSort(bson.D{{Key: "raw.block_number_high", Value: -1}, {Key: "raw.block_number_low", Value: -1}})).Decode(&rawBlock)
		var latestBlock uint64 = 0
		if err != nil {
			if !(errors.Is(err, mongo.ErrNoDocuments) || errors.Is(err, mongo.ErrNilDocument)) {
				return err
			}
		} else {
			latestBlock = logtypes.JoinUint64(rawBlock.Raw.High, rawBlock.Raw.Low)
		}
		// 스캔이 시작될때 +1 을 하기때문에 startBlock-1 계산
		s.scanBlock = max(latestBlock, startBlock-1) // DB 에 저장된 값, 입력된 값-1 중에 큰 값을 사용한다.
	}

	go func() {
		defer func() { s.scanStop <- struct{}{} }()
		// s.stopBlock 는 start() 가 시작할때 max(uint64), stop() 에서 s.scanBlock+1 으로 설정된다.
		for s.scanBlock < s.stopBlock {
			select {
			case err := <-sub.Err():
				s.logger.WithFields(logrus.Fields{
					"message":    err.Error(),
					"scan-block": s.scanBlock,
				}).Panic("err subscribe new head")
			case head := <-newHead:
				func(number uint64) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					documents := []interface{}{}
					for s.scanBlock < number {
						s.scanBlock++
						block := new(big.Int).SetUint64(s.scanBlock)

						filter := s.filterQuery()
						filter.FromBlock, filter.ToBlock = block, block
						logentry := s.logger.WithFields(logrus.Fields{
							"block-number": s.scanBlock,
							"filter-query": filter,
						})
						logentry.Trace()

						logs, err := s.client.FilterLogs(ctx, filter)
						if err != nil {
							logentry.WithField("message", err.Error()).Panic("fail to call filter logs")
						}
						documents = append(documents, logtypes.LogsToBson(logs)...)
						for _, log := range logs {
							logentry.WithFields(logrus.Fields{
								"address": log.Address,
								"eventid": log.Topics[0],
							}).Debug("filter log")
							if stream, ok := s.streams[log.Address]; ok {
								stream.lock.Lock()
								for _, c := range stream.clients {
									if err := c.sss.Send(logtypes.LogToProtobuf(log)); err != nil {
										c.err <- err
									}
								}
								stream.lock.Unlock()
							}
						}
					}
					if len(documents) != 0 {
						if _, err := s.collection.InsertMany(ctx, documents); err != nil {
							s.logger.WithFields(logrus.Fields{
								"document-count": len(documents),
								"message":        err.Error(),
							}).Error("fail to insert documents")
						}
					}
				}(head.Number.Uint64())
			}
		}
	}()

	return nil
}

func (s *LoggerServer) stop() {
	s.logger.WithField("scan-block", s.scanBlock).Trace("Stop")
	s.stopBlock = s.scanBlock + 1
	<-s.scanStop
	s.stopBlock = max(s.scanBlock, s.stopBlock)
	s.logger.WithField("stop-block", s.stopBlock).Debug("Stop")
	s.scanBlock = 0
}

func (s *LoggerServer) quit() {
	if s.scanBlock != 0 {
		s.stop()
	}
	close(s.scanStop)
}
