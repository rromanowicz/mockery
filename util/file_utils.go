// Package util
package util

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

func Export(exportDir string, mocks []model.Mock) ([]string, error) {
	os.MkdirAll(exportDir, os.ModePerm)
	var files []string
	for i := range mocks {
		mockData, err := json.Marshal(mocks[i])
		if err != nil {
			log.Printf("Failed to marshal mock [id=%v]. Error: %v", mocks[i].ID, err.Error())
			continue
		}
		var urlPath string
		if len(mocks[i].Path) != 0 {
			urlPath = strings.ReplaceAll(mocks[i].Path, "/", "_")
		} else {
			urlPath = strings.ReplaceAll(strings.ReplaceAll(mocks[i].RegexPath, "/", "_"), "\\", "")
		}
		fileName := fmt.Sprintf("%s/%v_%s%s.json", exportDir, mocks[i].ID, mocks[i].Method, urlPath)
		err = writeFile(fileName, mockData)
		if err != nil {
			log.Printf("Failed to write mock [%s]. Error: %v", fileName, err.Error())
			continue
		}
		files = append(files, fileName)
	}
	return files, nil
}

func Import(importDir string) ([]model.Mock, []string, error) {
	files, err := listJSONFiles(importDir)
	if err != nil {
		return []model.Mock{}, []string{}, err
	}
	var mocks []model.Mock
	var importedFiles []string
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
		ok, errors := mock.Validate()
		var importStatus string
		if ok {
			mocks = append(mocks, mock)
			importStatus = "[OK]"
		} else {
			log.Printf("Import failed for file[%s]. Found validation errors [%v]", files[i], errors)
			importStatus = "[FAILED]"
		}
		importedFiles = append(importedFiles, fmt.Sprintf("%-8s %s", importStatus, files[i]))
	}

	return mocks, importedFiles, nil
}

func listJSONFiles(dir string) ([]string, error) {
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
