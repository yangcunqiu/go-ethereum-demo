package dao

import (
	"go-ethereum-demo/event/global"
	"gorm.io/gorm"
	"time"
)

type Event struct {
	gorm.Model
	EventParseId int       `gorm:"type:bigint;not null;comment:事件解析原始信息, event_parse_plan表id"`
	NetworkId    int       `gorm:"type:bigint;not null;comment:链id"`
	Address      string    `gorm:"type:varchar(100);not null;comment:合约地址"`
	Method       string    `gorm:"type:varchar(200);not null;comment:事件名"`
	TxHash       string    `gorm:"type:varchar(100);not null;comment:交易hash"`
	TxTime       time.Time `gorm:"type:timestamp;comment:交易时间戳"`
	BlockNumber  int       `gorm:"type:bigint;comment:交易所在区块号"`
	Topic0       string    `gorm:"type:varchar(100);not null;comment:事件topic"`
	Topic1       string    `gorm:"type:varchar(100)"`
	Topic2       string    `gorm:"type:varchar(100)"`
	Topic3       string    `gorm:"type:varchar(100)"`
	Topic4       string    `gorm:"type:varchar(100)"`
	OriginalJson string    `gorm:"type:text;comment:事件原始数据"`
	ParseJson    string    `gorm:"type:text;comment:事件解析数据"`
}

func (event *Event) TableName() string {
	return "event"
}

func SaveEvent(event *Event) {
	global.DB.Create(event)
}
