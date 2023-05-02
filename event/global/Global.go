package global

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"go-ethereum-demo/event/model/config"
	"gorm.io/gorm"
)

var (
	DB        *gorm.DB
	Config    config.Config
	ClientMap map[string]*ethclient.Client
)
