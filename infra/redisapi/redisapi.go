package redisapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
)

const (
	DFT_MAX_REDIS_POOL_LIMIT = 1000

	// A ping is set to the server with this period to test for the health of
	// the connection and server.
	healthCheckPeriod = time.Second * 30
)

var (
	ErrRedisExecFailed = errors.New("redis exec failed")
	ErrRedisNotExisted = errors.New("redigo: nil returned")
	ErrRedisKeyNotSet  = errors.New("redis key not set")
	// for https://redis.io/commands/expire and https://redis.io/commands/expireat
	ErrRedisKeyNotExist = errors.New("redis key does not exist")

	ErrNil = redis.ErrNil
)

type HandleListLoopMessageFunc func(message *string) (code uint32, err error)

// RedisApi Conn exposes a set of callbacks for the various events that occur on a connection
type RedisApi struct {
	RedisPool   *redis.Pool
	redisServer string
}

// NewRedisApiProvider create new *RedisApi with maxPoolSize pool size, AUTH is enabled if passwd is not empty string
func NewRedisApiProvider(config *viper.Viper) (*RedisApi, error) {
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

	return &RedisApi{
		RedisPool:   redisConn,
		redisServer: "",
	}, nil
}

func (api *RedisApi) Get(key string) (value []byte, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Bytes(redisConn.Do("GET", key))

	return value, err
}

func (api *RedisApi) GetBit(key string, offset int) (value int, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int(redisConn.Do("GETBIT", key, offset))

	return value, err
}

func (api *RedisApi) GetSet(key, content string, expire int) (value string, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.String(redisConn.Do("GETSET", key, content))

	_ = api.Expire(key, uint64(expire))

	return value, err
}

func (api *RedisApi) Mget(field []string) (value []string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var args []interface{}
	for _, v := range field {
		args = append(args, v)
	}
	r, err := redisConn.Do("MGET", args...)
	value, err = redis.Strings(r, err)

	return value, err
}

// MGetSlice 批量得到slice
func (api *RedisApi) MGetSlice(field []string) (value [][]byte, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var args []interface{}
	for _, v := range field {
		args = append(args, v)
	}
	value, err = redis.ByteSlices(redisConn.Do("MGET", args...))

	return value, err
}

func (api *RedisApi) MSet(keyVal []interface{}) (err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	retValue, err := redis.String(redisConn.Do("MSET", keyVal...))
	if err == nil && retValue == "OK" {
	} else if err == nil {
		err = ErrRedisExecFailed
	}

	return err
}

func (api *RedisApi) GetInt64(key string) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int64(redisConn.Do("GET", key))

	return value, err
}

func (api *RedisApi) GetSetInt64(key string, content int64, expire int) (value int64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int64(redisConn.Do("GETSET", key, content))
	_ = api.Expire(key, uint64(expire))

	return value, err
}

func (api *RedisApi) Incr(key string) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int64(redisConn.Do("INCR", key))

	return value, err
}

func (api *RedisApi) IncrBy(key string, increment int64) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int64(redisConn.Do("INCRBY", key, increment))

	return value, err
}

func (api *RedisApi) Decr(key string) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int64(redisConn.Do("DECR", key))

	return value, err
}

func (api *RedisApi) DecrBy(key string, increment int64) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int64(redisConn.Do("DECRBY", key, increment))

	return value, err
}

func (api *RedisApi) SetNxEx(key, value string, expireTs int64) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	retValue, err := redis.String(redisConn.Do("SET", key, value, "NX", "EX", expireTs))
	defer func() {
		_ = redisConn.Close()
	}()

	if err != nil {
		return err
	}
	if retValue == "OK" { // 执行成功
		return nil
	}

	return ErrRedisExecFailed
}

func (api *RedisApi) Setex(key, value string, expireTs uint32) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	retValue, err := redis.String(redisConn.Do("SETEX", key, expireTs, value))
	if err == nil && retValue == "OK" {
	} else if err == nil {
		err = ErrRedisExecFailed
	}

	return err
}

