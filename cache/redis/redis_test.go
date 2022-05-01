package redis

import (
	"github.com/henrion-y/api-base/cache"
	"github.com/spf13/viper"
	"testing"
)

func getCache() cache.Cache {
	conf := viper.New()

	cacheRdb, err := NewRedisProvider(conf)
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
