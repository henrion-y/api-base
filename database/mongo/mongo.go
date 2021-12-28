package mongo

import (
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
)

type DataBase struct {
	Host     string
	User     string
	Password string
	Db       string
	Charset  string
}

func NewDbProvider(config *DataBase) (*mongo.Database, error) {
	var applyURI string
	if config.User != "" {
		applyURI = fmt.Sprintf("mongodb://%s:%s@%s", config.User, config.Password, config.Host)
	} else {
		applyURI = fmt.Sprintf("mongodb://%s", config.Host)
	}

	opts := options.Client().ApplyURI(applyURI)

	// 连接数据库
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatal(err)
	}

	// 判断服务是不是可用
	if err = client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}

	// 获取数据库和集合
	db := client.Database(config.Db)
	return db, nil
}
