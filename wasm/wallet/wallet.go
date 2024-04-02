package main

import "github.com/status-im/status-go/services/wallet"

func NewWalletService() (*wallet.Service, error) {
	return wallet.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil), nil

}

func StartWalletService(w *wallet.Service) error{
	return w.Start()
}

func StopWalletService(w *wallet.Service) error {
	return w.Stop()

}

func main() {}