// SetNMxInt64 return nil if no err and the key is not exists and set
// return ErrRedisKeyNotSet if no err and the key is not set
// return other err if command failed
func (api *RedisApi) SetNMxInt64(key string, value int64) (err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	retValue, err := redis.Int(redisConn.Do("SETNX", key, value))
	if err == nil && retValue == 1 {
		// not exists and set ok
	} else if err == nil && retValue == 0 {
		err = ErrRedisKeyNotSet
	}

	return err
}

func (api *RedisApi) Set(key, value interface{}) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	retValue, err := redis.String(redisConn.Do("SET", key, data))
	if err == nil && retValue == "OK" {
	} else if err == nil {
		err = ErrRedisExecFailed
	}

	return err
}

func (api *RedisApi) SetBit(key string, offset, value int) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	_, err = redis.Int(redisConn.Do("SETBIT", key, offset, value))

	return err
}

func (api *RedisApi) Del(key string) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	delCount, err := redis.Int(redisConn.Do("DEL", key))
	if err == nil && delCount == 1 {
		err = nil
	} else if err == nil && delCount == 0 {
		err = nil
	}

	return err
}

func (api *RedisApi) BatchDel(keys []string) (count int, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var args []interface{}
	for _, v := range keys {
		args = append(args, v)
	}
	if err != nil {
		return count, err
	}
	delCount, err := redis.Int(redisConn.Do("DEL", args...))

	return delCount, err
}

func (api *RedisApi) Exists(key string) (result bool, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	count, err := redis.Int(redisConn.Do("EXISTS", key))
	if err == nil {
		if count == 1 {
			result = true
		} else {
			result = false
		}
	} else {
		result = false
	}

	return result, err
}

// TTL 检查key的过期时间
func (api *RedisApi) TTL(key string) (result int, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	result, err = redis.Int(redisConn.Do("TTL", key))

	return result, err
}

// Expire 设置Key的生存时间
func (api *RedisApi) Expire(key string, seconds uint64) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	retValue, err := redis.Int(redisConn.Do("EXPIRE", key, seconds))
	if err == nil {
		if retValue == 1 {
			err = nil
		} else if retValue == 0 {
			// see https://redis.io/commands/expire
			// Return value
			// Integer reply, specifically:
			//    1 if the timeout was set.
			//    0 if key does not exist.
			err = ErrRedisKeyNotExist
		} else {
			// this should not happen
			err = ErrRedisExecFailed
		}
	}

	return err
}

// Expireat 设置Key的生存时间 使用UnixTimeStamp 设置
func (api *RedisApi) Expireat(key string, timeStamp int64) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	retValue, err := redis.Int(redisConn.Do("EXPIREAT", key, timeStamp))
	if err == nil {
		if retValue == 1 {
			err = nil
		} else if retValue == 0 {
			// see https://redis.io/commands/expire
			// Return value
			// Integer reply, specifically:
			//    1 if the timeout was set.
			//    0 if key does not exist.
			err = ErrRedisKeyNotExist
		} else {
			// this should not happen
			err = ErrRedisExecFailed
		}
	}

	return err
}

func (api *RedisApi) Keys(pattern string) (value []string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Strings(redisConn.Do("KEYS", pattern))

	return value, err
}

func (api *RedisApi) Hget(key, field string) (value string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.String(redisConn.Do("HGET", key, field))

	return value, err
}

func (api *RedisApi) HgetBytes(key, field string) (value []byte, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Bytes(redisConn.Do("HGET", key, field))

	return value, err
}

func (api *RedisApi) HgetInt64(key, field string) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int64(redisConn.Do("HGET", key, field))

	return value, err
}

func (api *RedisApi) Hset(key, field, value string) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	retValue, err := redis.Int(redisConn.Do("HSET", key, field, value))
	if err == nil && (retValue == 1 || retValue == 0) {
		err = nil
	} else if err == nil {
		err = ErrRedisExecFailed
	}

	return err
}

func (api *RedisApi) HsetInt64(key, field string, value int64) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	retValue, err := redis.Int(redisConn.Do("HSET", key, field, value))
	if err == nil && (retValue == 1 || retValue == 0) {
		err = nil
	} else if err == nil {
		err = ErrRedisExecFailed
	}

	return err
}

