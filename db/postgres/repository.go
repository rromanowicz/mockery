// Package postgres
package postgres

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rromanowicz/mockery/db"
	"github.com/rromanowicz/mockery/model"
)

type PostgresRepository struct {
	DBConn *sql.DB
}

func (mr PostgresRepository) InitDB() db.MockRepoInt {
	panic("Unimplemented")
}

func (mr PostgresRepository) CloseDB() {
	mr.DBConn.Close()
}

func (mr PostgresRepository) FindByMethodAndPath(method string, path string) ([]model.Mock, error) {
	return []model.Mock{}, nil
}

func (mr PostgresRepository) GetAll() ([]model.Mock, error) {
	return []model.Mock{}, nil
}

func (mr PostgresRepository) FindByID(id int64) (model.Mock, error) {
	return model.Mock{}, nil
}

func (mr PostgresRepository) DeleteByID(id int64) error {
	return nil
}

func (mr PostgresRepository) Save(mock model.Mock) (model.Mock, error) {
	return model.Mock{}, nil
}

func (mr PostgresRepository) Import() ([]string, error) {
	return []string{}, nil
}

func (mr PostgresRepository) Export() ([]string, error) {
	return []string{}, nil
}
