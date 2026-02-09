// Package context
package context

import (
	"github.com/rromanowicz/mockery/db"
	"github.com/rromanowicz/mockery/db/postgres"
	"github.com/rromanowicz/mockery/db/sqlite"
	"github.com/rromanowicz/mockery/db/sqliteorm"
	"github.com/rromanowicz/mockery/model"
	"github.com/rromanowicz/mockery/service"
)

var (
	SqLite    = sqlite.SqLiteRepository{}
	SqLiteORM = sqliteorm.SqLiteORMRepository{}
	Postgres  = postgres.PostgresRepository{}
)

type Context struct {
	Repository  db.MockRepoInt
	MockService service.MockInt
}

func InitContext(dbParams model.DBParams, repo db.MockRepoInt) Context {
	return Context{
		Repository:  repo,
		MockService: service.InitMockService(repo, dbParams),
	}
}

func (ctx Context) Close() {
	ctx.Repository.CloseDB()
}
