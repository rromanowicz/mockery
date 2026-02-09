// Package orm
package orm

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/rromanowicz/mockery/db"
	"github.com/rromanowicz/mockery/model"
	"gorm.io/gorm"
)

type OrmRepository struct {
	DBConn *gorm.DB
	lock   *sync.RWMutex
}

func (mr OrmRepository) InitDB(dbParams model.DBParams) db.MockRepoInt {
	log.Println("Initializing Postgres repository.")

	db, err := gorm.Open(dbParams.Driver(dbParams.ConnectionString), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&model.Mock{})

	mr.DBConn = db

	return mr
}

func (mr OrmRepository) CloseDB() {}
func (mr OrmRepository) FindByMethodAndPath(method string, path string) ([]model.Mock, error) {
	mocks, err := gorm.G[model.Mock](mr.DBConn).Where("method=? and path is not null and path=?", method, path).Find(context.Background())
	return mocks, err
}

func (mr OrmRepository) FindByID(id int64) (model.Mock, error) {
	ctx := context.Background()
	mock, err := gorm.G[model.Mock](mr.DBConn).Where("id = ?", id).First(ctx)
	return mock, err
}

func (mr OrmRepository) FindByIDs(ids []int64) ([]model.Mock, error) {
	idString := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]")
	mocks, err := gorm.G[model.Mock](mr.DBConn).Where("id in (?)", idString).Find(context.Background())
	return mocks, err
}

func (mr OrmRepository) DeleteByID(id int64) error {
	_, err := gorm.G[model.Mock](mr.DBConn).Where("id = ?", id).Delete(context.Background())
	return err
}

func (mr OrmRepository) Save(mock model.Mock) (model.Mock, error) {
	err := gorm.G[model.Mock](mr.DBConn).Create(context.Background(), &mock)
	if err != nil {
		panic(err)
	}
	return mock, err
}

func (mr OrmRepository) GetAll() ([]model.Mock, error) {
	mocks, err := gorm.G[model.Mock](mr.DBConn).Find(context.Background())
	return mocks, err
}

func (mr OrmRepository) Import() ([]string, error) {
	mocks, files, err := db.ImportMocks()
	if err != nil {
		log.Println("Failed to read mocks.")
		return []string{}, err
	}
	for i := range mocks {
		_, err = mr.Save(mocks[i])
		if err != nil {
			log.Println("Failed to save mock.")
			return []string{}, err
		}
	}
	return files, nil
}

func (mr OrmRepository) Export() ([]string, error) {
	mocks, err := mr.GetAll()
	if err != nil {
		log.Println("Failed to fetch mocks.")
		return []string{}, err
	}
	files, err := db.ExportMocks(mocks)
	if err != nil {
		log.Println("Failed to save mock.")
		return []string{}, err
	}
	return files, nil
}

func (mr OrmRepository) GetRegexpMatchers(method string) ([]model.RegexMatcher, error) {
	mocks, err := gorm.G[model.RegexMatcher](mr.DBConn).Raw("select id, method, regex_path from mocks where method=? and regex_path is not null and regex_path != ''", method).Find(context.Background())
	return mocks, err
}
