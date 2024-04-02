package contracts

import (
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/contracts/balancechecker"
)

type ContractMaker struct {
    RPCClient *rpc.Client
}

func (c *ContractMaker) NewBalanceChecker(chainID uint64) (*balancechecker.BalanceChecker, error) {
        contractAddr, err := balancechecker.ContractAddress(chainID)
        if err != nil {
                return nil, err
        }
        
        backend, err := c.RPCClient.EthClient(chainID)
        if err != nil {
                return nil, err
        }
        return balancechecker.NewBalanceChecker(
                contractAddr,
                backend,
        )
}
