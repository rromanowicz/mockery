// Package service
package service

import (
	"github.com/rromanowicz/mockery/db"
	"github.com/rromanowicz/mockery/model"
)

type MockInt interface {
	Get(method string, path string) ([]model.Mock, error)
	Add(mock model.Mock) (model.Mock, error)
	Delete(id int64) error
	List() ([]model.Mock, error)
	Import() ([]string, error)
	Export() ([]string, error)
}

type MockService struct {
	Repository db.MockRepoInt
}

func InitMockService(repo db.MockRepoInt) MockService {
	return MockService{Repository: repo.InitDB()}
}

func (ms MockService) Get(method string, path string) ([]model.Mock, error) {
	return ms.Repository.FindByMethodAndPath(method, path)
}

func (ms MockService) Add(mock model.Mock) (model.Mock, error) {
	return ms.Repository.Save(mock)
}

func (ms MockService) Delete(id int64) error {
	return ms.Repository.DeleteByID(id)
}

func (ms MockService) List() ([]model.Mock, error) {
	return ms.Repository.GetAll()
}

func (ms MockService) Import() ([]string, error) {
	return ms.Repository.Import()
}

func (ms MockService) Export() ([]string, error) {
	return ms.Repository.Export()
}
