package repository

import (
	"api-base/domain/model"
	"context"
)

type BaseRepository interface {
	Create(c context.Context, mod model.Model) error
	Update(c context.Context, mod model.Model, data map[string]interface{}, where string, args ...interface{}) error
	Delete(c context.Context, mod model.Model, where string, args ...interface{}) error
	FindOne(c context.Context, mod interface{}, order string, where string, args ...interface{}) error
	MultiGet(c context.Context, mod model.Model, result interface{}, order string, where string, args ...interface{}) error
	MultiGetByPage(c context.Context, mod model.Model, result interface{}, page, pageCount int, order string, where string, args ...interface{}) error
	Count(c context.Context, mod model.Model, where string, args ...interface{}) (int, error)
	// 差一个深度翻页
}
