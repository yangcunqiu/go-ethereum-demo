package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"log"
	"math/big"
	"strings"
)

func ParseEvent() {
	url := "https://polygon-mumbai.g.alchemy.com/v2/vAJ51DAbtpol-pyDks4dFlGeauFAFuOT"

	client, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("failed to connect to %s, err: %v", url, err)
	}

	// 读取 ABI 文件并解析
	abiBytes, err := ioutil.ReadFile("C:\\project\\go\\go-ethereum-demo\\demo\\HL.json")
	if err != nil {
		log.Fatalf("Failed to read ABI file: %v", err)
	}

	contractABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}

	contractAddress := common.HexToAddress("0x672465cbFa4306aDfaeDD1753Dfa036aBe08c4A6")

	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(34708174),
		ToBlock:   new(big.Int).SetUint64(35010304),
		Addresses: []common.Address{contractAddress},
		Topics:    [][]common.Hash{{crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))}},
	}
	log.Println(crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)")))

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatalf("Failed to get historical contract events: %v", err)
	}

	for _, vLog := range logs {
		event := struct {
			From  common.Address
			To    common.Address
			Value *big.Int
		}{}
		//err := contractABI.UnpackIntoInterface(&event, "Transfer", vLog.Data)
		//map1 := make(map[string]interface{})
		json, err := vLog.MarshalJSON()
		log.Println(string(json))
		err = contractABI.UnpackIntoInterface(&event, "Transfer", vLog.Data)
		//err := contractABI.UnpackIntoMap(map1, "Transfer", vLog.Data)
		if err != nil {
			log.Fatalf("Failed to unpack event data: %v", err)
		}

		fmt.Printf("Historical event: From: %s, To %s, Value: %s\n", common.BytesToAddress(vLog.Topics[1].Bytes()), common.BytesToAddress(vLog.Topics[2].Bytes()), event.Value.String())
		fmt.Printf("topic %s\n", common.BytesToHash(vLog.Topics[1].Bytes()))
		fmt.Printf("111 %s\n", common.BytesToAddress(vLog.Topics[1].Bytes()))
		fmt.Printf("m %s\n", vLog.Topics[1].String())

		//fmt.Printf("Historical event: %v\n", map1)
	}
}
