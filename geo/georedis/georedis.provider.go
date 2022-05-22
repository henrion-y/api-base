package georedis

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/henrion-y/api-base/geo"
	"github.com/spf13/viper"
)

func NewRedisProvider(config *viper.Viper) (geo.Geo, error) {
	host := config.GetString("redis.Host")
	password := config.GetString("redis.Password")
	maxIdle := config.GetInt("redis.MaxIdle")
	maxActive := config.GetInt("redis.MaxActive")
	idleTimeout := config.GetInt("redis.IdleTimeout")
	// cacheTimeOut := config.GetInt("redis,CacheTimeOut")
	if len(host) == 0 {
		return nil, errors.New("host  is empty")
	}

	redisConn := &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: time.Duration(idleTimeout),
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return &Geo{pool: redisConn}, nil
}
