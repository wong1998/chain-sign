package ton

import (
	"strconv"

	"crypto/ed25519"

	"github.com/ethereum/go-ethereum/log"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton/wallet"

	"github.com/wong1998/chain-sign/chain"
	"github.com/wong1998/chain-sign/config"
	"github.com/wong1998/chain-sign/rpc/account"
	"github.com/wong1998/chain-sign/rpc/common"
)

const ChainName = "Ton"

type ChainAdaptor struct {
	tonClient     *TonClient
	tonDataClient *TonDataClient
}

func NewChainAdaptor(conf *config.Config) (chain.IChainAdaptor, error) {
	tonClient, err := NewTonClients(conf)
	if err != nil {
		log.Error("new ton client fail", "err", err)
		return nil, err
	}

	tonDataClient, err := NewTonDataClient(conf.WalletNode.Ton.DataApiUrl)
	if err != nil {
		log.Error("new ton data client fail", "err", err)
		return nil, err
	}

	return &ChainAdaptor{
		tonClient:     tonClient,
		tonDataClient: tonDataClient,
	}, nil
}

func (c *ChainAdaptor) GetSupportChains(req *account.SupportChainsRequest) (*account.SupportChainsResponse, error) {
	return &account.SupportChainsResponse{
		Code:    common.ReturnCode_SUCCESS,
		Msg:     "Support this chain",
		Support: true,
	}, nil
}

func (c *ChainAdaptor) ConvertAddress(req *account.ConvertAddressRequest) (*account.ConvertAddressResponse, error) {
	addr, err := wallet.AddressFromPubKey(ed25519.PublicKey(req.PublicKey), req.Type, 0)
	if err != nil {
		return &account.ConvertAddressResponse{
			Code: common.ReturnCode_ERROR,
			Msg:  err.Error(),
		}, nil
	} else {
		return &account.ConvertAddressResponse{
			Code:    common.ReturnCode_SUCCESS,
			Msg:     "convert address successs",
			Address: addr.String(),
		}, nil
	}
}

// ValidAddress 验证地址
func (c *ChainAdaptor) ValidAddress(req *account.ValidAddressRequest) (*account.ValidAddressResponse, error) {
	_, err := address.ParseAddr(req.Address)
	return &account.ValidAddressResponse{
		Code:  common.ReturnCode_SUCCESS,
		Msg:   "convert address successs",
		Valid: err == nil,
	}, nil
}

func (c *ChainAdaptor) GetBlockByNumber(req *account.BlockNumberRequest) (*account.BlockResponse, error) {
	return &account.BlockResponse{
		Code: common.ReturnCode_SUCCESS,
		Msg:  "Do not support this rpc interface",
	}, nil
}

func (c *ChainAdaptor) GetBlockByHash(req *account.BlockHashRequest) (*account.BlockResponse, error) {
	return &account.BlockResponse{
		Code: common.ReturnCode_SUCCESS,
		Msg:  "Do not support this rpc interface",
	}, nil
}

func (c *ChainAdaptor) GetBlockHeaderByHash(req *account.BlockHeaderHashRequest) (*account.BlockHeaderResponse, error) {
	return &account.BlockHeaderResponse{
		Code: common.ReturnCode_SUCCESS,
		Msg:  "Do not support this rpc interface",
	}, nil
}

func (c *ChainAdaptor) GetBlockHeaderByNumber(req *account.BlockHeaderNumberRequest) (*account.BlockHeaderResponse, error) {
	return &account.BlockHeaderResponse{
		Code: common.ReturnCode_SUCCESS,
		Msg:  "Do not support this rpc interface",
	}, nil
}

func (c *ChainAdaptor) GetAccount(req *account.AccountRequest) (*account.AccountResponse, error) {
	balance, nonce, err := c.tonClient.GetAccountInfo(req.Address)
	if err != nil {
		return &account.AccountResponse{
			Code: common.ReturnCode_ERROR,
			Msg:  "get account info error",
		}, err
	}

	sequence := strconv.FormatUint(nonce, 10)
	return &account.AccountResponse{
		Code:     common.ReturnCode_SUCCESS,
		Msg:      "get account info success",
		Balance:  balance,
		Sequence: sequence,
	}, nil

}

