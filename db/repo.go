// Package db
package db

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"

	"github.com/rromanowicz/mockery/model"
)

const dir = "./stubs"

type MockRepoInt interface {
	InitDB() MockRepoInt
	CloseDB()
	FindByMethodAndPath(method string, path string) ([]model.Mock, error)
	FindByID(id int64) (model.Mock, error)
	FindByIDs(ids []int64) ([]model.Mock, error)
	DeleteByID(id int64) error
	Save(mock model.Mock) (model.Mock, error)
	GetAll() ([]model.Mock, error)
	Import() ([]string, error)
	Export() ([]string, error)
	GetRegexpMatchers(method string) ([]model.RegexMatcher, error)
}

func ExportMocks(mocks []model.Mock) ([]string, error) {
	os.MkdirAll(dir, os.ModePerm)
	var files []string
	for i := range mocks {
		mockData, err := json.Marshal(mocks[i])
		if err != nil {
			log.Printf("Failed to marshall mock [id=%v]. Error: %v", mocks[i].ID, err.Error())
			continue
		}
		var urlPath string
		if len(mocks[i].Path) != 0 {
			urlPath = strings.ReplaceAll(mocks[i].Path, "/", "_")
		} else {
			urlPath = strings.ReplaceAll(strings.ReplaceAll(mocks[i].RegexPath, "/", "_"), "\\", "")
		}
		fileName := fmt.Sprintf("%s/%v_%s%s.json", dir, mocks[i].ID, mocks[i].Method, urlPath)
		err = writeFile(fileName, mockData)
		if err != nil {
			log.Printf("Failed to write mock [%s]. Error: %v", fileName, err.Error())
			continue
		}
		files = append(files, fileName)
	}
	return files, nil
}

func ImportMocks() ([]model.Mock, []string, error) {
	files, err := listFiles()
	if err != nil {
		return []model.Mock{}, []string{}, err
	}
	var mocks []model.Mock
	for i := range files {
		contents, err := readFile(files[i])
		if err != nil {
			log.Printf("Failed to read [%v]. Error: %v", files[i], err.Error())
			continue
		}
		var mock model.Mock
		err = json.Unmarshal(contents, &mock)
		if err != nil {
			log.Printf("Failed to parse [%v], Error: %v", files[i], err.Error())
			continue
		}
		mocks = append(mocks, mock)
	}

	return mocks, files, nil
}

func listFiles() ([]string, error) {
	root := os.DirFS(dir)
	mdFiles, err := fs.Glob(root, "*.json")
	if err != nil {
		log.Println(err.Error())
		return []string{}, err
	}

	var files []string
	for _, v := range mdFiles {
		files = append(files, path.Join(dir, v))
	}
	return files, nil
}

func readFile(path string) ([]byte, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return []byte{}, err
	}
	return contents, nil
}

func writeFile(path string, contents []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(contents)
	if err != nil {
		return err
	}

	return nil
}
