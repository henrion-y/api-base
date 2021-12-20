package gorm

import (
	"context"
	"github.com/henrion-y/api-base/domain/model"
	"github.com/henrion-y/api-base/domain/repository"
	"github.com/jinzhu/gorm"
)

type BaseRepository struct {
	DB *gorm.DB
}

func NewBaseRepository(db *gorm.DB) repository.BaseRepository {
	return &BaseRepository{db}
}

func (r *BaseRepository) Create(c context.Context, mod model.Model) error {
	db := r.DB.Create(mod)
	return db.Error
}

func (r *BaseRepository) Update(c context.Context, mod model.Model, data map[string]interface{}, where string, args ...interface{}) error {
	return r.DB.Model(mod).Where(where, args...).Update(data).Error
}

func (r *BaseRepository) Delete(c context.Context, mod model.Model, where string, args ...interface{}) error {
	return r.DB.Where(where, args...).Delete(mod).Error
}

func (r *BaseRepository) FindOne(c context.Context, mod interface{}, order string, where string, args ...interface{}) error {
	db := r.DB.Where(where, args...).Order(order).Limit(1).Find(mod)
	if db.Error != nil && db.RecordNotFound() {
		return nil
	}
	return db.Error
}

func (r *BaseRepository) MultiGet(c context.Context, mod model.Model, result interface{}, order string, where string, args ...interface{}) error {
	db := r.DB.Table(mod.TableName()).Where(where, args...).Order(order).Scan(result)
	if db.RecordNotFound() {
		return nil
	}
	return db.Error
}

func (r *BaseRepository) MultiGetByPage(c context.Context, mod model.Model, result interface{}, page, pageCount int, order string, where string, args ...interface{}) error {
	db := r.DB.Table(mod.TableName()).Where(where, args...).Order(order)
	if pageCount > 0 {
		db = db.Limit(pageCount)
	}
	if page > 1 {
		db = db.Offset((page - 1) * pageCount)
	}

	db = db.Scan(result)
	if db.RecordNotFound() {
		return nil
	}
	return db.Error
}

func (r *BaseRepository) Count(c context.Context, mod model.Model, where string, args ...interface{}) (int, error) {
	var count int
	db := r.DB.Table(mod.TableName()).Where(where, args...).Count(&count)
	return count, db.Error
}
