package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go-ethereum-demo/event/global"
	"go-ethereum-demo/event/model/dao"
	"go-ethereum-demo/event/router"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"strconv"
)

func main() {
	global.ClientMap = make(map[string]*ethclient.Client)
	initConfig()
	initDB()
	initRouter()
}

func initRouter() {
	r := gin.Default()
	router.RegisterRouter(r)
	var addr string
	if global.Config.Server.Port == 0 {
		addr = ":8080"
	} else {
		addr = fmt.Sprintf(":%s", strconv.Itoa(global.Config.Server.Port))
	}
	err := r.Run(addr)
	if err != nil {
		panic(err)
	}
}

func initConfig() {
	// 设置配置文件名称 类型
	viper.SetConfigName("application")
	viper.SetConfigType("yml")
	// 设置viper的查找路径(viper会去这个路径查找上面配置的文件名的文件)
	viper.AddConfigPath("./event")
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	// 把配置信息反序列化到结构体, 方便使用
	err = viper.Unmarshal(&global.Config)
	if err != nil {
		log.Println(err)
	}
	// 运行时监控配置文件的更新
	viper.WatchConfig()
	// 配置文件更新时的回调函数
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("config changed:", e.Name)
		// 重新反序列化
		err = viper.Unmarshal(&global.Config)
		if err != nil {
			log.Println(err)
		}
	})
}

func initDB() {
	db, err := gorm.Open(mysql.Open(getDsn()),
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	if err != nil {
		panic(err)
	}
	global.DB = db
	syncTable()
}

func getDsn() string {
	// "test:pwd1234@tcp(172.31.54.123)/gorm_demo?charset=utf8mb4&parseTime=True&loc=Local"
	if global.Config.Mysql.Dsn != "" {
		return global.Config.Mysql.Dsn
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		global.Config.Mysql.Username,
		global.Config.Mysql.Password,
		global.Config.Mysql.Ip,
		strconv.Itoa(global.Config.Mysql.Port),
		global.Config.Mysql.Database,
	)
}

func syncTable() {
	err := global.DB.AutoMigrate(
		&dao.Event{},
		&dao.EventParsePlan{},
		&dao.NetworkInfo{},
	)

	if err != nil {
		panic(err)
	}
}
