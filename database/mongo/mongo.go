package mongo

import (
	"context"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
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

func NewDbProvider(config *viper.Viper) (*mongo.Database, error) {

	user := config.GetString("mongo.User")
	password := config.GetString("mongo.Password")

	host := config.GetString("mongo.Host")
	if len(host) == 0 {
		return nil, errors.New("host is empty")
	}

	db := config.GetString("mongo.DB")
	if len(db) == 0 {
		return nil, errors.New("db is empty")
	}

	var applyURI string
	if user != "" {
		applyURI = fmt.Sprintf("mongodb://%s:%s@%s", user, password, host)
	} else {
		applyURI = fmt.Sprintf("mongodb://%s", host)
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
	mongoDb := client.Database(db)
	return mongoDb, nil
}
