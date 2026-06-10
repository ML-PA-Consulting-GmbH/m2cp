package persistence

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func SaveStructCommon(structure any, path string) (string, error) {
	return saveStruct(structure, "common", path)
}

func LoadStructCommon(pointer any, path string) error {
	return loadStruct(pointer, "common", path)
}

func SaveStructCurrent(structure any, path string) (string, error) {
	return saveStruct(structure, "current", path)
}

func LoadStructCurrent(pointer any, path string) error {
	return loadStruct(pointer, "current", path)
}

func SaveBytesCommon(data []byte, path string) (string, error) {
	return saveBytes(data, "common", path)
}

func LoadBytesCommon(path string) ([]byte, error) {
	return loadBytes("common", path)
}

func SaveBytesCurrent(data []byte, path string) (string, error) {
	return saveBytes(data, "current", path)
}

func LoadBytesCurrent(path string) ([]byte, error) {
	return loadBytes("current", path)
}

func saveStruct(structure any, location string, path string) (string, error) {
	if reflect.ValueOf(structure).Kind() != reflect.Struct {
		return "", errors.New("structure must be a struct")
	}

	marshalled, err := json.Marshal(structure)
	if err != nil {
		return "", errors.New("could not marshal struct into storage format (JSON): " + err.Error())
	}

	entryPath := ""

	if location == "common" {
		entryPath, err = PathSnapCommon()
	} else if location == "current" {
		entryPath, err = PathSnapCurrent()
	} else {
		return "", errors.New("unknown location")
	}

	if err != nil {
		return "", errors.New("could not resolve path: " + err.Error())
	}

	truePath := filepath.Join(entryPath, path)

	if filepath.Ext(truePath) == "" {
		truePath += ".json"
	}

	err = ensureSubPath(entryPath, truePath)
	if err != nil {
		return "", err
	}

	return saveToFile(marshalled, truePath)
}

func loadStruct(pointer any, location string, path string) error {

	if reflect.ValueOf(pointer).Kind() != reflect.Ptr {
		return errors.New("argument \"pointer\" must be a pointer")
	}
	if reflect.ValueOf(pointer).Elem().Kind() != reflect.Struct {
		return errors.New("pointer must point to a struct")
	}
	if pointer == nil {
		return errors.New("pointer must not be nil")
	}

	var err error
	entryPath := ""

	if location == "common" {
		entryPath, err = PathSnapCommon()
	} else if location == "current" {
		entryPath, err = PathSnapCurrent()
	} else {
		return errors.New("unknown location")
	}

	if err != nil {
		return errors.New("could not resolve path: " + err.Error())
	}

	truePath := filepath.Join(entryPath, path)

	if filepath.Ext(truePath) == "" {
		truePath += ".json"
	}

	err = ensureSubPath(entryPath, truePath)
	if err != nil {
		return err
	}

	data, err := loadFromFile(truePath)

	if err != nil {
		return errors.New("could not load struct: " + err.Error())
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	err = decoder.Decode(pointer)
	if err != nil {
		return errors.New("could not decode JSON into struct: " + err.Error())
	}

	return nil
}

func saveBytes(data []byte, location string, path string) (string, error) {
	var err error
	entryPath := ""

	if location == "common" {
		entryPath, err = PathSnapCommon()
	} else if location == "current" {
		entryPath, err = PathSnapCurrent()
	} else {
		return "", errors.New("unknown location")
	}

	if err != nil {
		return "", errors.New("could not resolve path: " + err.Error())
	}

	truePath := filepath.Join(entryPath, path)

	return saveToFile(data, truePath)
}

func loadBytes(location string, path string) ([]byte, error) {

	var err error
	entryPath := ""

	if location == "common" {
		entryPath, err = PathSnapCommon()
	} else if location == "current" {
		entryPath, err = PathSnapCurrent()
	} else {
		return nil, errors.New("unknown location")
	}

	if err != nil {
		return nil, errors.New("could not resolve path: " + err.Error())
	}

	truePath := filepath.Join(entryPath, path)

	data, err := loadFromFile(truePath)
	if err != nil {
		return []byte{}, errors.New("could not load bytes: " + err.Error())
	}

	return data, nil
}

func saveToFile(data []byte, absPath string) (string, error) {

	dir := filepath.Dir(absPath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", errors.New("could not create directory: " + err.Error())
	}

	file, err := os.Create(absPath)
	if err != nil {
		return "", errors.New("could not create file at path: " + absPath + " error: " + err.Error())
	}

	_, err = file.Write(data)
	if err != nil {
		return "", errors.New("could not write to file at path: " + absPath + " error: " + err.Error())
	}

	err = file.Close()
	if err != nil {
		return absPath, errors.New("could not close file at path: " + absPath + " error: " + err.Error())
	}
	return absPath, nil
}

func loadFromFile(absPath string) ([]byte, error) {

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, errors.New("could not read file at path: " + absPath + " error: " + err.Error())
	}

	return data, nil
}

// ensureSubPath: both paths must be absolute
func ensureSubPath(basePath, targetPath string) error {

	if !filepath.IsAbs(basePath) || !filepath.IsAbs(targetPath) {
		return errors.New("both paths must be absolute")
	}

	relPath, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return err
	}

	// Check if relPath starts with ".." which means it's outside basePath
	if strings.HasPrefix(relPath, "..") {
		return errors.New("path: " + targetPath + " is outside of base path: " + basePath)
	}

	return nil
}