func (api *RedisApi) HclearSSDB(key string) error {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	_, err := redisConn.Do("HCLEAR", key)

	return err
}

func (api *RedisApi) Hdel(key, field string) (result bool, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	delCount, err := redis.Int(redisConn.Do("HDEL", key, field))
	if err == nil && delCount == 1 {
		result = true
	} else if err == nil {
		err = ErrRedisExecFailed
		result = false
	}

	return result, err
}

func (api *RedisApi) HBatchDel(key string, field []string) (delCount int, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var args []interface{}
	args = append(args, key)
	for _, v := range field {
		args = append(args, v)
	}
	delCount, err = redis.Int(redisConn.Do("HDEL", args...))

	return delCount, err
}

func (api *RedisApi) HMGet(key string, field []string) (value []string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var args []interface{}
	args = append(args, key)
	for _, v := range field {
		args = append(args, v)
	}
	value, err = redis.Strings(redisConn.Do("HMGET", args...))

	return value, err
}

// HMGetBytes 得到bytes的数组
func (api *RedisApi) HMGetBytes(key string, field []string) (value [][]byte, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var args []interface{}
	args = append(args, key)
	for _, v := range field {
		args = append(args, v)
	}
	value, err = redis.ByteSlices(redisConn.Do("HMGET", args...))

	return value, err
}

func (api *RedisApi) HMGetInt(key string, field []string) (value []int64, err error) {
	vals, err := api.HMGet(key, field)
	if err != nil {
		return value, err
	}

	for _, v := range vals {
		valInt, _ := strconv.ParseInt(v, 10, 64)
		value = append(value, valInt)
	}

	return value, err
}

func (api *RedisApi) HMGetFloat32(key string, field []string) (value []float64, err error) {
	vals, err := api.HMGet(key, field)
	if err != nil {
		return value, err
	}

	for _, v := range vals {
		valInt, _ := strconv.ParseFloat(v, 32)
		value = append(value, valInt)
	}

	return value, err
}

func (api *RedisApi) HMGetFloat64(key string, field []string) (value []float64, err error) {
	vals, err := api.HMGet(key, field)
	if err != nil {
		return value, err
	}

	for _, v := range vals {
		valInt, _ := strconv.ParseFloat(v, 64)
		value = append(value, valInt)
	}

	return value, err
}

func (api *RedisApi) Hlen(key string) (count int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	count, err = redis.Int64(redisConn.Do("HLEN", key))

	return count, err
}

// HmSet 使用key-value 键值对组成的slice
func (api *RedisApi) HmSet(key string, keyValue []string) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var args []interface{}
	args = append(args, key)
	for _, v := range keyValue {
		args = append(args, v)
	}
	retValue, err := redis.String(redisConn.Do("HMSET", args...))
	if err == nil && retValue == "OK" {
	} else if err == nil {
		err = ErrRedisExecFailed
	}

	return err
}

func (api *RedisApi) HGetAll(key string) (keyValue map[string]string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	keyValue, err = redis.StringMap(redisConn.Do("HGETALL", key))

	return keyValue, err
}

func (api *RedisApi) HIncrBy(key, field string, increment int) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int64(redisConn.Do("HINCRBY", key, field, increment))

	return value, err
}

// HScan 当hash map 较小时返回整个hashmap,当hashmap 较大时，返回前面的10条
func (api *RedisApi) HScan(key string, cursor int64) (items []string, outCursor int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var values []interface{}
	values, err = redis.Values(redisConn.Do("HSCAN", key, cursor))
	if err != nil {
		return
	}
	_, err = redis.Scan(values, &outCursor, &items)

	return items, outCursor, err
}

// HscanSSDB 在 SSDB 中 HSCAN 用法和 Redis 不一样，详情可以看其文档：
// http://ssdb.io/docs/commands/hscan.html
// 需要注意 (key_start, key_end]

func (api *RedisApi) HScanSSDB(key, start, end string, limit int64) (items []string, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err := redis.Strings(redisConn.Do("HSCAN", key, start, end, limit))

	return value, err
}

