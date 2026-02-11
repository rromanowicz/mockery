// Package db
package db

import (
	"github.com/rromanowicz/mockery/model"
	"github.com/rromanowicz/mockery/util"
	"gorm.io/gorm"
)

type MockRepoInt interface {
	InitDB(driverFn func(str string) gorm.Dialector, dbParams model.DBParams) MockRepoInt
	CloseDB()
	FindByMethodAndPath(method string, path string) ([]model.Mock, error)
	FindByID(id int64) (model.Mock, error)
	FindByIDs(ids []int64) ([]model.Mock, error)
	DeleteByID(id int64) error
	Save(mock model.Mock) (model.Mock, error)
	GetAll() ([]model.Mock, error)
	Import() ([]string, error)
	Export() ([]string, error)
	GetRegexpMatchers(method string) ([]model.RegexMatcher, error)
}

func ExportMocks(mocks []model.Mock) ([]string, error) {
	return util.Export(model.ExportDir, mocks)
}

func ImportMocks() ([]model.Mock, []string, error) {
	return util.Import(model.ImportDir)
}
