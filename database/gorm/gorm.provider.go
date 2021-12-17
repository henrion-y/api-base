package gorm

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"time"
)

type DataBase struct {
	Driver   string
	Host     string
	User     string
	Password string
	Db       string
	Charset  string
}

func NewDbProvider(config DataBase) (*gorm.DB, error) {

	if len(config.Driver) == 0 {
		return nil, errors.New("driver is empty")
	}

	dial := "%s:%s@(%s)/%s?charset=%s&parseTime=True&loc=Local"
	dial = fmt.Sprintf(dial,
		config.User,
		config.Password,
		config.Host,
		config.Db,
		config.Charset)

	db, err := gorm.Open(config.Driver, dial)
	if err != nil {
		return nil, err
	}

	// defer db.Close()
	db.LogMode(true)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetConnMaxLifetime(3 * time.Minute)

	return db, nil
}
