package persistence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
		return "", fmt.Errorf("could not marshal struct into storage format (JSON): %w", err)
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
		return "", fmt.Errorf("could not resolve path: %w", err)
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
		return fmt.Errorf("could not resolve path: %w", err)
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
		return fmt.Errorf("could not load struct: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	err = decoder.Decode(pointer)
	if err != nil {
		return fmt.Errorf("could not decode JSON into struct: %w", err)
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
		return "", fmt.Errorf("could not resolve path: %w", err)
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
		return nil, fmt.Errorf("could not resolve path: %w", err)
	}

	truePath := filepath.Join(entryPath, path)

	data, err := loadFromFile(truePath)
	if err != nil {
		return []byte{}, fmt.Errorf("could not load bytes: %w", err)
	}

	return data, nil
}

func saveToFile(data []byte, absPath string) (string, error) {

	dir := filepath.Dir(absPath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("could not create directory: %w", err)
	}

	file, err := os.Create(absPath)
	if err != nil {
		return "", fmt.Errorf("could not create file at path %s: %w", absPath, err)
	}

	_, err = file.Write(data)
	if err != nil {
		return "", fmt.Errorf("could not write to file at path %s: %w", absPath, err)
	}

	err = file.Close()
	if err != nil {
		return absPath, fmt.Errorf("could not close file at path %s: %w", absPath, err)
	}
	return absPath, nil
}

func loadFromFile(absPath string) ([]byte, error) {

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("could not read file at path %s: %w", absPath, err)
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
		return fmt.Errorf("path %s is outside of base path %s", targetPath, basePath)
	}

	return nil
}
