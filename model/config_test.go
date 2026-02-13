package model_test

import (
	"strings"
	"testing"

	"github.com/rromanowicz/mockery/model"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name          string // description of this test case
		expectedError string
		input         model.Config
	}{
		{"Valid empty Config", "", validEmptyConfig},
		{"Valid SqLite Config", "", validFullSqLite},
		{"Valid Postgres Config", "", validFullPostgres},
		{"Valid InMemory Config", "", validFullInMemory},
		{"Missing SqLite Connection String", model.MissingConnectionString, missingSqLiteConnStr},
		{"Missing Postgres Connection String", model.MissingConnectionString, missingPostgresConnStr},
		{"Unsupported DB Type", model.UnsupportedDBType, unsupportedDBType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.input.Validate()
			if gotErr != nil {
				if !strings.Contains(gotErr.Error(), tt.expectedError) {
					t.Errorf("Validate() failed, expected error [%v] not found: %v", tt.expectedError, gotErr)
				}
			} else {
				if len(tt.expectedError) != 0 {
					t.Errorf("Validate() failed: %v", gotErr)
				}
			}
			if !checkDefaultValues(tt.input) {
				t.Errorf("CheckDefaultValues() failed: %v", gotErr)
			}
		})
	}
}

func checkDefaultValues(config model.Config) bool {
	return config.ImportDir == model.ImportDir &&
		config.ExportDir == model.ExportDir &&
		config.Port != 0
}

var (
	validEmptyConfig       = model.Config{}
	validFullSqLite        = model.Config{DBType: "SqLite", DBConfig: model.DBConfig{SqLite: model.DBParams{ConnectionString: "file:test.db"}}}
	validFullPostgres      = model.Config{DBType: "Postgres", DBConfig: model.DBConfig{SqLite: model.DBParams{ConnectionString: "postgresql://test@test..."}}}
	validFullInMemory      = model.Config{DBType: "InMemory"}
	missingSqLiteConnStr   = model.Config{DBType: "SqLite"}
	missingPostgresConnStr = model.Config{DBType: "Postgres"}
	unsupportedDBType      = model.Config{DBType: "TEST"}
)
