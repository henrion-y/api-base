package redis

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/henrion-y/api-base/cache"
	"reflect"
)

type Cache struct {
	r *redis.Pool
}

func (m *Cache) Get(key string, resultPtr interface{}) error {

	c := m.r.Get()
	defer c.Close()

	data, err := redis.Bytes(c.Do("GET", key))
	if err == redis.ErrNil {
		return cache.ErrNil
	} else if err != nil {
		return err
	}
	return json.Unmarshal(data, resultPtr)
}

func (m *Cache) BatchGet(keys []string, resultsPtr interface{}) error {

	realKeys := make([]interface{}, len(keys))
	for i, key := range keys {
		realKeys[i] = key
	}

	c := m.r.Get()
	defer c.Close()

	var replies interface{}
	var err error
	if replies, err = c.Do("MGET", realKeys...); err != nil {
		return err
	}

	sliceType := reflect.TypeOf(resultsPtr).Elem()
	valuePtrType := sliceType.Elem()
	valueType := valuePtrType.Elem()
	slice := reflect.MakeSlice(sliceType, 0, 0)
	for _, v := range replies.([]interface{}) {
		var data []byte
		if data, err = redis.Bytes(v, nil); err == redis.ErrNil {
			slice = reflect.Append(slice, reflect.Zero(valuePtrType))
			continue
		} else if err != nil {
			return err
		}
		valuePtr := reflect.New(valueType)
		if err = json.Unmarshal(data, valuePtr.Interface()); err != nil {
			return err
		}
		slice = reflect.Append(slice, valuePtr)
	}
	reflect.ValueOf(resultsPtr).Elem().Set(slice)
	return nil
}

func (m *Cache) Set(key string, value interface{}, expireTs uint) error {

	c := m.r.Get()
	defer c.Close()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if expireTs > 0 {
		_, err = c.Do("SETEX", key, expireTs, value)
	} else {
		_, err = c.Do("SET", key, data)
	}
	return err

}

func (m *Cache) SetNX(key string, value interface{}, expiryTs uint) error {

	c := m.r.Get()
	defer c.Close()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = c.Do("SETNX", key, expiryTs, data)
	return err
}

func (m *Cache) Delete(key string) error {

	c := m.r.Get()
	defer c.Close()

	_, err := c.Do("DEL", key)
	return err
}

func (m *Cache) BatchDelete(keys []string) error {

	realKeys := make([]interface{}, len(keys))
	for i, key := range keys {
		realKeys[i] = key
	}

	c := m.r.Get()
	defer c.Close()

	_, err := c.Do("DEL", realKeys...)
	return err
}

func (m *Cache) Close() error {
	return m.r.Close()
}
