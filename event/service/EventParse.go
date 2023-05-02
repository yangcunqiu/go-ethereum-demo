package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"go-ethereum-demo/event/global"
	"go-ethereum-demo/event/model"
	"go-ethereum-demo/event/model/dao"
	"log"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type EventParseInfo struct {
	eventName   string
	eventHash   common.Hash
	eventParams []abi.ArgumentMarshaling
}

func CreateEventParsePlan(c *gin.Context) {
	var plan dao.EventParsePlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		log.Printf(fmt.Sprintf("参数解析失败, err: %v", err))
		c.JSON(http.StatusOK, model.ApiResult{
			Code:    10001,
			Message: "参数解析失败",
		})
		return
	}
	dao.CreatePlan(&plan)
	listenEvent(c, &plan)
}

func listenEvent(c *gin.Context, plan *dao.EventParsePlan) {
	// 建立和节点的连接
	client := global.ClientMap[plan.NetworksUrl]
	if client == nil {
		var err error // 首先声明 err 变量
		client, err = ethclient.Dial(plan.NetworksUrl)
		if err != nil {
			c.JSON(http.StatusOK, model.ApiResult{
				Code:    10002,
				Message: fmt.Sprintf("节点url错误, err: %v", err),
			})
			return
		}
		global.ClientMap[plan.NetworksUrl] = client
	}
	networkId, err := client.NetworkID(context.Background())
	if err != nil {
		log.Printf("Failed to get chain ID: %v", err)
	}

	// 解析事件声明
	eventParseInfo := parseEvent(plan.EventString)

	// 获取当前最新区块号
	header, _ := client.HeaderByNumber(context.Background(), nil)

	// 获取历史事件
	go getHistoryEvent(client, plan, eventParseInfo, header.Number, networkId)

	// 实时监听新事件
	go listen(client, plan, eventParseInfo, networkId)
}

func listen(client *ethclient.Client, plan *dao.EventParsePlan, eventParseInfo map[common.Hash]*EventParseInfo, networkId *big.Int) {
	eventHashList := make([]common.Hash, 0)
	for hash := range eventParseInfo {

		eventHashList = append(eventHashList, hash)
	}
	// 过滤条件
	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(plan.ContractAddress)},
		Topics:    [][]common.Hash{eventHashList},
	}

	// 实时监听新产生的 Transfer 事件
	logsCh := make(chan types.Log)

	sub, err := client.SubscribeFilterLogs(context.Background(), query, logsCh)
	if err != nil {
		log.Fatalf("Failed to subscribe to new contract events: %v", err)
	}
	defer sub.Unsubscribe()
	for {
		select {
		case err := <-sub.Err():
			log.Fatalf("Subscription error: %v", err)
		case vLog := <-logsCh:
			// 创建一个参数类型列表，根据事件声明
			args := eventParseInfo[vLog.Topics[0]]
			noIndexedParams := make([]abi.ArgumentMarshaling, 0)
			for _, arg := range args.eventParams {
				if !arg.Indexed {
					noIndexedParams = append(noIndexedParams, arg)
				}
			}
			argTypes, _ := abi.NewType("tuple", "", noIndexedParams)
			// 根据类型和日志数据解析事件参数
			eventArgs := abi.Arguments{abi.Argument{Type: argTypes}}
			resultMap := make(map[string]interface{})
			err := eventArgs.UnpackIntoMap(resultMap, vLog.Data)
			if err != nil {
				log.Println(err)
			}
			i := 0
			for _, arg := range args.eventParams {
				if arg.Indexed {
					if arg.Type == "address" {
						resultMap[arg.Name] = common.BytesToAddress(vLog.Topics[i].Bytes())
					} else {
						resultMap[arg.Name] = vLog.Topics[i].String()
					}
					i++
				}
			}

			// 获取时间戳
			block, _ := client.BlockByNumber(context.Background(), big.NewInt(int64(vLog.BlockNumber)))

			// 保存
			event := new(dao.Event)
			event.EventParseId = int(plan.ID)
			event.NetworkId = int(networkId.Int64())
			event.Address = plan.ContractAddress
			event.Method = args.eventName
			event.TxHash = vLog.TxHash.String()
			event.TxTime = time.Unix(int64(block.Time()), 0)
			event.BlockNumber = int(vLog.BlockNumber)
			for i, topic := range vLog.Topics {
				switch i {
				case 0:
					event.Topic0 = topic.String()
				case 1:
					event.Topic1 = topic.String()
				case 2:
					event.Topic2 = topic.String()
				case 3:
					event.Topic3 = topic.String()
				case 4:
					event.Topic4 = topic.String()
				}
			}
			vLogJson, err := vLog.MarshalJSON()
			if err != nil {
				log.Println(err)
			}
			event.OriginalJson = string(vLogJson)
			resultJson, err := json.Marshal(resultMap)
			if err != nil {
				log.Println(err)
			}
			event.ParseJson = string(resultJson)
			dao.SaveEvent(event)
		}
	}
}

