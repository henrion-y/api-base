package geo

import "errors"

type Geo interface {
	Add(key string, data []*Member) error
	Pos(key string, names ...string) ([]*Member, error)
	RadiusByName(key string, name string, radius int, unit string, options ...Option) ([]*Neighbor, error)
	Radius(key string, coord Coordinate, radius int, unit string, options ...Option) ([]*Neighbor, error)
	Dist(key, member1, member2 string, unit Unit) (float64, error)
	Hash(key string, list ...string) ([]string, error)
}

var ErrNil = errors.New("nil returned")
