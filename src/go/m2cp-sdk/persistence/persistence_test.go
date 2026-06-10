package persistence

import (
	"github.com/stretchr/testify/suite"
	"m2cp"
	"m2cp/m2cp_new"
	"m2cp/tests"
	"os"
	"path/filepath"
	"testing"
)

type friend struct {
	Name string
}

type measures struct {
	Height   int
	ShoeSize int
}

type house struct {
	Location string
}

type person struct {
	Name string
	Age  int

	Measures measures
	Friends  []friend
	House    *house
}
type TestSuite struct {
	suite.Suite
	ctx m2cp.ContextPlus
}

func (s *TestSuite) TearDownSuite() {
	s.ctx.Cancel()

	homeDir, err := os.UserHomeDir()
	s.NoError(err)
	createdPath := filepath.Join(homeDir, ".m2cp", "snaps", "persistence_test")
	err = os.RemoveAll(createdPath)
	s.NoError(err)

}

func (s *TestSuite) SetupSuite() {
	s.ctx = m2cp_new.ContextPlus()
	err := os.Setenv("SNAP", "persistence_test")
	s.NoError(err)

}

func (s *TestSuite) TearDownTest() {
	tests.FreeResource("persistence")
}

func (s *TestSuite) SetupTest() {

	tests.LockResource("persistence")
	homeDir, err := os.UserHomeDir()
	s.NoError(err)
	createdPath := filepath.Join(homeDir, ".m2cp", "snaps", "persistence_test")
	err = os.RemoveAll(createdPath)
	s.NoError(err)

}

func (s *TestSuite) TestSaveStruct() {

	p := person{
		Name: "Stella",
		Age:  30,
		Measures: measures{
			Height:   0,
			ShoeSize: 0,
		},
		Friends: []friend{
			{Name: "Alice"},
			{Name: "Bob"},
		},
		House: &house{
			Location: "London",
		},
	}

	_, err := SaveStructCommon(p, "person")
	s.NoError(err)

	// check if the file was created
	homeDir, err := os.UserHomeDir()
	s.NoError(err)
	createdPath := filepath.Join(homeDir, ".m2cp", "snaps", "persistence_test", "common", "person.json")
	_, err = os.Stat(createdPath)
	s.NoError(err)

	// check if the file contains the correct data

	data := "{\"Name\":\"Stella\",\"Age\":30,\"Measures\":{\"Height\":0,\"ShoeSize\":0},\"Friends\":[{\"Name\":\"Alice\"},{\"Name\":\"Bob\"}],\"House\":{\"Location\":\"London\"}}"
	fileData, err := os.ReadFile(createdPath)
	s.NoError(err)
	s.Equal(data, string(fileData))

}

func (s *TestSuite) TestDifferentFilePaths() {

	p := person{
		Name: "Stella",
		Age:  30,
		Measures: measures{
			Height:   0,
			ShoeSize: 0,
		},
		Friends: []friend{
			{Name: "Alice"},
			{Name: "Bob"},
		},
		House: &house{
			Location: "London",
		},
	}
	homeDir, err := os.UserHomeDir()
	s.NoError(err)

	_, err = SaveStructCommon(p, "person/test.json")
	s.NoError(err)

	createdPath := filepath.Join(homeDir, ".m2cp", "snaps", "persistence_test", "common", "person/test.json")
	_, err = os.Stat(createdPath)
	s.NoError(err)

	_, err = SaveStructCommon(p, "person/test.html")
	s.NoError(err)

	createdPath = filepath.Join(homeDir, ".m2cp", "snaps", "persistence_test", "common", "person/test.html")
	_, err = os.Stat(createdPath)
	s.NoError(err)

	_, err = SaveStructCommon(p, "person/test.bin")
	s.NoError(err)

	createdPath = filepath.Join(homeDir, ".m2cp", "snaps", "persistence_test", "common", "person/test.bin")
	_, err = os.Stat(createdPath)
	s.NoError(err)

	_, err = SaveStructCommon(p, "person/person/")
	s.NoError(err)

	createdPath = filepath.Join(homeDir, ".m2cp", "snaps", "persistence_test", "common", "person/person.json")
	_, err = os.Stat(createdPath)
	s.NoError(err)

}

