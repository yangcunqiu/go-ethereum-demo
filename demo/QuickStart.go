package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
)

func QuickStart() {
	url := "http://127.0.0.1:7545"

	dial, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("failed to connect to %s, err: %v", url, err)
	}

	address := "0x79Bd8A4f61D71329585FFddfAAF99163aAA32320"

	ctx := context.Background()
	balance, err := dial.BalanceAt(ctx, common.HexToAddress(address), nil)
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}

	fmt.Printf("The balance of address %s is: %s\n", address, balance.String())
}
