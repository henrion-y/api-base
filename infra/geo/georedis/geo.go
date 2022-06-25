package georedis

import (
	"fmt"
	"reflect"

	geo2 "github.com/henrion-y/api-base/infra/geo"

	"github.com/gomodule/redigo/redis"
)

// Geo is the core service for geolocation-related operation
type Geo struct {
	pool *redis.Pool
}

// Add adds key and related meta data to redis
func (s *Geo) Add(key string, data []*geo2.Member) error {
	conn := s.pool.Get()
	defer conn.Close()

	for _, d := range data {
		_, err := conn.Do("GEOADD", key, d.Coordinate.Lon, d.Coordinate.Lat, d.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

// Pos gets the meta data by key
// returned meta data hase the same order of names
// leave nil for the keys have no data
func (s *Geo) Pos(key string, names ...string) ([]*geo2.Member, error) {
	conn := s.pool.Get()
	defer conn.Close()

	// get data from redis
	args := []interface{}{key}
	for i := range names {
		args = append(args, names[i])
	}
	r, err := redis.Positions(conn.Do("GEOPOS", args...))
	if err != nil {
		return nil, err
	}

	// create meta data
	data := make([]*geo2.Member, len(r))
	for i := range r {
		if r[i] == nil {
		} else {
			data[i] = geo2.NewMember(names[i], r[i][geo2.LatIdx], r[i][geo2.LonIdx])
		}
	}

	return data, nil
}

// RadiusByName find nearby members of member
// the result include the name itself
func (s *Geo) RadiusByName(key string, name string, radius int, unit string, options ...geo2.Option) ([]*geo2.Neighbor, error) {
	mems, err := s.Pos(key, name)
	if err != nil {
		return nil, err
	}
	if len(mems) != 1 {
		return nil, fmt.Errorf("have multiple or zero results, key: %v, name: %v, members: %v", key, name, mems)
	}

	return s.Radius(key, mems[0].Coordinate, radius, unit, options...)
}

// Radius find the neighbor with coordinate
func (s *Geo) Radius(key string, coord geo2.Coordinate, radius int, unit string, options ...geo2.Option) ([]*geo2.Neighbor, error) {
	conn := s.pool.Get()
	defer conn.Close()

	// basic command
	args := []interface{}{key, coord.Lon, coord.Lat, radius, unit}

	// set options
	for _, opt := range options {
		args = append(args, geo2.OptMap[opt])
	}

	// execute command
	r, err := conn.Do("GEORADIUS", args...)
	if err != nil {
		return nil, err
	}

	return rawToNeighbors(r, options...)
}

// Dist cc todo 不存在的点要做兼容
func (s *Geo) Dist(key, member1, member2 string, unit string) (float64, error) {
	conn := s.pool.Get()
	defer conn.Close()

	r, err := conn.Do("GEODIST", key, member1, member2, unit)
	if err != nil {
		return 0, err
	}
	if r == nil {
		return 0, geo2.ErrNil
	}

	v := reflect.ValueOf(r)
	f, err := toFloat64(v)
	if err != nil {
		return 0, err
	}

	return f, nil
}

// Hash return the geohash of place
func (s *Geo) Hash(key string, list ...string) ([]string, error) {
	conn := s.pool.Get()
	defer conn.Close()

	args := []interface{}{key}
	for _, l := range list {
		args = append(args, l)
	}
	r, err := conn.Do("GEOHASH", args...)
	if err != nil {
		return nil, err
	}
	v := reflect.ValueOf(r)
	hashs := make([]string, len(list))
	for i := 0; i < v.Len(); i++ {
		hashv := unpackValue(v.Index(i))
		hash, err := toString(hashv)
		if err != nil {
			return nil, err
		}
		hashs[i] = hash
	}
	return hashs, nil
}

// Del 删除地理位置
func (s *Geo) Del(key string, names ...string) error {
	conn := s.pool.Get()
	defer conn.Close()

	for _, name := range names {
		_, err := conn.Do("ZREM", key, name)
		if err != nil {
			return err
		}
	}

	return nil
}
