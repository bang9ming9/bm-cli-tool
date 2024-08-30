package deploy_test

import (
	"context"
	"testing"
	"time"

	"github.com/bang9ming9/bm-cli-tool/deploy"
	"github.com/bang9ming9/go-hardhat/bms"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	backend := bms.NewBacked(t)

	stop := make(chan struct{})
	ticker := time.NewTicker(1e9)

	// 1초마다 블럭 생성
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				backend.Commit()
			}
		}
	}()
	require.NoError(t,
		deploy.Deploy(context.Background(), backend.Client, backend.Owner),
	)
	close(stop)
}