// HrScanSSDB 在 SSDB 中 HSCAN 用法和 Redis 不一样，详情可以看其文档：
// http://ssdb.io/docs/commands/hscan.html
// 需要注意 (key_start, key_end]
func (api *RedisApi) HrScanSSDB(key, start, end string, limit int64) (items []string, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err := redis.Strings(redisConn.Do("HRSCAN", key, start, end, limit))

	return value, err
}

func (api *RedisApi) HExists(key, field string) (ret int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	ret, err = redis.Int64(redisConn.Do("HEXISTS", key, field))

	return ret, err
}

// 兼容SSDB的keys遍历
func (api *RedisApi) HScanWithRange(key string, startSubKey, endSubKey string, maxCounts int64) (value [][]byte, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.ByteSlices(redisConn.Do("hscan", key, startSubKey, endSubKey, maxCounts))

	return value, err
}

// HKeys
func (api *RedisApi) HKeys(key string) (value []string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Strings(redisConn.Do("Hkeys", key))

	return value, err
}

// HKeysSSDB 在 SSDB 中 HKEYS 用法和 Redis 不一样，详情可以看其文档：
// http://ssdb.io/docs/commands/hkeys.html
// 需要注意 (key_start, key_end]
// redisapi 不知为何无法用 ssdb hkeys 命令指定区间，会报 ERR wrong number of arguments 暂时用 scan 实现
func (api *RedisApi) HKeysSSDB(key, start, end string, limit int64) ([]string, error) {
	items, err := api.HScanSSDB(key, start, end, limit)
	if err != nil {
		return nil, err
	}
	var keys []string
	for i, v := range items {
		if i%2 == 0 {
			keys = append(keys, v)
		}
	}
	return keys, nil
}

// HKeysRange 兼容SSDB的keys遍历
func (api *RedisApi) HKeysRange(key string, startSubKey, endSubKey string, maxCounts int64) (value map[string]string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.StringMap(redisConn.Do("hscan", key, startSubKey, endSubKey, maxCounts))

	return value, err
}

// ZAdd 返回 被成功添加的新成员的数量，不包括那些被更新的、已经存在的成员
func (api *RedisApi) ZAdd(key string, score int64, member string) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int64(redisConn.Do("ZADD", key, score, member))

	return value, err
}

// new added: 20200808 reason: the score should allow float64 value
func (api *RedisApi) ZAddFloat64(key string, score float64, member string) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int64(redisConn.Do("ZADD", key, score, member))

	return value, err
}

func (api *RedisApi) ZAddSlice(key string, keyScore []uint32) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var args []interface{}
	args = append(args, key)
	keyScoreLne := len(keyScore)
	for i := 0; i < keyScoreLne; i += 2 { // key-score pari
		member := fmt.Sprintf("%v", keyScore[i])
		score := keyScore[i+1]
		args = append(args, score)
		args = append(args, member)
	}
	value, err = redis.Int64(redisConn.Do("ZADD", args...))

	return value, err
}

// ZAddStringSlice Deprecated 这个方法设计得有点反 redis, 是按 member, score 反过来的
func (api *RedisApi) ZAddStringSlice(key string, keyScore []string) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var args []interface{}
	args = append(args, key)
	keyScoreLne := len(keyScore)
	for i := 0; i < keyScoreLne; i += 2 { // key-score pari
		member := keyScore[i]
		score := keyScore[i+1]
		args = append(args, score)
		args = append(args, member)
	}
	value, err = redis.Int64(redisConn.Do("ZADD", args...))

	return value, err
}

// ZAddBatch key score member [score member ...]
func (api *RedisApi) ZAddBatch(key string, scoreMemberPairs []string) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var args []interface{}
	args = append(args, key)
	keyScoreLne := len(scoreMemberPairs)
	for i := 0; i < keyScoreLne; i += 2 { // key-score pari
		score := scoreMemberPairs[i]
		member := scoreMemberPairs[i+1]
		args = append(args, score)
		args = append(args, member)
	}
	value, err = redis.Int64(redisConn.Do("ZADD", args...))

	return value, err
}

