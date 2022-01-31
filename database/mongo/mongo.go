package mongo

import (
	"context"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
)

func NewDbProvider(config *viper.Viper) (*mongo.Database, error) {

	host := config.GetStringSlice("mongo.Hosts")
	if len(host) == 0 {
		return nil, errors.New("hosts is empty")
	}

	db := config.GetString("mongo.DB")
	if len(db) == 0 {
		return nil, errors.New("db is empty")
	}

	opt := options.Client()
	opt.Hosts = host

	user := config.GetString("mongo.User")
	if user != "" {
		password := config.GetString("mongo.Password")
		authSource := config.GetString("mongo.AuthSource")
		opt.SetAuth(options.Credential{
			AuthSource: authSource,
			Username:   user,
			Password:   password,
		})
	}

	// 连接数据库
	client, err := mongo.Connect(context.Background(), opt)
	if err != nil {
		return nil, err
	}

	// 判断服务是不是可用
	if err = client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}

	// 获取数据库和集合
	mongoDb := client.Database(db)
	return mongoDb, nil
}
