package geo

import "errors"

type Geo interface {
	Add(key string, data []*Member) error
	Del(key string, names ...string) error
	Hash(key string, list ...string) ([]string, error)
	Pos(key string, names ...string) ([]*Member, error)
	Dist(key, member1, member2 string, unit Unit) (float64, error)
	Radius(key string, coord Coordinate, radius int, unit string, options ...Option) ([]*Neighbor, error)
	RadiusByName(key string, name string, radius int, unit string, options ...Option) ([]*Neighbor, error)
}

var ErrNil = errors.New("nil returned")
