package service

import (
	"database/sql"
	"encoding/json"
	"log"
)

type MockRepoInt interface {
	findByMethodAndPath(method string, path string) ([]Mock, error)
	findByID(id int64) (Mock, error)
	deleteByID(id int64) error
	save(mock Mock) (Mock, error)
	getAll() ([]Mock, error)
}

type MockRepository struct {
	DBConn *sql.DB
}

func (mr MockRepository) findByMethodAndPath(method string, path string) ([]Mock, error) {
	rows, err := mr.DBConn.Query("select * from mock where method=? and path=?", method, path)
	if err != nil {
		log.Println(err.Error())
		return []Mock{}, err
	}
	defer rows.Close()
	return parseResult(rows), nil
}

func (mr MockRepository) getAll() ([]Mock, error) {
	rows, err := mr.DBConn.Query("select * from mock")
	if err != nil {
		log.Println(err.Error())
		return []Mock{}, err
	}
	defer rows.Close()
	return parseResult(rows), nil
}

func (mr MockRepository) findByID(id int64) (Mock, error) {
	rows, err := mr.DBConn.Query("select * from mock where id=?", id)
	if err != nil {
		log.Println(err.Error())
		return Mock{}, err
	}
	result := parseResult(rows)
	if len(result) == 0 {
		log.Printf("Mock [id=%v] not found.", id)
		return Mock{}, nil
	}
	return result[0], nil
}

func (mr MockRepository) deleteByID(id int64) error {
	_, err := mr.DBConn.Exec("delete from mock where id=?", id)
	return err
}

func (mr MockRepository) save(mock Mock) (Mock, error) {
	responseBodyJs, _ := json.Marshal(mock.ResponseBody)
	bodyMatchersJs, _ := json.Marshal(mock.RequestBodyMatchers)
	queryMatchersJs, _ := json.Marshal(mock.RequestQueryMatchers)
	headerMatchersJs, _ := json.Marshal(mock.RequestHeaderMatchers)
	result, err := mr.DBConn.Exec(
		`insert into mock(method, path, request_header_matchers, request_query_matchers, request_body_matchers, response_status, response_body)
		values (?, ?, ?, ?, ?, ?, ?)`,
		mock.Method, mock.Path, string(headerMatchersJs), string(queryMatchersJs), string(bodyMatchersJs), mock.ResponseStatus, string(responseBodyJs))
	if err != nil {
		log.Printf("Failed to save. %s", err.Error())
		return Mock{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Saved but could not fetch inserted row id. %s", err.Error())
		return mock, err
	}
	return mr.findByID(id)
}

func parseResult(rows *sql.Rows) []Mock {
	mocks := []Mock{}
	for rows.Next() {
		var mock Mock
		var bodyMatchers string
		var queryMatchers string
		var headerMatchers string
		err := rows.Scan(&mock.ID, &mock.Method, &mock.Path, &headerMatchers, &queryMatchers, &bodyMatchers, &mock.ResponseStatus, &mock.ResponseBody)
		if err != nil {
			log.Println(err.Error())
		}
		var parsedBodyMatchers []BodyMatcher
		json.Unmarshal([]byte(bodyMatchers), &parsedBodyMatchers)
		mock.RequestBodyMatchers = parsedBodyMatchers

		var parsedQueryMatchers []QueryMatcher
		json.Unmarshal([]byte(queryMatchers), &parsedQueryMatchers)
		mock.RequestQueryMatchers = parsedQueryMatchers

		var parsedHeaderMatchers []HeaderMatcher
		json.Unmarshal([]byte(headerMatchers), &parsedHeaderMatchers)
		mock.RequestHeaderMatchers = parsedHeaderMatchers

		mocks = append(mocks, mock)
	}
	return mocks
}
