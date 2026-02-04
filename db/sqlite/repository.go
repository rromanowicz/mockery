// Package sqlite
package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rromanowicz/mockery/db"
	"github.com/rromanowicz/mockery/model"
)

type SqLiteRepository struct {
	DBConn *sql.DB
	lock   *sync.RWMutex
}

func (mr SqLiteRepository) InitDB() db.MockRepoInt {
	log.Println("Initializing SqLite repository.")

	DB, err := sql.Open("sqlite3", "file:app.db?cache=shared&mode=rwc&_journal_mode=WAL")
	if err != nil {
		log.Fatal(err.Error())
	}
	sqlStmt := `
 CREATE TABLE IF NOT EXISTS mock (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  method TEXT,
  path TEXT,
  regex_path TEXT,
	request_header_matchers TEXT,
	request_query_matchers TEXT,
	request_body_matchers TEXT,
	response_status INTEGER,
	response_body TEXT
 );`

	_, err = DB.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("Error creating table: %q: %s\n", err, sqlStmt)
	}
	mr.DBConn = DB
	mr.lock = &sync.RWMutex{}
	return mr
}

func (mr SqLiteRepository) CloseDB() {
	mr.DBConn.Close()
}

func (mr SqLiteRepository) FindByMethodAndPath(method string, path string) ([]model.Mock, error) {
	rows, err := mr.DBConn.Query("select * from mock where method=? and path is not null and path=?", method, path)
	if err != nil {
		log.Println(err.Error())
		return []model.Mock{}, err
	}
	defer rows.Close()
	return parseResult(rows), nil
}

func (mr SqLiteRepository) GetAll() ([]model.Mock, error) {
	rows, err := mr.DBConn.Query("select * from mock")
	if err != nil {
		log.Println(err.Error())
		return []model.Mock{}, err
	}
	defer rows.Close()
	return parseResult(rows), nil
}

func (mr SqLiteRepository) FindByID(id int64) (model.Mock, error) {
	rows, err := mr.DBConn.Query("select * from mock where id=?", id)
	if err != nil {
		log.Println(err.Error())
		return model.Mock{}, err
	}
	result := parseResult(rows)
	if len(result) == 0 {
		log.Printf("Mock [id=%v] not found.", id)
		return model.Mock{}, nil
	}
	return result[0], nil
}

func (mr SqLiteRepository) FindByIDs(ids []int64) ([]model.Mock, error) {
	idString := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]")
	rows, err := mr.DBConn.Query("select * from mock where id in (?)", idString)
	if err != nil {
		log.Println(err.Error())
		return []model.Mock{}, err
	}
	results := parseResult(rows)
	if len(results) == 0 {
		log.Printf("Mock [id in (%v)] not found.", ids)
		return []model.Mock{}, nil
	}
	return results, nil
}

func (mr SqLiteRepository) DeleteByID(id int64) error {
	defer mr.lock.Unlock()
	mr.lock.Lock()
	_, err := mr.DBConn.Exec("delete from mock where id=?", id)
	return err
}

func (mr SqLiteRepository) Save(mock model.Mock) (model.Mock, error) {
	defer mr.lock.Unlock()
	mr.lock.Lock()
	responseBodyJs, _ := json.Marshal(mock.ResponseBody)
	bodyMatchersJs, _ := json.Marshal(mock.RequestBodyMatchers)
	queryMatchersJs, _ := json.Marshal(mock.RequestQueryMatchers)
	headerMatchersJs, _ := json.Marshal(mock.RequestHeaderMatchers)
	result, err := mr.DBConn.Exec(
		`insert into mock(method, path, regex_path, request_header_matchers, request_query_matchers, request_body_matchers, response_status, response_body)
		values (?, ?, ?, ?, ?, ?, ?, ?)`,
		mock.Method, mock.Path, mock.RegexPath, string(headerMatchersJs), string(queryMatchersJs), string(bodyMatchersJs), mock.ResponseStatus, string(responseBodyJs))
	if err != nil {
		log.Printf("Failed to save. %s", err.Error())
		return model.Mock{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Saved but could not fetch inserted row id. %s", err.Error())
		return mock, err
	}
	return mr.FindByID(id)
}

func (mr SqLiteRepository) Import() ([]string, error) {
	mocks, files, err := db.ImportMocks()
	if err != nil {
		log.Println("Failed to read mocks.")
		return []string{}, err
	}
	for i := range mocks {
		_, err = mr.Save(mocks[i])
		if err != nil {
			log.Println("Failed to save mock.")
			return []string{}, err
		}
	}
	return files, nil
}

func (mr SqLiteRepository) Export() ([]string, error) {
	mocks, err := mr.GetAll()
	if err != nil {
		log.Println("Failed to fetch mocks.")
		return []string{}, err
	}
	files, err := db.ExportMocks(mocks)
	if err != nil {
		log.Println("Failed to save mock.")
		return []string{}, err
	}
	return files, nil
}

func (mr SqLiteRepository) GetRegexpMatchers(method string) ([]model.RegexMatcher, error) {
	rows, err := mr.DBConn.Query("select id, method, regex_path from mock where method=? and regex_path is not null and regex_path != ''", method)
	if err != nil {
		log.Println(err.Error())
		return []model.RegexMatcher{}, err
	}
	defer rows.Close()
	matchers := []model.RegexMatcher{}
	for rows.Next() {
		var matcher model.RegexMatcher
		var regex string
		err := rows.Scan(&matcher.ID, &matcher.Method, &regex)
		if err != nil {
			log.Println(err.Error())
		}
		compiledRegexp, err := regexp.Compile(regex)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		matcher.Regexp = compiledRegexp
		matchers = append(matchers, matcher)
	}
	return matchers, nil
}

func parseResult(rows *sql.Rows) []model.Mock {
	mocks := []model.Mock{}
	for rows.Next() {
		var mock model.Mock
		var bodyMatchers string
		var queryMatchers string
		var headerMatchers string
		var response string
		err := rows.Scan(&mock.ID, &mock.Method, &mock.Path, &mock.RegexPath, &headerMatchers, &queryMatchers, &bodyMatchers, &mock.ResponseStatus, &response)
		if err != nil {
			log.Println(err.Error())
		}
		var parsedBodyMatchers []model.BodyMatcher
		json.Unmarshal([]byte(bodyMatchers), &parsedBodyMatchers)
		mock.RequestBodyMatchers = parsedBodyMatchers

		var parsedQueryMatchers []model.QueryMatcher
		json.Unmarshal([]byte(queryMatchers), &parsedQueryMatchers)
		mock.RequestQueryMatchers = parsedQueryMatchers

		var parsedHeaderMatchers []model.HeaderMatcher
		json.Unmarshal([]byte(headerMatchers), &parsedHeaderMatchers)
		mock.RequestHeaderMatchers = parsedHeaderMatchers

		var parsedResponse any
		json.Unmarshal([]byte(strings.ReplaceAll(response, "\\\"", "\"")), &parsedResponse)
		mock.ResponseBody = parsedResponse

		mocks = append(mocks, mock)
	}
	return mocks
}
