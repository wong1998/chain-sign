package internet_computer

import (
	"context"

	"github.com/ethereum/go-ethereum/log"
	"github.com/wong1998/chain-sign/config"
)

type ChainAdaptor struct {
	icpClient *IcpClient
}

func NewChainAdaptor(conf *config.Config) (*ChainAdaptor, error) {
	icpClient, err := NewIcpClient(context.Background(), conf.WalletNode.Icp.RpcUrl, conf.WalletNode.Icp.TimeOut)
	if err != nil {
		log.Error("new icp client failed:", err)
		return nil, err
	}

	return &ChainAdaptor{
		icpClient: icpClient,
	}, nil
}
