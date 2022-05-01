package cache

import (
	"errors"
)

type Cache interface {
	Get(key string, resultPtr interface{}) error
	BatchGet(keys []string, resultsPtr interface{}) error
	Set(key string, value interface{}, expiryTs uint) error
	SetNX(key string, value interface{}, expiryTs uint) error
	Delete(key string) error
	BatchDelete(keys []string) error
}

var ErrNil = errors.New("nil returned")
