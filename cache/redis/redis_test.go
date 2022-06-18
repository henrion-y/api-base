package redis

import (
	"testing"

	"github.com/henrion-y/api-base/infra/redisapi"

	"github.com/henrion-y/api-base/cache"
	"github.com/spf13/viper"
)

func getCache() cache.Cache {
	conf := viper.New()
	redisApi, _ := redisapi.NewRedisApiProvider(conf)
	cacheRdb, err := NewRedisProvider(redisApi)
	if err != nil {
		panic(err)
	}
	return cacheRdb
}

func TestCache_Set(t *testing.T) {
	rdb := getCache()
	err := rdb.Set("", "", 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCache_Get(t *testing.T) {
	rdb := getCache()
	data := 0
	err := rdb.Get("", &data)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(data)
}
