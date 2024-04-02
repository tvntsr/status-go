package bridge

import (
	"math/big"

	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/services/wallet/token"
	"github.com/status-im/status-go/transactions"
)

const AddressLength = 20
type Address [AddressLength]byte
type HexBytes []byte
const HashLength = 32
type Hash [HashLength]byte

const IncreaseEstimatedGasFactor = 1.1

func getSigner(chainID uint64, from  Address, verifiedAccount *account.SelectedExtKey) bind.SignerFn {
	return func(addr common.Address, tx *ethTypes.Transaction) (*ethTypes.Transaction, error) {
		s := ethTypes.NewLondonSigner(new(big.Int).SetUint64(chainID))
		return ethTypes.SignTx(tx, s, verifiedAccount.AccountKey.PrivateKey)
	}
}

type TransactionBridge struct {
	BridgeName        string
	ChainID           uint64
	TransferTx        *transactions.SendTxArgs
	HopTx             *HopTxArgs
	CbridgeTx         *CBridgeTxArgs
	ERC721TransferTx  *ERC721TransferTxArgs
	ERC1155TransferTx *ERC1155TransferTxArgs
}

func (t *TransactionBridge) Value() *big.Int {
	if t.TransferTx != nil && t.TransferTx.To != nil {
		return t.TransferTx.Value.ToInt()
	} else if t.HopTx != nil {
		return t.HopTx.Amount.ToInt()
	} else if t.CbridgeTx != nil {
		return t.CbridgeTx.Amount.ToInt()
	} else if t.ERC721TransferTx != nil {
		return big.NewInt(1)
	} else if t.ERC1155TransferTx != nil {
		return t.ERC1155TransferTx.Amount.ToInt()
	}

	return big.NewInt(0)
}

func (t *TransactionBridge) From()  Address {
	if t.TransferTx != nil && t.TransferTx.To != nil {
		return Address(t.TransferTx.From)
	} else if t.HopTx != nil {
		return Address(t.HopTx.From)
	} else if t.CbridgeTx != nil {
		return Address(t.CbridgeTx.From)
	} else if t.ERC721TransferTx != nil {
		return Address(t.ERC721TransferTx.From)
	} else if t.ERC1155TransferTx != nil {
		return Address(t.ERC1155TransferTx.From)
	}

	return Address{}
}

func (t *TransactionBridge) To()  Address {
	if t.TransferTx != nil && t.TransferTx.To != nil {
		return Address(*t.TransferTx.To)
	} else  if t.HopTx != nil {
		return Address(t.HopTx.Recipient)
	} else if t.CbridgeTx != nil {
		return Address(t.HopTx.Recipient)
	} else if t.ERC721TransferTx != nil {
		return Address(t.ERC721TransferTx.Recipient)
	} else if t.ERC1155TransferTx != nil {
		return  Address(t.ERC1155TransferTx.Recipient)
	}

	return Address{}
}

func (t *TransactionBridge) Data()  HexBytes {
	if t.TransferTx != nil && t.TransferTx.To != nil {
		return HexBytes(t.TransferTx.Data)
	} else if t.HopTx != nil {
		return  HexBytes("")
	} else if t.CbridgeTx != nil {
		return  HexBytes("")
	} else if t.ERC721TransferTx != nil {
		return  HexBytes("")
	} else if t.ERC1155TransferTx != nil {
		return  HexBytes("")
	}

	return  HexBytes("")
}

type Bridge interface {
	Name() string
	Can(from *params.Network, to *params.Network, token *token.Token, balance *big.Int) (bool, error)
	CalculateFees(from, to *params.Network, token *token.Token, amountIn *big.Int, nativeTokenPrice, tokenPrice float64, gasPrice *big.Float) (*big.Int, *big.Int, error)
	EstimateGas(fromNetwork *params.Network, toNetwork *params.Network, from common.Address, to common.Address, token *token.Token, amountIn *big.Int) (uint64, error)
	CalculateAmountOut(from, to *params.Network, amountIn *big.Int, symbol string) (*big.Int, error)
	Send(sendArgs *TransactionBridge, verifiedAccount *account.SelectedExtKey) ( Hash, error)
	GetContractAddress(network *params.Network, token *token.Token) *common.Address
	BuildTransaction(sendArgs *TransactionBridge) (*ethTypes.Transaction, error)
}
