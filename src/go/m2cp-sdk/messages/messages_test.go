package messages

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"m2cp"
	"m2cp/contextplus"
	"math/rand"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	ctp m2cp.ContextPlus
}

func (t *TestSuite) SetupSuite() {
	fmt.Println(">>> From SetupSuite")
}

func (t *TestSuite) TearDownSuite() {
	fmt.Println(">>> From TearDownSuite")
}

func (t *TestSuite) SetupTest() {
	fmt.Println("-- From SetupTest")
	t.ctp = contextplus.NewContextPlus()
}

func (t *TestSuite) TearDownTest() {
	fmt.Println("-- From TearDownTest")
}

func TestDaemonTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (t *TestSuite) messageFromBody(mType, body string) string {
	return "{'Header': {'Type': '" + mType + "','Topic': 'foo/bar','Timestamp': '123456'," +
		"'Id': '4bf12e81-a907-4905-902f-bd6b7fa8a942','Origin': 'node123.device456'},'Body': " + body + "}"
}

func (t *TestSuite) randomParameters() []map[string]string {
	rand.Seed(time.Now().UnixNano())
	length := rand.Intn(10) + 1 // Generate a random length between 1 and 10

	dataSlice := make([]map[string]string, length)

	for i := 0; i < length; i++ {
		data := make(map[string]string)

		// Generate random key-value pairs for each map
		for j := 0; j < rand.Intn(5)+1; j++ { // Random number of key-value pairs between 1 and 5
			key := generateRandomString(rand.Intn(10) + 1)
			value := generateRandomString(rand.Intn(10))
			data[key] = value
		}
		dataSlice[i] = data
	}

	return dataSlice
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-=_+"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