func (s *TestSuite) TestForbiddenFilePaths() {

	p := person{
		Name: "Stella",
		Age:  30,
		Measures: measures{
			Height:   0,
			ShoeSize: 0,
		},
		Friends: []friend{
			{Name: "Alice"},
			{Name: "Bob"},
		},
		House: &house{
			Location: "London",
		},
	}

	_, err := SaveStructCommon(p, "../../../../../../etc/passwd")
	s.Error(err)

	_, err = SaveStructCommon(p, "")
	s.Error(err)

	_, err = SaveStructCommon(p, "/")
	s.Error(err)
}

func (s *TestSuite) TestLoadStructCommon() {

	p := person{
		Name: "Stella",
		Age:  30,
		Measures: measures{
			Height:   0,
			ShoeSize: 0,
		},
		Friends: []friend{
			{Name: "Alice"},
			{Name: "Bob"},
		},
		House: &house{
			Location: "London",
		},
	}

	_, err := SaveStructCommon(p, "person")
	s.NoError(err)

	var loadedPerson person
	err = LoadStructCommon(&loadedPerson, "person")
	s.NoError(err)

	s.Equal(p, loadedPerson)

}

func (s *TestSuite) TestLoadStructDifferentFilePaths() {

	p := person{
		Name: "Stella",
		Age:  30,
		Measures: measures{
			Height:   0,
			ShoeSize: 0,
		},
		Friends: []friend{
			{Name: "Alice"},
			{Name: "Bob"},
		},
		House: &house{
			Location: "London",
		},
	}

	_, err := SaveStructCommon(p, "person/test.json")
	s.NoError(err)

	var loadedPerson person
	err = LoadStructCommon(&loadedPerson, "person/test.json")
	s.NoError(err)

	s.Equal(p, loadedPerson)

	_, err = SaveStructCommon(p, "person/test.html")
	s.NoError(err)

	err = LoadStructCommon(&loadedPerson, "person/test.html")
	s.NoError(err)

	s.Equal(p, loadedPerson)

	_, err = SaveStructCommon(p, "person/test.bin")
	s.NoError(err)

	err = LoadStructCommon(&loadedPerson, "person/test.bin")
	s.NoError(err)

	s.Equal(p, loadedPerson)

	_, err = SaveStructCommon(p, "person/person/")
	s.NoError(err)

	err = LoadStructCommon(&loadedPerson, "person/person.json")
	s.NoError(err)

	s.Equal(p, loadedPerson)

}

func (s *TestSuite) TestOverwriteStruct() {

	p := person{
		Name: "Stella",
		Age:  30,
		Measures: measures{
			Height:   0,
			ShoeSize: 0,
		},
		Friends: []friend{
			{Name: "Alice"},
			{Name: "Bob"},
		},
		House: &house{
			Location: "London",
		},
	}

	_, err := SaveStructCurrent(p, "person")
	s.NoError(err)

	p.Age = 31
	p.Measures.Height = 170

	_, err = SaveStructCurrent(p, "person")
	s.NoError(err)

	var loadedPerson person
	err = LoadStructCurrent(&loadedPerson, "person")
	s.NoError(err)

	s.Equal(p, loadedPerson)
}

func (s *TestSuite) TestLoadStructWrongStruct() {

	p := person{
		Name: "Stella",
		Age:  30,
		Measures: measures{
			Height:   0,
			ShoeSize: 0,
		},
		Friends: []friend{
			{Name: "Alice"},
			{Name: "Bob"},
		},
		House: &house{
			Location: "London",
		},
	}

	_, err := SaveStructCommon(p, "person")
	s.NoError(err)

	var loadedMeasures measures
	err = LoadStructCommon(&loadedMeasures, "person")
	s.Error(err)
}

func (s *TestSuite) TestSaveBytes() {

	data := []byte("test data")
	_, err := SaveBytesCommon(data, "test")
	s.NoError(err)

	homeDir, err := os.UserHomeDir()
	s.NoError(err)
	createdPath := filepath.Join(homeDir, ".m2cp", "snaps", "persistence_test", "common", "test")
	_, err = os.Stat(createdPath)
	s.NoError(err)

	fileData, err := os.ReadFile(createdPath)
	s.NoError(err)
	s.Equal(data, fileData)
}

func (s *TestSuite) TestLoadBytes() {

	data := []byte("test data")
	_, err := SaveBytesCommon(data, "test")
	s.NoError(err)

	loadedData, err := LoadBytesCommon("test")
	s.NoError(err)
	s.Equal(data, loadedData)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
