// Package context
package context

import (
	"fmt"
	"log"

	"github.com/rromanowicz/mockery/db"
	"github.com/rromanowicz/mockery/model"
	"github.com/rromanowicz/mockery/service"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	SqLite   = db.MockRepoImpl{}
	Postgres = db.MockRepoImpl{}
)

type Context struct {
	Repository  db.MockRepoInt
	MockService service.MockInt
}

func InitContext(config model.Config) (Context, error) {
	var repo db.MockRepoInt
	var dbParams model.DBParams
	var dbDriverFn func(str string) gorm.Dialector
	dbType := model.Database(config.DBType)
	switch dbType {
	case model.SqLite:
		repo = SqLite
		dbParams = config.DBConfig.SqLite
	case model.Postgres:
		repo = Postgres
		dbParams = config.DBConfig.Postgres
	default:
		if len(dbType) == 0 {
			return Context{}, fmt.Errorf("db type not set")
		}
		return Context{}, fmt.Errorf("unsupported DB Type")
	}

	switch config.DBType {
	case model.SqLite:
		dbDriverFn = sqlite.Open
		dbParams = config.DBConfig.SqLite
	case model.Postgres:
		dbDriverFn = postgres.Open
		dbParams = config.DBConfig.Postgres
	}

	log.Printf("Starting server [Port: %v, DB: %s]", config.Port, config.DBType)

	service := service.InitMockService(repo, dbDriverFn, dbParams)
	if config.AutoImport {
		imported, err := service.Import()
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Imported mocks: %v", imported)
		}
	}

	return Context{
		Repository:  repo,
		MockService: service,
	}, nil
}

func (ctx Context) Close() {
	ctx.Repository.CloseDB()
}