func (api *RedisApi) ZAddInterfaceSlice(key string, keyScore []interface{}) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var args []interface{}
	args = append(args, key)
	keyScoreLne := len(keyScore)
	for i := 0; i < keyScoreLne; i += 2 { // key-score pari
		member := keyScore[i]
		score := keyScore[i+1]
		args = append(args, score)
		args = append(args, member)
	}
	value, err = redis.Int64(redisConn.Do("ZADD", args...))

	return value, err
}

// 返回 被成功移除的成员的数量，不包括被忽略的成员。
func (api *RedisApi) ZRem(key, member string) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int64(redisConn.Do("ZREM", key, member))

	return value, err
}

// Removes the specified members from the sorted set stored at key. Non existing members are ignored.
// https://redis.io/commands/zrem
// An error is returned when key exists and does not hold a sorted set.
func (api *RedisApi) ZRemBatch(key string, member []string) (delCount int, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var args []interface{}
	args = append(args, key)
	for _, v := range member {
		args = append(args, v)
	}
	delCount, err = redis.Int(redisConn.Do("ZREM", args...))

	return delCount, err
}

// 返回 指定区间内，带有 score 值(可选)的有序集成员的列表
func (api *RedisApi) ZRange(key string, start, stop int64) (valSlice []string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	valSlice, err = redis.Strings(redisConn.Do("ZRANGE", key, start, stop, "WITHSCORES"))

	return valSlice, err
}

func (api *RedisApi) ZRangeWithOutScore(key string, start, stop int64) (keys []string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	keys, err = redis.Strings(redisConn.Do("ZRANGE", key, start, stop))

	return keys, err
}

func (api *RedisApi) ZRevrange(key string, start, stop int64) (valSlice []string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	valSlice, err = redis.Strings(redisConn.Do("ZREVRANGE", key, start, stop, "WITHSCORES"))

	return valSlice, err
}

func (api *RedisApi) ZRevrangeWithOutScore(key string, start, stop int64) (keys []string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	keys, err = redis.Strings(redisConn.Do("ZREVRANGE", key, start, stop))

	return keys, err
}

// 返回 member 成员的 score 值，以字符串形式表示
func (api *RedisApi) Zscore(key, member string) (value uint64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Uint64(redisConn.Do("ZSCORE", key, member))

	return value, err
}

// 将有序集合src复制到des
func (api *RedisApi) ZunionStore(des, src string) (value uint64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	r, err := redisConn.Do("ZUNIONSTORE", des, 1, src)
	if err == io.EOF {
		redisConn.Close()
		redisConn = api.RedisPool.Get()
		r, err = redisConn.Do("ZUNIONSTORE", des, 1, src)
	}

	value, err = redis.Uint64(r, err)

	return value, err
}

func (api *RedisApi) Lpush(key, value string) (listSize int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	listSize, err = redis.Int64(redisConn.Do("LPUSH", key, value))

	return listSize, err
}

func (api *RedisApi) Rpush(key, value string) (listSize int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	listSize, err = redis.Int64(redisConn.Do("RPUSH", key, value))

	return listSize, err
}

func (api *RedisApi) Sadd(key, value string) (ret int64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	ret, err = redis.Int64(redisConn.Do("SADD", key, value))

	return ret, err
}

func (api *RedisApi) Srem(key, value string) (ret int64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	ret, err = redis.Int64(redisConn.Do("SREM", key, value))

	return ret, err
}

func (api *RedisApi) Sismember(key, value string) (ret int64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	ret, err = redis.Int64(redisConn.Do("SISMEMBER", key, value))

	return ret, err
}

///blpop 操作
func (api *RedisApi) BLpop(key string, timeout uint8) (string, error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var value []string
	var err error
	value, err = redis.Strings(redisConn.Do("blpop", key, timeout))
	redisConn.Close()
	if value == nil {
		return "", err
	}

	return value[len(value)-1], err
}

func (api *RedisApi) StartLoop(
	key string,
	timeout uint8,
	maxGoroutineLimit uint32,
	callback HandleListLoopMessageFunc,
) {
	for i := uint32(0); i < maxGoroutineLimit; i++ {
		go func() {
			for {
				info, rerr := api.BLpop(key, timeout)
				if rerr != nil {
					continue
				}
				if code, err := callback(&info); err != nil {
					log.Printf("StartLoop() callback exec failed, redis_addr=%v code=%v err=%v", api.redisServer, code, err)
				}
			}
		}()
	}
}

