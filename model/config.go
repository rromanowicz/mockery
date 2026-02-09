package model

import (
	"fmt"
)

type Database string

const (
	SqLite         Database = "SqLite"
	SqLiteORM      Database = "SqLiteORM"
	Postgres       Database = "Postgres"
	defaultDir     string   = "./stubs"
	defaultPort    int      = 8080
	defaultConnStr string   = "file:mockery.db?cache=shared&mode=rwc&_journal_mode=WAL"
)

type Config struct {
	DBType    Database `json:"dbType"`
	Port      int      `json:"port"`
	DBConfig  DBConfig `json:"dbConfig"`
	ExportDir string   `json:"exportDir"`
	ImportDir string   `json:"importDir"`
}

type DBConfig struct {
	SqLite   DBParams `json:"sqlite"`
	Postgres DBParams `json:"postgres"`
}

type DBParams struct {
	ConnectionString string `json:"connStr"`
}

func (c *Config) Validate(port *int, dbType *string) error {
	if *port != 0 {
		c.Port = *port
	}
	if len(*dbType) != 0 {
		c.DBType = Database(*dbType)
	}
	if len(c.ExportDir) == 0 {
		c.ExportDir = defaultDir
	}
	if len(c.ImportDir) == 0 {
		c.ImportDir = defaultDir
	}
	if c.Port == 0 {
		c.Port = defaultPort
	}
	switch c.DBType {
	case SqLite, SqLiteORM:
		if len(c.DBConfig.SqLite.ConnectionString) == 0 {
			return fmt.Errorf("connection string missing for [%s] connection", c.DBType)
		}
	case Postgres:
		if len(c.DBConfig.Postgres.ConnectionString) == 0 {
			return fmt.Errorf("connection string missing for [%s] connection", c.DBType)
		}
	default:
		c.DBType = SqLite
		c.DBConfig.SqLite.ConnectionString = defaultConnStr
	}
	return nil
}
