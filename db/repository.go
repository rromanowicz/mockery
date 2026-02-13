package db

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/rromanowicz/mockery/model"
	"gorm.io/gorm"
)

type MockRepoImpl struct {
	DBConn *gorm.DB
	lock   *sync.RWMutex
}

func (mr MockRepoImpl) InitDB(driverFn func(str string) gorm.Dialector, dbParams model.DBParams) MockRepoInt {
	db, err := gorm.Open(driverFn(dbParams.ConnectionString), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&model.Mock{})

	mr.DBConn = db

	return mr
}

func (mr MockRepoImpl) CloseDB() {}
func (mr MockRepoImpl) FindByMethodAndPath(method string, path string) ([]model.Mock, error) {
	mocks, err := gorm.G[model.Mock](mr.DBConn).Where("method=? and path is not null and path=?", method, path).Find(context.Background())
	return mocks, err
}

func (mr MockRepoImpl) FindByID(id int64) (model.Mock, error) {
	ctx := context.Background()
	mock, err := gorm.G[model.Mock](mr.DBConn).Where("id = ?", id).First(ctx)
	return mock, err
}

func (mr MockRepoImpl) FindByIDs(ids []int64) ([]model.Mock, error) {
	idString := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]")
	mocks, err := gorm.G[model.Mock](mr.DBConn).Where("id in (?)", idString).Find(context.Background())
	return mocks, err
}

func (mr MockRepoImpl) DeleteByID(id int64) error {
	_, err := gorm.G[model.Mock](mr.DBConn).Where("id = ?", id).Delete(context.Background())
	return err
}

func (mr MockRepoImpl) Save(mock model.Mock) (model.Mock, error) {
	result := mr.DBConn.Save(&mock)
	if result.Error != nil {
		log.Println(result.Error)
	}
	return mock, result.Error
}

func (mr MockRepoImpl) GetAll() ([]model.Mock, error) {
	mocks, err := gorm.G[model.Mock](mr.DBConn).Find(context.Background())
	return mocks, err
}

func (mr MockRepoImpl) Import() ([]string, error) {
	mocks, files, err := ImportMocks()
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

func (mr MockRepoImpl) Export() ([]string, error) {
	mocks, err := mr.GetAll()
	if err != nil {
		log.Println("Failed to fetch mocks.")
		return []string{}, err
	}
	files, err := ExportMocks(mocks)
	if err != nil {
		log.Println("Failed to save mock.")
		return []string{}, err
	}
	return files, nil
}

func (mr MockRepoImpl) GetRegexpMatchers(method string) ([]model.RegexMatcher, error) {
	mocks, err := gorm.G[model.RegexMatcher](mr.DBConn).Raw("select id, method, regex_path from mocks where method=? and regex_path is not null and regex_path != ''", method).Find(context.Background())
	return mocks, err
}