func parseEvent(strList []string) map[common.Hash]*EventParseInfo {
	result := make(map[common.Hash]*EventParseInfo, 0)
	for _, str := range strList {
		// Transfer(address indexed from, address indexed to, uint256 value)
		// 匹配事件名称
		eventName := str[0:strings.Index(str, "(")]

		// 匹配参数部分
		regex := regexp.MustCompile(`\((.+)\)`)
		matches := regex.FindStringSubmatch(str)
		// 逗号分隔参数
		args := strings.Split(matches[1], ",")
		var list []abi.ArgumentMarshaling
		// 遍历参数
		for _, arg := range args {
			arg = strings.TrimSpace(arg)

			// 解析名称、类型和是否为索引
			parts := strings.Fields(arg)
			argType := parts[0]
			argName := parts[1]
			isIndexed := false

			if strings.HasPrefix(argName, "indexed") {
				isIndexed = true
				argName = parts[2]
			}

			list = append(list, abi.ArgumentMarshaling{
				Name:    argName,
				Type:    argType,
				Indexed: isIndexed,
			})
		}
		// 去除无用信息
		tempStr := str
		for _, arg := range list {
			tempStr = strings.ReplaceAll(tempStr, arg.Name, "")
			tempStr = strings.ReplaceAll(tempStr, "indexed", "")
		}
		tempStr = strings.ReplaceAll(tempStr, " ", "")
		eventHash := crypto.Keccak256Hash([]byte(tempStr))

		result[eventHash] = &EventParseInfo{
			eventName:   eventName,
			eventHash:   eventHash,
			eventParams: list,
		}
	}

	return result
}

func getHistoryEvent(client *ethclient.Client, plan *dao.EventParsePlan,
	eventParseInfo map[common.Hash]*EventParseInfo, latestBlockNumber *big.Int, networkId *big.Int) {

	eventHashList := make([]common.Hash, 0)
	for hash := range eventParseInfo {

		eventHashList = append(eventHashList, hash)
	}
	// 过滤条件
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(plan.DeployedBlockNumber)),
		ToBlock:   latestBlockNumber,
		Addresses: []common.Address{common.HexToAddress(plan.ContractAddress)},
		Topics:    [][]common.Hash{eventHashList},
	}
	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Printf("get logs fail, err: %v", err)
	}

	for _, vLog := range logs {
		args := eventParseInfo[vLog.Topics[0]]
		noIndexedParams := make([]abi.ArgumentMarshaling, 0)
		for _, arg := range args.eventParams {
			if !arg.Indexed {
				noIndexedParams = append(noIndexedParams, arg)
			}
		}
		argTypes, _ := abi.NewType("tuple", "", noIndexedParams)
		// 根据类型和日志数据解析事件参数
		eventArgs := abi.Arguments{abi.Argument{Type: argTypes}}
		resultMap := make(map[string]interface{})
		err := eventArgs.UnpackIntoMap(resultMap, vLog.Data)
		if err != nil {
			log.Println(err)
		}
		i := 0
		for _, arg := range args.eventParams {
			if arg.Indexed {
				if arg.Type == "address" {
					resultMap[arg.Name] = common.BytesToAddress(vLog.Topics[i].Bytes())
				} else {
					resultMap[arg.Name] = vLog.Topics[i].String()
				}
				i++
			}
		}

		// 获取时间戳
		block, _ := client.BlockByNumber(context.Background(), big.NewInt(int64(vLog.BlockNumber)))

		// 保存
		event := new(dao.Event)
		event.EventParseId = int(plan.ID)
		event.NetworkId = int(networkId.Int64())
		event.Address = plan.ContractAddress
		event.Method = args.eventName
		event.TxHash = vLog.TxHash.String()
		event.TxTime = time.Unix(int64(block.Time()), 0)
		event.BlockNumber = int(vLog.BlockNumber)
		for i, topic := range vLog.Topics {
			switch i {
			case 0:
				event.Topic0 = topic.String()
			case 1:
				event.Topic1 = topic.String()
			case 2:
				event.Topic2 = topic.String()
			case 3:
				event.Topic3 = topic.String()
			case 4:
				event.Topic4 = topic.String()
			}
		}
		vLogJson, err := vLog.MarshalJSON()
		if err != nil {
			log.Println(err)
		}
		event.OriginalJson = string(vLogJson)
		resultJson, err := json.Marshal(resultMap)
		if err != nil {
			log.Println(err)
		}
		event.ParseJson = string(resultJson)
		dao.SaveEvent(event)
	}
}