func (api *RedisApi) SPop(key string, params ...interface{}) (ret []string, err error) {
	args := make([]interface{}, len(params)+1)
	args[0] = key
	copy(args[1:], params)
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	ret, err = redis.Strings(redisConn.Do("SPOP", args...))

	return ret, err
}

// 发布
func (api *RedisApi) Publish(key, value string) (ret int64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	ret, err = redis.Int64(redisConn.Do("PUBLISH", key, value))

	return ret, err
}

func (api *RedisApi) Sets(key, value interface{}) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	retValue, err := redis.String(redisConn.Do("SET", key, value))
	if err == nil && retValue == "OK" {
	} else if err == nil {
		err = ErrRedisExecFailed
	}

	return err
}

func (api *RedisApi) Setexs(key, value interface{}, expireTs uint32) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	retValue, err := redis.String(redisConn.Do("SETEX", key, expireTs, value))
	if err == nil && retValue == "OK" {
	} else if err == nil {
		err = ErrRedisExecFailed
	}

	return err
}

func (api *RedisApi) ClusterNodes() (nodes []string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	r, e := redisConn.Do("cluster", "nodes")
	if e != nil {
		return nil, e
	}
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)

	r1, ok := r.([]byte)
	if !ok {
		return nil, nil
	}

	values := string(r1)
	lines := strings.Split(values, "\n")

	masterNodes := make([]string, 0)

	for _, line := range lines {
		if strings.Contains(line, "master") {
			masterNodes = append(masterNodes, strings.Split(line, " ")[0])
		}
	}

	return masterNodes, nil
}

func (api *RedisApi) Scan(cursor uint64, match string, count int64) (keys []string, step int, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	r, e := redisConn.Do("scan", cursor, "match", match, "count", count)
	if e != nil {
		err = e
		return
	}
	r1, ok := r.([]interface{})
	if !ok {
		err = fmt.Errorf("Scans assertion fail r=%#v", r)
		return
	}
	keys, e = redis.Strings(r1[1], e)
	if err != nil {
		err = e
		return
	}
	step, err = redis.Int(r1[0], e)
	return
}

// 获取有序集合的个数
func (api *RedisApi) Zlexcount(key string, params []interface{}) (value uint64, err error) {
	// 获取一条Redis连接
	args := make([]interface{}, len(params)+1)
	args[0] = key
	copy(args[1:], params)
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Uint64(redisConn.Do("ZLEXCOUNT", args...))

	return
}

func (api *RedisApi) GetSlice(key string) (value []byte, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Bytes(redisConn.Do("GET", key))

	return value, err
}

func (api *RedisApi) HmSets(key string, params []interface{}) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	args := make([]interface{}, len(params)+1)
	args[0] = key
	copy(args[1:], params)
	retValue, err := redis.String(redisConn.Do("HMSET", args...))
	if err == nil && retValue == "OK" {
	} else if err == nil {
		err = ErrRedisExecFailed
	}

	return err
}

func (api *RedisApi) BFreserve(key string, errorRate float64, size uint64) (ret string, err error) {
	args := []interface{}{key, errorRate, size}
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	ret, err = redis.String(redisConn.Do("BF.RESERVE", args...))

	return ret, err
}

func (api *RedisApi) BFadd(key string, value interface{}) (ret uint64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	ret, err = redis.Uint64(redisConn.Do("BF.ADD", key, value))

	return ret, err
}

func (api *RedisApi) BFexists(key string, value interface{}) (ret uint64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	ret, err = redis.Uint64(redisConn.Do("BF.EXISTS", key, value))

	return ret, err
}

func (api *RedisApi) BFmadd(key string, params []interface{}) (ret []int64, err error) {
	args := make([]interface{}, len(params)+1)
	args[0] = key
	copy(args[1:], params)
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	ret, err = redis.Int64s(redisConn.Do("BF.MADD", args...))

	return
}

func (api *RedisApi) BFmexists(key string, params []interface{}) (ret []int64, err error) {
	args := make([]interface{}, len(params)+1)
	args[0] = key
	copy(args[1:], params)
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	ret, err = redis.Int64s(redisConn.Do("BF.MEXISTS", args...))

	return
}

