// Package  context
package context

import (
	"database/sql"

	"github.com/rromanowicz/mockery/db"
	"github.com/rromanowicz/mockery/service"
)

var (
	dBConn      *sql.DB         = db.InitDB()
	mockService service.MockInt = service.InitMockService(dBConn)
)

type Context struct {
	DBConn      *sql.DB
	MockService service.MockInt
}

func InitContext() Context {
	return Context{
		DBConn:      dBConn,
		MockService: mockService,
	}
}

func (ctx Context) Close() {
	ctx.DBConn.Close()
}
