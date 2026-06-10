package env

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"testing"
)

func TestJsonWebToken_IsValid(t *testing.T) {
	var err error
	defer goleak.VerifyNone(t)

	str := `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJkaXNwbGF5X25hbWUiOiJKYWtvYiBEdWViZWwiLCJlbWFpbCI6Impha29iLmR1ZWJlbEBtbC1wYS5jb20iLCJzY29wZXMiOiIqLiouQWxsIiwic2Vzc2lvbl9pZCI6IjY0ZmVhZTg1LWZiY2ItNDE0NS1iYWMzLTE0Mzg3ZWIzNDdkYSIsInRlbmFudF9pZCI6IjIyNmRmODEwLWIxNDYtNDU5ZS1iYzY0LWRlZmRmYzVlNjA2MiIsInVzZXJfaWQiOiJlZmEzOGNmMC1jOTJlLTQzMzktYWNjMC1hY2YxMDE1YWI5YWEiLCJuYmYiOjE2OTM4NDMyNjgsImV4cCI6MTY5Mzg4NjQ2OCwiaWF0IjoxNjkzODQzMjY4LCJpc3MiOiJodHRwczovL3d3dy5tbC1wYS5jb20vIn0.fREXUAR7N-RnHBQARy56pGGfVGpQEgADFR0J2nVk46Ux-99s4kotNXU4WqnE8fx-6YhGlOoBRG0fmtmd5ObS4lH1LiGi5PT4Qsh4j6zdELp3rPkvJvBhMCd4JUMXmW3BEES1QM_-HCKtmgF2O58MdNHjq3vXLfl9rGSuTeFP-QJJtfd54kx6MDZdyC8LXz5iIPXclA7FACDgZwD2T9IV5EUoHK9cxGSdwo-6cVgQQg_Xnw_l0qKceNlzK9fXGKtcK2qA-9zeNuxPOOS01KgwRlYgLKxgdqCStcG1e_6OSRln2nRUgWpXrqc2ob5J1blRKe2wnJH7VcLKhpdLKu98zg`
	jwt, err := NewJsonWebToken(str)
	assert.NoError(t, err)
	assert.NotNil(t, jwt)

	assert.False(t, jwt.IsValid())
}

func TestJsonWebToken_EmptyStringIsValid(t *testing.T) {
	var err error
	defer goleak.VerifyNone(t)

	str := ""
	jwt, err := NewJsonWebToken(str)
	assert.NoError(t, err)
	assert.Nil(t, jwt)
	assert.False(t, jwt.IsValid())
}