func (api *RedisApi) BFInsertNoCreate(key string, values []interface{}) (ret []int64, err error) {
	args := make([]interface{}, 0, len(values)+3)
	args = append(args, key, "NOCREATE", "ITEMS")
	args = append(args, values...)
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	ret, err = redis.Int64s(redisConn.Do("BF.INSERT", args...))

	return ret, err
}

func (api *RedisApi) BFInsert(key string, errRate float64, size int64, values []interface{}) (ret []int64, err error) {
	args := make([]interface{}, 0, len(values)+3)
	args = append(args, key, "CAPACITY", size, "ERROR", errRate, "ITEMS")
	args = append(args, values...)
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	ret, err = redis.Int64s(redisConn.Do("BF.INSERT", args...))

	return ret, err
}

func (api *RedisApi) Sadds(key string, params []interface{}) (value uint64, err error) {
	args := make([]interface{}, len(params)+1)
	args[0] = key
	copy(args[1:], params)
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Uint64(redisConn.Do("SADD", args...))

	return
}

func (api *RedisApi) Smembers(key string) (value []string, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Strings(redisConn.Do("SMEMBERS", key))

	return
}

func (api *RedisApi) SRandMember(key string, num interface{}) (value []string, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Strings(redisConn.Do("SRANDMEMBER", key, num))

	return value, err
}

func (api *RedisApi) SRandMemberInt64(key string, num interface{}) (value []int64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int64s(redisConn.Do("SRANDMEMBER", key, num))

	return value, err
}

func (api *RedisApi) SCard(key string) (value int64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	value, err = redis.Int64(redisConn.Do("SCARD", key))

	return value, err
}

// SRemBatch Remove the specified members from the set stored at key.
// Specified members that are not a member of this set are ignored.
// If key does not exist, it is treated as an empty set and this command returns 0.
// An error is returned when the value stored at key is not a set.
// SREM key member [member ...]
func (api *RedisApi) SRemBatch(key string, members []string) (removedNum uint64, err error) {
	args := make([]interface{}, 0, len(members)+1)
	args = append(args, key)
	for _, v := range members {
		args = append(args, v)
	}
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	removedNum, err = redis.Uint64(redisConn.Do("SREM", args...))

	return
}

func (api *RedisApi) ZAddIntSlice(key string, keyScore []int) (value int64, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var args []interface{}
	args = append(args, key)
	keyScoreLne := len(keyScore)
	for i := 0; i < keyScoreLne; i += 2 { // key-score pari
		member := keyScore[i]
		score := keyScore[i+1]
		args = append(args, score)
		args = append(args, member)
	}
	value, err = redis.Int64(redisConn.Do("ZADD", args...))

	return value, err
}

func (api *RedisApi) ZRevrangeInt(key string, start, stop int64) (valSlice []int, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	valSlice, err = redis.Ints(redisConn.Do("ZREVRANGE", key, start, stop, "WITHSCORES"))

	return valSlice, err
}

func (api *RedisApi) RpushSlice(key string, params []interface{}) (listSize int64, err error) {
	// 获取一条Redis连接
	args := make([]interface{}, len(params)+1)
	args[0] = key
	copy(args[1:], params)
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	listSize, err = redis.Int64(redisConn.Do("RPUSH", args...))

	return listSize, err
}

// Returns all the elements in the sorted set at key with a score between min and max (including elements with score equal to min or max).
// The elements are considered to be ordered from low to high scores.
func (api *RedisApi) ZRangeByScore(key string, min float64, max float64, offset, count int64) (valSlice []string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	// zrangebyscore key min max [WITHSCORES] [LIMIT offset count]
	// member1, score1, member2, score2, member3, score3
	valSlice, err = redis.Strings(redisConn.Do("ZRANGEBYSCORE", key, min, max, "WITHSCORES", "LIMIT", offset, count))

	return valSlice, err
}

