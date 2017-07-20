package gormrepo

import (
	"errors"

	"github.com/jinzhu/gorm"
)

var (
	ErrPrimaryNotBlank = errors.New("primary key not blank")
)

type Fields map[string]interface{}
type CriteriaOption func(db *gorm.DB) *gorm.DB

const (
	Find int = iota + 1
	First
	Last
)

func And(query interface{}, args ...interface{}) CriteriaOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args)
	}
}

func Not(query interface{}, args ...interface{}) CriteriaOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Not(query, args)
	}
}

func Or(query interface{}, args ...interface{}) CriteriaOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Or(query, args)
	}
}

func Select(columns interface{}, args ...interface{}) CriteriaOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select(columns, args...)
	}
}

func OrderBy(name string, orientation string, reorder bool) CriteriaOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(name+" "+orientation, reorder)
	}
}

func Limit(limit int) CriteriaOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

func Offset(offset int) CriteriaOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	}
}

func Preload(field string) CriteriaOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(field)
	}
}
