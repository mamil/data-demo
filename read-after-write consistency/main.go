package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"gorm.io/gorm"
)

var (
	_db       *gorm.DB
	_dbMaster *gorm.DB
)

// User 用户模型
type User struct {
	gorm.Model
	UserName string
	Email    string
}

// 记录需要指向master的用户id
type PinningUser struct {
	gorm.Model
	UserId uint
}

func main() {
	if err := initConfig("./config.ini"); err != nil {
		return
	}

	initDatabase()
	initMasterDb()

	var rc1Cmd int
	var rc2Cmd int
	var userCmd int
	flag.IntVar(&rc1Cmd, "c1", 0, "rc check times from master")
	flag.IntVar(&rc2Cmd, "c2", 0, "rc check times from second")
	flag.IntVar(&userCmd, "u", 0, "create how much user?")
	flag.Parse()

	log.Infof("command: rc1Cmd:%v, rc2Cmd:%v, userCmd:%v", rc1Cmd, rc2Cmd, userCmd)
	if rc1Cmd != 0 { // 有写操作的从主数据库读取
		rcCheck(1, rc1Cmd)
		log.Infof("rc1 check done")
	} else if rc2Cmd != 0 { // 全部从从数据库读取
		rcCheck(2, rc2Cmd)
		log.Infof("rc2 check done")
	} else if userCmd != 0 { // 创建用户
		createUser(userCmd)
	}
}

func rcCheck(rc int, times int) {
	lastUser := readLast()
	lastId := uint32(lastUser.ID)
	rand.Seed(time.Now().Unix())
	for i := 0; i < times; i++ {
		modId := rand.Uint32() % lastId // userid从1开始
		for {
			if modId == 0 {
				modId = rand.Uint32() % lastId
			} else {
				break
			}
		}
		modStr := uuid.NewV1().String()
		// 修改某人的信息
		modUser(uint(modId), modStr)

		// 马上读取，检查信息是否正常
		if rc == 1 {
			readStr := readUserRc(uint(modId))
			if readStr != modStr {
				log.Errorf("rcCheck1, fail, modId:%v, modStr:%v, readStr:%v",
					modId, modStr, readStr)
			}
		} else if rc == 2 {
			readStr := readUser(uint(modId))
			if readStr != modStr {
				log.Errorf("rcCheck2, fail, modId:%v, modStr:%v, readStr:%v",
					modId, modStr, readStr)
			}
		}
	}
}

// 这里修改用户信息，并且记录这个用户
func modUser(userId uint, email string) error {
	if err := _db.Model(&User{}).
		Where("id = ?", userId).
		Update("email", email).Error; err != nil {
		log.Errorf("modUser, id:%v, email:%v, err:%v", userId, email, err)
		return err
	}
	pinUser := PinningUser{UserId: userId}
	if err := _db.Create(&pinUser).Error; err != nil {
		log.Errorf("modUser, PinningUser id:%v, err:%v", userId, err)
		return err
	}
	return nil
}

// 全部从节点读取，测试能否复现这个问题
func readUser(userId uint) string {
	user := User{}
	if err := _db.Where("id = ?", userId).
		Find(&user).Error; err != nil {
		log.Errorf("readUser, id:%v, err:%v", userId, err)
		return ""
	} else {
		return user.Email
	}

}

// 如果这个用户数据被修改了，就从主节点读取
func readUserRc(userId uint) string {
	user := User{}
	pinUser := int64(0)
	if err := _dbMaster.Model(&PinningUser{}).
		Where("user_id = ?", userId).
		Count(&pinUser).Error; err != nil {
		log.Errorf("readUser, id:%v, err:%v", userId, err)
		return ""
	}

	if pinUser == 0 {
		if err := _db.Where("id = ?", userId).
			Find(&user).Error; err != nil {
			log.Errorf("readUser, from second id:%v, err:%v", userId, err)
			return ""
		} else {
			return user.Email
		}
	} else {
		if err := _dbMaster.Where("id = ?", userId).
			Find(&user).Error; err != nil {
			log.Errorf("readUser, from master id:%v, err:%v", userId, err)
			return ""
		} else {
			return user.Email
		}
	}
}

func readLast() *User {
	user := User{}
	if err := _db.Model(&User{}).Last(&user).Error; err != nil {
		log.Errorf("readLast err:%v", err)
	}
	return &user
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