func (api *RedisApi) ZRangeByScoreInt64(key string, min int64, max int64, offset, count int64) (valSlice []string, err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	// zrangebyscore key min max [WITHSCORES] [LIMIT offset count]
	// member1, score1, member2, score2, member3, score3
	valSlice, err = redis.Strings(redisConn.Do("ZRANGEBYSCORE", key, min, max, "WITHSCORES", "LIMIT", offset, count))

	return valSlice, err
}

// Deprecated: Use ZRemBatch(key string, member []string) (delCount int, err error) instead.
func (api *RedisApi) ZRemMulti(key string, member []string) (int64, error) {
	delCnt, err := api.ZRemBatch(key, member)
	return int64(delCnt), err
}

// Zcard returns the sorted set cardinality (number of elements) of the sorted set stored at key.
// Integer reply: the cardinality (number of elements) of the sorted set, or 0 if key does not exist.
func (api *RedisApi) Zcard(key string) (count int64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	count, err = redis.Int64(redisConn.Do("ZCARD", key))

	return count, err
}

// Zclear delete all members in a zset
func (api *RedisApi) Zclear(key string) (err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	_, err = redisConn.Do("ZCLEAR", key)

	return err
}

// Zcount returns the number of elements in the sorted set at key with a score between min and max.
func (api *RedisApi) Zcount(key string, min float64, max float64) (count int64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	// ZCOUNT key min max
	count, err = redis.Int64(redisConn.Do("ZCOUNT", key, min, max))

	return count, err
}

// Zcount returns the number of elements in the sorted set at key with a score between min and max.
func (api *RedisApi) ZcountInt64(key string, min int64, max int64) (count int64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	// ZCOUNT key min max
	count, err = redis.Int64(redisConn.Do("ZCOUNT", key, min, max))

	return count, err
}

// ZRemRangeByScore removes all elements in the sorted set stored at key with a score between min and max (inclusive).
// Integer reply: the number of elements removed.
func (api *RedisApi) ZRemRangeByScore(key string, min float64, max float64) (count int64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	//  ZREMRANGEBYSCORE key min max
	count, err = redis.Int64(redisConn.Do("ZREMRANGEBYSCORE", key, min, max))

	return count, err
}

func (api *RedisApi) ZRemRangeByScoreInt64(key string, min int64, max int64) (count int64, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	//  ZREMRANGEBYSCORE key min max
	count, err = redis.Int64(redisConn.Do("ZREMRANGEBYSCORE", key, min, max))

	return count, err
}

// load the script
func (api *RedisApi) LoadScript(keyCount int, src string) (*redis.Script, error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	scr := redis.NewScript(keyCount, src)
	err := scr.Load(redisConn)
	redisConn.Close()
	if err != nil {
		return nil, err
	}
	return scr, nil
}

// DoScript evaluates the script
func (api *RedisApi) DoScript(scr *redis.Script, keysAndArgs ...interface{}) (interface{}, error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	r, err := scr.Do(redisConn, keysAndArgs...)

	return r, err
}

// evaluates the script without waiting for the reply
func (api *RedisApi) DoScriptNoWait(scr *redis.Script, keysAndArgs ...interface{}) error {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	err := scr.SendHash(redisConn, keysAndArgs...)

	return err
}

// Remove the existing timeout on key, turning the key from volatile (a key with an expire set) to
// persistent (a key that will never expire as no timeout is associated).
func (api *RedisApi) Persist(key string) (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	var retValue int
	retValue, err = redis.Int(redisConn.Do("PERSIST", key))
	if err == nil && retValue == 1 {
		// do nothing
	} else if err == nil {
		err = ErrRedisExecFailed
	}

	return err
}

func (api *RedisApi) FlushDB() (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	_, err = redisConn.Do("FLUSHDB")

	return
}

func (api *RedisApi) FlushAll() (err error) {
	// 获取一条Redis连接
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	_, err = redisConn.Do("FLUSHALL")

	return
}

func (api *RedisApi) MGetBit(key string, offsets []int) (value []int, err error) {
	redisConn := api.RedisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(redisConn)
	args := []interface{}{key}
	for i := range offsets {
		args = append(args, "GET", "u1", offsets[i])
	}

	value, err = redis.Ints(redisConn.Do("BITFIELD", args...))
	if err != nil {
		return nil, err
	}

	return value, err
}
