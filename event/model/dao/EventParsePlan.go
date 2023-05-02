package dao

import (
	"go-ethereum-demo/event/global"
	"gorm.io/gorm"
)

type EventParsePlan struct {
	gorm.Model
	NetworksUrl         string   `gorm:"type:varchar(200);not null;comment:网络url" json:"networksUrl" binding:"required"`
	ContractAddress     string   `gorm:"type:varchar(100);not null;comment:合约地址" json:"contractAddress" binding:"required"`
	DeployedBlockNumber int      `gorm:"type:bigint;comment:合约部署区块号" json:"deployedBlockNumber" binding:"required"`
	EventString         []string `gorm:"type:varchar(500);not null;comment:事件声明" json:"eventString" binding:"required"`
}

func (ep EventParsePlan) TableName() string {
	return "event_parse_plan"
}

func CreatePlan(plan *EventParsePlan) {
	global.DB.Create(plan)
}
