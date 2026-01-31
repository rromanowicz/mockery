// Package service
package service

import "database/sql"

type MockInt interface {
	Get(method string, path string) ([]Mock, error)
	Add(mock Mock) (Mock, error)
	Delete(id int64) error
	List() ([]Mock, error)
}

type MockService struct {
	Repository MockRepoInt
}

func InitMockService(DBConn *sql.DB) MockService {
	return MockService{Repository: MockRepository{DBConn: DBConn}}
}

func (ms MockService) Get(method string, path string) ([]Mock, error) {
	return ms.Repository.findByMethodAndPath(method, path)
}

func (ms MockService) Add(mock Mock) (Mock, error) {
	return ms.Repository.save(mock)
}

func (ms MockService) Delete(id int64) error {
	return ms.Repository.deleteByID(id)
}

func (ms MockService) List() ([]Mock, error) {
	return ms.Repository.getAll()
}
