package firmwaredb

import (
	"github.com/stretchr/testify/suite"
	"m2cp"
	coap_server "m2cp/coap/coap-server"
	"m2cp/m2cp_new"
	"testing"
)

type FirmwareDbTestSuite struct {
	suite.Suite
	firmwareDb *FirmwareDb
}

func (t *FirmwareDbTestSuite) SetupTest() {
}

func (t *FirmwareDbTestSuite) SetupSuite() {
	ctp := m2cp_new.ContextPlus()
	t.firmwareDb = New(ctp, "../../testdata")
}

func (t *FirmwareDbTestSuite) TearDownTest() {
}

func TestFirmwareDbTestSuite(t *testing.T) {
	suite.Run(t, new(FirmwareDbTestSuite))
}

func (t *FirmwareDbTestSuite) TestFindLatestRevisionFoundUpdate() {
	sensor := &coap_server.Peer{
		FirmWareType:     42,
		HardWareRevision: 5,
		SequenceNumber:   1,
	}
	revision, err := t.firmwareDb.FindNewerRevision(sensor)
	t.NoError(err)
	t.NotNil(revision)
	t.Equal(42, revision.FWT)
	t.Equal(5, revision.HWR)
	t.Equal(3, revision.FWR)
}

func (t *FirmwareDbTestSuite) TestFindLatestRevisionNothingNew() {
	sensor := m2cp.CoapPeer(&coap_server.Peer{
		FirmWareType:     42,
		HardWareRevision: 5,
		SequenceNumber:   3,
	})
	revision, err := t.firmwareDb.FindNewerRevision(sensor)
	t.Error(err)
	t.Contains(err.Error(), "No compatible and allowed revision found")
	t.Nil(revision)
}

func (t *FirmwareDbTestSuite) TestIsCompatible() {
	// Your test logic here
	// Create a Revision object and use IsCompatible() method
	t.T().Skip("Not implemented")
}

func (t *FirmwareDbTestSuite) TestIsAllowed() {
	// Your test logic here
	// Create a Revision object and use IsAllowed() method
	t.T().Skip("Not implemented")
}

func (t *FirmwareDbTestSuite) TestIsIntegrityOk() {
	// Your test logic here
	// Create a Revision object and use IsHashesValid() method
	t.T().Skip("Not implemented")
}
