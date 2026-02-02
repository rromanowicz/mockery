// Package context
package context

import (
	"github.com/rromanowicz/mockery/db"
	"github.com/rromanowicz/mockery/db/postgres"
	"github.com/rromanowicz/mockery/db/sqlite"
	"github.com/rromanowicz/mockery/service"
)

var (
	SqLite   = sqlite.SqLiteRepository{}
	Postgres = postgres.PostgresRepository{}
)

type Context struct {
	Repository  db.MockRepoInt
	MockService service.MockInt
}

func InitContext(repo db.MockRepoInt) Context {
	return Context{
		Repository:  repo,
		MockService: service.InitMockService(repo),
	}
}

func (ctx Context) Close() {
	ctx.Repository.CloseDB()
}
