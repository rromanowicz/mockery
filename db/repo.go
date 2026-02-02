// Package db
package db

import (
	"github.com/rromanowicz/mockery/model"
)

type MockRepoInt interface {
	InitDB() MockRepoInt
	CloseDB()
	FindByMethodAndPath(method string, path string) ([]model.Mock, error)
	FindByID(id int64) (model.Mock, error)
	DeleteByID(id int64) error
	Save(mock model.Mock) (model.Mock, error)
	GetAll() ([]model.Mock, error)
}
