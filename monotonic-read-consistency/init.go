package main

import (
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func initConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("read config fail, err:%v", err)
		return err
	}
	return nil
}

func initReadDb() {
	RDbHost1 := viper.GetString("MYSQLRead1.HostName")
	RDbPort1 := viper.GetString("MYSQLRead1.Port")
	RDbUser1 := viper.GetString("MYSQLRead1.UserName")
	RDbPassWord1 := viper.GetString("MYSQLRead1.Pwd")
	RDbName1 := viper.GetString("MYSQLRead1.DatabaseName")

	RDbHost2 := viper.GetString("MYSQLRead2.HostName")
	RDbPort2 := viper.GetString("MYSQLRead2.Port")
	RDbUser2 := viper.GetString("MYSQLRead2.UserName")
	RDbPassWord2 := viper.GetString("MYSQLRead2.Pwd")
	RDbName2 := viper.GetString("MYSQLRead2.DatabaseName")

	pathRead1 := strings.Join([]string{RDbUser1, ":", RDbPassWord1, "@tcp(", RDbHost1, ":", RDbPort1, ")/", RDbName1, "?charset=utf8&parseTime=true"}, "")
	pathRead2 := strings.Join([]string{RDbUser2, ":", RDbPassWord2, "@tcp(", RDbHost2, ":", RDbPort2, ")/", RDbName2, "?charset=utf8&parseTime=true"}, "")

	db, err := gorm.Open(mysql.Open(pathRead1), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Second * 30)
	_dbRead = append(_dbRead, db)

	db, err = gorm.Open(mysql.Open(pathRead2), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, _ = db.DB()
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Second * 30)
	_dbRead = append(_dbRead, db)
}

func initDatabase() {
	PDbHost := viper.GetString("MYSQL.HostName")
	PDbPort := viper.GetString("MYSQL.Port")
	PDbUser := viper.GetString("MYSQL.UserName")
	PDbPassWord := viper.GetString("MYSQL.Pwd")
	PDbName := viper.GetString("MYSQL.DatabaseName")

	SDbHost := viper.GetString("MYSQLRead.HostName")
	SDbPort := viper.GetString("MYSQLRead.Port")
	SDbUser := viper.GetString("MYSQLRead.UserName")
	SDbPassWord := viper.GetString("MYSQLRead.Pwd")
	SDbName := viper.GetString("MYSQLRead.DatabaseName")

	pathWrite := strings.Join([]string{PDbUser, ":", PDbPassWord, "@tcp(", PDbHost, ":", PDbPort, ")/", PDbName, "?charset=utf8&parseTime=true"}, "")
	pathRead := strings.Join([]string{SDbUser, ":", SDbPassWord, "@tcp(", SDbHost, ":", SDbPort, ")/", SDbName, "?charset=utf8&parseTime=true"}, "")

	db, err := gorm.Open(mysql.Open(pathWrite), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Second * 30)
	_db = db
	_ = _db.Use(dbresolver.
		Register(dbresolver.Config{
			Sources:  []gorm.Dialector{mysql.Open(pathWrite)}, // 写操作
			Replicas: []gorm.Dialector{mysql.Open(pathRead)},  // 读操作,headless自动选择
			Policy:   dbresolver.RandomPolicy{},               // sources/replicas 负载均衡策略
		}))
	Migration()
}

func Migration() {
	err := _db.Set("gorm:table_options", "charset=utf8mb4").
		AutoMigrate(&User{})
	if err != nil {
		log.Errorf("Migration table fail")
		os.Exit(0)
	}
	log.Infof("Migration table success")
}
