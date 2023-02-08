package main

import (
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	_db     *gorm.DB
	_dbRead []*gorm.DB
	UserID  uint
)

type User struct {
	gorm.Model
	UserName    string
	CommentTerm uint64
}

func main() {
	if err := initConfig("./config.ini"); err != nil {
		return
	}

	initDatabase()
	initReadDb()

	var rc1Cmd int
	var rc2Cmd int
	var userCmd int
	flag.IntVar(&rc1Cmd, "c1", 0, "rc check times from master")
	flag.IntVar(&rc2Cmd, "c2", 0, "rc check times from second")
	flag.IntVar(&userCmd, "u", 0, "create how much user?")
	flag.Parse()

	log.Infof("command: rc1Cmd:%v, rc2Cmd:%v, userCmd:%v", rc1Cmd, rc2Cmd, userCmd)
	if rc1Cmd != 0 { // 读写分离，随机从从数据库读取
		readWriteCheck(rc1Cmd)
		log.Infof("rc1 check done")
	} else if rc2Cmd != 0 { // 单调读
		monotonicReadCheck(rc2Cmd)
		log.Infof("rc2 check done")
	} else if userCmd != 0 { // 创建用户
		createUser(userCmd)
	}
}

func readWriteCheck(num int) {
	UserID = 1
	readTerm := uint64(0)
	stop := false

	go func() {
		for {
			if stop {
				return
			}
			// read
			user := User{}
			if err := _db.Where("id = ?", UserID).
				Find(&user).Error; err != nil {
				log.Errorf("readWriteCheck, id:%v, err:%v", UserID, err)
				return
			} else {
				if user.CommentTerm < readTerm {
					log.Errorf("readWriteCheck, get CommentTerm rollback! %v -> %v", readTerm, user.CommentTerm)
				} else {
					readTerm = user.CommentTerm
				}
			}
		}
	}()

	for i := 0; i < num; i++ {
		// update
		if err := _db.Model(&User{}).
			Where("id = ?", UserID).
			Update("comment_term", i).Error; err != nil {
			log.Errorf("readWriteCheck, id:%v, comment_term:%v, err:%v", UserID, i, err)
			stop = true
			return
		}
	}
	stop = true
}

func monotonicReadCheck(num int) {
	UserID = 2
	readTerm := uint64(0)
	dbId := getUserNode(UserID)
	stop := false

	go func() {
		for {
			if stop {
				return
			}
			// read
			user := User{}
			if err := _dbRead[dbId].Where("id = ?", UserID).
				Find(&user).Error; err != nil {
				log.Errorf("monotonicReadCheck, id:%v, err:%v", UserID, err)
				return
			} else {
				if user.CommentTerm < readTerm {
					log.Errorf("monotonicReadCheck, get CommentTerm rollback! %v -> %v", readTerm, user.CommentTerm)
				} else {
					readTerm = user.CommentTerm
				}
			}
		}
	}()

	for i := 0; i < num; i++ {
		// update
		if err := _db.Model(&User{}).
			Where("id = ?", UserID).
			Update("comment_term", i).Error; err != nil {
			log.Errorf("monotonicReadCheck, id:%v, comment_term:%v, err:%v", UserID, i, err)
			stop = true
			return
		}
	}
	stop = true
}
func getUserNode(id uint) uint {
	length := len(_dbRead)
	i := id % uint(length)
	return i
}

func createUser(nums int) {
	for i := 0; i < nums; i++ {
		user := User{
			UserName: fmt.Sprintf("%d", i),
		}
		if err := _db.Create(&user).Error; err != nil {
			log.Errorf("writeUser err:%v", err)
		}
	}
}
