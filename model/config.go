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
	DBType     Database `json:"dbType" yaml:"dbType"`
	Port       int      `json:"port" yaml:"port"`
	DBConfig   DBConfig `json:"dbConfig" yaml:"dbConfig"`
	ExportDir  string   `json:"exportDir" yaml:"exportDir"`
	ImportDir  string   `json:"importDir" yaml:"importDir"`
	AutoImport bool     `json:"autoImport" yaml:"autoImport"`
}

type DBConfig struct {
	SqLite   DBParams `json:"sqlite" yaml:"sqlite"`
	Postgres DBParams `json:"postgres" yaml:"postgres"`
}

type DBParams struct {
	ConnectionString string `json:"connStr" yaml:"connStr"`
}

func (c *Config) Validate() error {
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