func (c *ChainAdaptor) GetFee(req *account.FeeRequest) (*account.FeeResponse, error) {
	ret, err := c.tonDataClient.GetEstimateFee(req.Address, req.RawTx)
	if err != nil {
		return &account.FeeResponse{
			Code: common.ReturnCode_SUCCESS,
			Msg:  "get fee fail",
		}, err
	}

	normalFee := ret.InFwdFee + ret.StorageFee + ret.GasFee + ret.FwdFee
	return &account.FeeResponse{
		Code:      common.ReturnCode_SUCCESS,
		Msg:       "get fee success",
		NormalFee: strconv.FormatUint(normalFee, 10),
	}, nil
}

func (c *ChainAdaptor) SendTx(req *account.SendTxRequest) (*account.SendTxResponse, error) {
	hash, err := c.tonDataClient.PostSendTx(req.RawTx)
	if err != nil {
		log.Error("send transaction fail", "err", err)
		return nil, err
	}
	return &account.SendTxResponse{
		Code:   common.ReturnCode_SUCCESS,
		Msg:    "success",
		TxHash: hash,
	}, nil
}

func (c *ChainAdaptor) GetTxByAddress(req *account.TxAddressRequest) (*account.TxAddressResponse, error) {
	ret, err := c.tonDataClient.GetTxByAddr(req.Address, uint64(req.Page), uint64(req.Pagesize))
	if err != nil {
		return nil, err
	}
	var txList []*account.TxMessage
	for _, transactionInfo := range ret.Transactions {
		txMessage, blockRespErr := ParseTxMessage(ret, &transactionInfo)
		if blockRespErr != nil {
			return &account.TxAddressResponse{
				Code: common.ReturnCode_ERROR,
				Msg:  "get transactions fail",
			}, blockRespErr
		}
		txList = append(txList, txMessage)
	}
	return &account.TxAddressResponse{
		Code: common.ReturnCode_SUCCESS,
		Msg:  "get transactions fail",
		Tx:   txList,
	}, nil
}

func (c *ChainAdaptor) GetTxByHash(req *account.TxHashRequest) (*account.TxHashResponse, error) {
	ret, err := c.tonDataClient.GetTxByTxHash(req.Hash)
	if err != nil {
		log.Error("get transaction by hash fail", "err", err)
		return nil, err
	}
	if len(ret.Transactions) == 0 {
		return &account.TxHashResponse{
			Code: common.ReturnCode_ERROR,
			Msg:  "get transactions fail",
		}, nil
	}

	tx := ret.Transactions[0]
	txMsg, _ := ParseTxMessage(ret, &tx)
	res := &account.TxHashResponse{
		Code: common.ReturnCode_SUCCESS,
		Msg:  "get transaction by hash success",
		Tx:   txMsg,
	}
	return res, nil
}

func (c *ChainAdaptor) GetBlockByRange(req *account.BlockByRangeRequest) (*account.BlockByRangeResponse, error) {
	return &account.BlockByRangeResponse{
		Code: common.ReturnCode_SUCCESS,
		Msg:  "Do not support this rpc interface",
	}, nil
}

func (c *ChainAdaptor) CreateUnSignTransaction(req *account.UnSignTransactionRequest) (*account.UnSignTransactionResponse, error) {
	return &account.UnSignTransactionResponse{
		Code: common.ReturnCode_SUCCESS,
		Msg:  "Do not support this rpc interface",
	}, nil
}

func (c *ChainAdaptor) BuildSignedTransaction(req *account.SignedTransactionRequest) (*account.SignedTransactionResponse, error) {
	return &account.SignedTransactionResponse{
		Code: common.ReturnCode_SUCCESS,
		Msg:  "Do not support this rpc interface",
	}, nil
}

func (c *ChainAdaptor) DecodeTransaction(req *account.DecodeTransactionRequest) (*account.DecodeTransactionResponse, error) {
	return &account.DecodeTransactionResponse{
		Code: common.ReturnCode_SUCCESS,
		Msg:  "Do not support this rpc interface",
	}, nil
}

func (c *ChainAdaptor) VerifySignedTransaction(req *account.VerifyTransactionRequest) (*account.VerifyTransactionResponse, error) {
	return &account.VerifyTransactionResponse{
		Code: common.ReturnCode_SUCCESS,
		Msg:  "Do not support this rpc interface",
	}, nil
}

func (c *ChainAdaptor) GetExtraData(req *account.ExtraDataRequest) (*account.ExtraDataResponse, error) {
	return &account.ExtraDataResponse{
		Code:  common.ReturnCode_ERROR,
		Msg:   "Do not support this api",
		Value: req.Chain,
	}, nil
}
