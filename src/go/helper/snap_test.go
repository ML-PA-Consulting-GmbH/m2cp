package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestRetrieveSnapFilenameFromUrl_ValidWithUpstream(t *testing.T) {
	defer goleak.VerifyNone(t)

	url := "https://fake.blob.core.windows.net/snap-revisions/mlpa_m2cp-os-imx8mp-gadget.5.2.11-dev+up_50_arm64.snap?sv=2015-01-01&sr=b&sig=fo%&st=2026-01-20T14%3A25%3A48Z&se=2026-01-20T15%3A25%3A48Z&sp=r"
	filename, err := RetrieveSnapFilenameFromUrl(url)
	assert.NoError(t, err)
	assert.Equal(t, "mlpa_m2cp-os-imx8mp-gadget.5.2.11-dev+up_50_arm64.snap", filename)
}

func TestRetrieveSnapFilenameFromUrl_ValidWithoutUpstream(t *testing.T) {
	defer goleak.VerifyNone(t)

	url := "https://fake.blob.core.windows.net/snap-revisions/mlpa_m2cp-gateway.0.16.2-dev_1_arm64.snap?sv=2018-03-28&sr=b&sig=bla%3D&st=2026-01-20T14%3A49%3A09Z&se=2026-01-20T15%3A49%3A09Z&sp=r"
	filename, err := RetrieveSnapFilenameFromUrl(url)
	assert.NoError(t, err)
	assert.Equal(t, "mlpa_m2cp-gateway.0.16.2-dev_1_arm64.snap", filename)
}

func TestRetrieveSnapFilenameFromUrl_Invalid(t *testing.T) {
	defer goleak.VerifyNone(t)

	url := "http://localhost"
	filename, err := RetrieveSnapFilenameFromUrl(url)
	assert.Error(t, err)
	assert.Empty(t, filename)
	assert.Equal(t, "failed to retrieve snap filename from URL \"http://localhost\"", err.Error())
}

func TestRetrieveSnapFilenameFromUrl_Empty(t *testing.T) {
	defer goleak.VerifyNone(t)

	url := ""
	filename, err := RetrieveSnapFilenameFromUrl(url)
	assert.Error(t, err)
	assert.Empty(t, filename)
	assert.Equal(t, "failed to retrieve snap filename from URL \"\"", err.Error())
}
