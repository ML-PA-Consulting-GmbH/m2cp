package auth

import (
	"context"
	"encoding/hex"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
	"testing"
)

type ChallengeTestSuite struct {
	suite.Suite
	ctx       context.Context
	ctxCancel context.CancelFunc

	// for invalidate json web token tests
	url                string
	userEmail          string
	publicKey          []byte
	privateKey         []byte
	privateKeyPassword []byte
}

// TODO: how shall this work? ~/.m2cp/config must be a JSON to match the ChallengeTestSuite struct? But why is it completely overwritten?
func (s *ChallengeTestSuite) SetupSuite() {
	s.url = "https://app-as3-int-gateway-svc-dev-euw.azurewebsites.net/graphql"

	s.userEmail = "test@ml-pa.test.com"

	s.publicKey = []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC9IGVL8in3V27nx2E0QmLFffn1xhxGC0B8pC7cVvyRSuPHG4SPJxwqYk85VV/yfEz2bmSd7hcb2pGxaEBTYYF0gUNx3/nX7dbgd0bGMsQljQymySWs/XPwLNyVj7o5MvZ0R69UkRLUFMVixKEVeOpPU9gjMStiSC+Mk166ymGluxQhs0W++Mm2Hk9u8/2Vs8Ff6yJFMp+xOnxZR6pWjvIBNiCKXPrUgvvOVwg5rK2b+d3U892I/JjA2aaIvYWMiQWFIpwuqgGh2FzWpQXTqL9Ta6GvH2+wADLOjaPKvmRXTyz9kSKEvQ7E8RGDN9YCsE4bI6NAIyAOGjKiLw7Uw4Jx0QOBf7zZWuJE3h+xVa/Vy6JLD3AazEuCM2WbERWq1A5XXqw2g34s++y3rpOosJXgC9ONNRfWBko7y1kr2EpkEkQLJr8Pue8wfIcF4lVyc5325kcidvi3XNNGWlpOFFsxXuhc3E++jxYQiAyU5TaUBChMnpeVK0YoQXvWqq0y64E= flo@MLPA-NB105")
	s.Equal(567, len(s.publicKey))

	// TODO: this is not matching the public key anymore!
	s.privateKey = []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
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
-----END OPENSSH PRIVATE KEY-----`)
	s.Equal(2601, len(s.privateKey))

	s.privateKeyPassword = []byte("test1234")
}

func (s *ChallengeTestSuite) TearDownSuite() {
}

func (s *ChallengeTestSuite) SetupTest() {
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
}

func (s *ChallengeTestSuite) TearDownTest() {
	s.ctxCancel()
	<-s.ctx.Done()

	goleak.VerifyNone(s.T())
}

func TestSuiteRunner(t *testing.T) {
	suite.Run(t, new(ChallengeTestSuite))
}

func (s *ChallengeTestSuite) TestGetChallengeForTestuserWithRSAKeyNoPassword() {
	challenge, err := GetChallenge(s.ctx, s.url, s.userEmail, s.publicKey)
	s.NoError(err)
	s.Equal(48, len(challenge))
}

//func (s *ChallengeTestSuite) TestGetChallengeFromWrongURL() {
//	var err error
//
//	url := "https://example.com/wrong/graphql/" // wrong URL
//
//	challenge, err := GetChallenge(s.ctx, url, s.userEmail, s.publicKey)
//	s.Error(err)
//	s.Nil(challenge)
//	fmt.Println(err.Error())
//	//assert.True(t, strings.HasPrefix(err.Error(), "on mutation: non-200 OK status code: 404 Not Found"))
//	//s.Equal("decoding response: invalid character '<' looking for beginning of value", err.Error()) // TODO: why?
//}

func (s *ChallengeTestSuite) TestGetChallengeForInvalidUser() {
	var err error
	userEmail := "wronguser@ml-pa.com" // invalid user

	challenge, err := GetChallenge(s.ctx, s.url, userEmail, s.publicKey)
	s.Error(err)
	s.Nil(challenge)
	s.Equal("input: The requested 'User' could not be found.\n", err.Error())
}

func (s *ChallengeTestSuite) TestGetChallengeWithBadPublicKey() {
	var err error

	sshPublicKey := []byte("ssh-rsa BAD= user@host")

	challenge, err := GetChallenge(s.ctx, s.url, s.userEmail, sshPublicKey)
	s.Error(err)
	s.Nil(challenge)
	s.Equal("input: The provided SSH public key does not match. Please verify your key and try again.\n", err.Error())
}

func (s *ChallengeTestSuite) TestGetChallengeForEmptyEmail() {
	var err error

	challenge, err := GetChallenge(s.ctx, s.url, "", s.publicKey)
	s.Error(err)
	s.Nil(challenge)
	s.Equal("input: User email is missing. Please provide a valid email address.\n", err.Error())
}

func (s *ChallengeTestSuite) TestGetChallengeForInvalidEmail() {
	var err error

	invalidUserEmail := "ml-pa.com"
	sshPublicKey := "ssh-rsa BAD= user@host"

	challenge, err := GetChallenge(s.ctx, s.url, invalidUserEmail, []byte(sshPublicKey))
	s.Error(err)
	s.Nil(challenge)
	s.Equal("input: The requested 'User' could not be found.\n", err.Error())
}

func (s *ChallengeTestSuite) TestGetChallengeWithoutPublicKey() {
	var err error

	invalidUserEmail := "testuser@ml-pa.com"

	challenge, err := GetChallenge(s.ctx, s.url, invalidUserEmail, nil)
	s.Error(err)
	s.Nil(challenge)
	s.Equal("input: SSH public key is missing. Please provide a valid SSH public key.\n", err.Error())
}

func (s *ChallengeTestSuite) TestSignChallengeForRSAWithPassword() {
	const privKeyPem = `-----BEGIN OPENSSH PRIVATE KEY-----
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
f6dVuIzZNzZNhG8aSPw6Cbfmn+c=
-----END OPENSSH PRIVATE KEY-----`

	const privKeyPemPassword = "test1234"

	const nonceReceivedFromServer = "4beb89062bb2d09c01345ea58fa999e0f63773bbb0edc23bc726c638e3fe45889b992184face1a3390c3f44ec018df7a"

	nonceBytes, err := hex.DecodeString(nonceReceivedFromServer)
	s.NoError(err)

	signature, err := SignChallenge(nonceBytes, []byte(privKeyPem), []byte(privKeyPemPassword))
	s.NoError(err)
	s.NotNil(signature)

	signatureEncoded := hex.EncodeToString(signature)
	s.Equal(768, len(signatureEncoded))

	const expectedSignatureEncoded = "40d2a00c5752f04b1fda763262026eb75b146c430ee722bca913bec8b5d7badd3c3afed0e927a2a232b7df7841071f60ca492eb23c51d4805aceba4e3471fb502306f14fc6dcc6b6e771bd0799a2fc3138510a5e0b2b4758178ae529a1f382b866021cba23aa3f9ab97e46ec6c3b3b6dbf6af0ee61314006d8b3a6e4b15f5a23d31efed2de6399f54711a529347a05b21785142ea6297ef6eed7940c6a655a3650d056233d9c7104a43f42c34e911e914778c4c09f31c78d72402b11e436f73d2372a5fc625aa807288dff76cf04cd5d0a665f0511446cf8aae7e6bff798fb5dd2286dd0a2277b4d79a5a9d082d1ba46c9dff5ccb8b2bcda41c03c50ad6516e3ad658ef0dc0b894aab7b73aedf955780be80fd10749b5485915b29d222e64cd10739352df45c841a89118e0e19d7803ff093fbcab2a099c16684b56832b24ed89cc08d9ffb7229dc9a9561d5a1933589396559a1690c00287df94197915ab0a407f057d039bc7e42b37ec669cce86b43c19e822fd773ba1b6949b5b7af5f2c16"
	s.Equal(expectedSignatureEncoded, signatureEncoded)
}

func (s *ChallengeTestSuite) TestSignChallengeForED25519WithoutPassword() {
	// ED25519 testkey
	const privKeyPem = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACB1xjmiURQOOAbIsK49Am32SLCRR0nqgZd61i0F6V4dWwAAALAnA27UJwNu
1AAAAAtzc2gtZWQyNTUxOQAAACB1xjmiURQOOAbIsK49Am32SLCRR0nqgZd61i0F6V4dWw
AAAEDUgwg5LH1UTVGne28eSvO9qygdFYG3EzSqh+WelkEofnXGOaJRFA44Bsiwrj0CbfZI
sJFHSeqBl3rWLQXpXh1bAAAAKGZsb0A5YzEyYzAzNy1jMzc5LTQzYzQtYmFkOS0zZGQ5OT
JjZDlhZGUBAgMEBQ==
-----END OPENSSH PRIVATE KEY-----`

	const nonceReceivedFromServer = "4beb89062bb2d09c01345ea58fa999e0f63773bbb0edc23bc726c638e3fe45889b992184face1a3390c3f44ec018df7a"

	nonceBytes, err := hex.DecodeString(nonceReceivedFromServer)
	s.NoError(err)

	signature, err := SignChallenge(nonceBytes, []byte(privKeyPem), nil)
	s.NoError(err)
	s.NotNil(signature)

	signatureEncoded := hex.EncodeToString(signature)
	s.Equal(128, len(signatureEncoded))

	// TODO: this value is just copied, not verified to be correct
	const expectedSignatureEncoded = "1f8c939fa21dbedbb1130650dce53c688b1abaa60fa46f09225bda041155d969a5c73f89de3b93caa272168f1482b89df9c1c655aefc11f69bacae25588e5d01"
	s.Equal(expectedSignatureEncoded, signatureEncoded)
}

func (s *ChallengeTestSuite) TestSignChallengeForRSAWithoutPassword() {
	const privKeyPem = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`

	const nonceReceivedFromServer = "4beb89062bb2d09c01345ea58fa999e0f63773bbb0edc23bc726c638e3fe45889b992184face1a3390c3f44ec018df7a"

	nonceBytes, err := hex.DecodeString(nonceReceivedFromServer)
	s.NoError(err)

	signature, err := SignChallenge(nonceBytes, []byte(privKeyPem), nil)
	s.NoError(err)
	s.NotNil(signature)

	signatureEncoded := hex.EncodeToString(signature)
	s.Equal(768, len(signatureEncoded))

	// TODO: this value is just copied, not verified to be correct
	// TODO: It is the same as with the private key with password protection above?!
	const expectedSignatureEncoded = "40d2a00c5752f04b1fda763262026eb75b146c430ee722bca913bec8b5d7badd3c3afed0e927a2a232b7df7841071f60ca492eb23c51d4805aceba4e3471fb502306f14fc6dcc6b6e771bd0799a2fc3138510a5e0b2b4758178ae529a1f382b866021cba23aa3f9ab97e46ec6c3b3b6dbf6af0ee61314006d8b3a6e4b15f5a23d31efed2de6399f54711a529347a05b21785142ea6297ef6eed7940c6a655a3650d056233d9c7104a43f42c34e911e914778c4c09f31c78d72402b11e436f73d2372a5fc625aa807288dff76cf04cd5d0a665f0511446cf8aae7e6bff798fb5dd2286dd0a2277b4d79a5a9d082d1ba46c9dff5ccb8b2bcda41c03c50ad6516e3ad658ef0dc0b894aab7b73aedf955780be80fd10749b5485915b29d222e64cd10739352df45c841a89118e0e19d7803ff093fbcab2a099c16684b56832b24ed89cc08d9ffb7229dc9a9561d5a1933589396559a1690c00287df94197915ab0a407f057d039bc7e42b37ec669cce86b43c19e822fd773ba1b6949b5b7af5f2c16"
	s.Equal(expectedSignatureEncoded, signatureEncoded)
}

func (s *ChallengeTestSuite) TestSignChallengeForRSAWithPasswordWithSecondNounce() {
	const privKeyPem = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAACmFlczI1Ni1jdHIAAAAGYmNyeXB0AAAAGAAAABBb6hx1vi
17xYPK1f7n7gwdAAAAEAAAAAEAAAGXAAAAB3NzaC1yc2EAAAADAQABAAABgQDBmefSwoqM
qG+9wLNq+O5nI/MP4Sjq/7ArGRlX/UlDG4kQSbaB9yF9OknPQYGVzDAOfuKfEITiYq2yyr
vAvdyc/iAE/YHHb5LzRm1JeRGaE6m+RAha5rb7BCwSjGxg9wA+3VSTuRm7I31gpdgYA/n2
G1yl2cew7XpP9LFT8XhJ5KBEaMqFBikjOxBTTrAj7fBG2NGid1oSC1GhpNWbiOvwKKUoB4
gLj4vRmFXFeFjl+fWXX3Zqcr8X6ennTjPeaqHpVWJY5i3WE2YaJtfIOEum2gnSlmN4dnnN
iu6DAfodV9k1NAC6ZqzZ/19+qBWlRrBjfWOaI5pRzvRfv0sKOhGzSlKZ2WEKg3ZXWbHOHt
KyM5pUIzvLZ+A6GHCzBVV8qvKs/RVYE/VQzHFhGrVeKAGgj5vZWuOZ2UetarTnduF1I2Uq
b2gjaukWFHeSee9d3DWrFO9fvJmPcNyjbDq6N5QXEeSz03W38JIoWOl40d84cSC0G5kVog
KXhKj3eKvK/CMAAAWQ3rYfwL2pmyButK5zsAGPwnNjKwB1tFoziZquZY8MnLNn9lY4j0Z9
mk7b6ZXVzpLq9wfcMrPZE2Odo/WFz7zfvy0JDiZr1G2lYi9eTtMiY+oJVfB+HV0yHANnAH
wkW1orEzkyYuRDHylcaXdOr1mnKcXZrY+zeHT0Vd/aH5jtXudicMZnzIoT/kWxfXH7Lo+O
CgGauOAavnOAb9J20sjRyAWLDGtWJMNy+NEXfhLPJmyr6Sju9D5dsrMQdSaBqakSnno6LR
wFvkTcE9DW87AnPkQ6MqD4T8cloh/P1mtS2hFv4ZOjicrMWp/2V6kfmrAJPNByizdeyNbc
TmF64njSArfnThCUv/IxturS+ESG9YexzrlqvsrTE47s+JH8DwQuJqZIZ3lXSiG7mgdOG0
inCTaQLjGr/QbgSO4QtFQe1N71KSZ2uaskXeyWg30hKv0GdgC/W85oH0/DVok4W32tGahn
kqAaO09McvmfljUvyuSxtyQYezEdir6u9lFhRUpTOTNyVffjU4fW2XGQsxTIhuKB7fPyR5
aF5r7RSUcCQASWa72mSvJkx7FaOk0kr76TkpfR90944+qoc1DwoVLbjBwpjmox78Yrf287
qgB1P2UVWkdSBqKOtoZ1jvFI783hPJHlEzfn0QQY8NR183KSG/kfR1LZKRILYneyRDDhtv
7Y49PYPttctTDy9nyl2Hb4Bt2wz/7iJZZ4FjQ80gKXQrVRS1bCLcOk8eQpZONRI59tyTfR
M7R3SOidE9/pCb1sH03QxpWFk2sl3AxrckQgNxSLFjFP46YKcxyUF7q7KnlWuHTQ6lF8tZ
/BJNFLDAy02X3ogloTtgvWBti0hd51rb7TYlsxdIr2j7fcxuatF59IJ3hRBLbeLdYtJxhr
45DNJrvx01LGGVAfcjsS8X/oXkq+afN9vgCfpyDIDNn8ekA1Aiq/ENmAu48DRdV79YECCG
km/f+7eWlUxBHMRwM7O7tqnTIKY0POvV23uC2s8EFKk2ue47XSEJqf4Mo9vmPAB9YYgvO5
KMpNjd0r8e0g7/wfWrum2CEvgGeWpmxPvDuSozIUGE8ym2IWn3//0YlSHcQnGofzRGbtEO
uMj1JOGIDkrEfOx5gNYPbjYa+E9/sSyIshLt/U04tir9uNSY1nwuO0EnzoomhHUZYzPDFj
jpZTecIeJRAadT56bQ/lZ9u6GZe825RpmIsxemwc90nFedFrfo2cXoKvGEv9lB/w3rfEmE
BqaOYFzNV4bcwakV1il+Y5dFThBgRUnCvRqJD8Iq5+uJVQLY0AzS5AkA6LPwEqA+TM5Wyd
uH1oYMeoBUVTod+x+2p1c+GWcsihvd/nJ1JY/EGsDF9L9hfDRMsirZss3oYcnxya58va4E
kl66bt//DJDhyCrFZAkE772GeZdd6QtWSwl14BD/sxmTQQz94skFOUBqAj+NZs3Xk+NtIm
k7cpuMpaL2ir9QdL+FTv6BNSnV+jU9ZWSmW8KAtXr4KOVthcUTbLYvQWW6LdF5+hSm/pfy
7UAYNF8DoI30STs0omblCWPnMw6SoorXdl4eGzhY+aocOmlu2tDfI3gq+3KfgllVwAsRLg
eg+saPHkYMQ6jeexro9XSaTBkpylzfAGXjau9GQr1fszN8yvkBqiSA98oP2Idv4k3+YBjN
J7sn/Y++8+yjCbDf62fEZG89x0WerooUSzP66IRyAdARLD03869EL6wW/cMd4azr6I0z+N
6OKQuo94vZqt6dJRm9ILcyjMNxxSV36dTsWoPkdUZ78ZVFsAkGeKxbnxb7jm6MY+6JnkHX
rdlB5egUQ/URxKNf1OwjaEojw6YwwVsy+cSywApo/WTIY3LN68Ut2hZE0D6yD7UvIOKKfU
f6dVuIzZNzZNhG8aSPw6Cbfmn+c=
-----END OPENSSH PRIVATE KEY-----`

	const privKeyPemPassword = "test1234"

	const nonceReceivedFromServer = "9f8cf40d89d02e23040bb3890aa9042ee961cddaf3fa9c794686084ddfa13f062360347a7e908f4d606ee11ac6259e78"

	nonceBytes, err := hex.DecodeString(nonceReceivedFromServer)
	s.NoError(err)

	signature, err := SignChallenge(nonceBytes, []byte(privKeyPem), []byte(privKeyPemPassword))
	s.NoError(err)
	s.NotNil(signature)

	signatureEncoded := hex.EncodeToString(signature)
	s.Equal(768, len(signatureEncoded))

	const expectedSignatureEncoded = "7c5b80b860026c6d2ba561792c91e5cabd6828d4ab4f4b5fe3bf5880227f1527591ab57e786e24cc61e79442d639d0dcb70be5d305ba80ad373e774491815555cf696f2a40b66f54dba5087f0fc8f695caf000519fd0a90086485e29c6eb12f55b3be4f3da8bb8a0b6eee315765aa4b6b3f1c2c625b537435a11b43ab8d5c5efe8eb96d24bbbe8caeb9c97ea70a7446ba9de5ceea59871a33bc4b4f96c66dd10d406633ac8b8f2d50baa7c25fa6556c4da9713600ddfd460036d9d2af1056481014874cb9e9e0075fdd00ebfdcb3987794e7b5e29d7c281c6edb8371c5a85e43b1899361572361629487c8d2bc5bb1405ea532e4bff6bc93988015c9728ac9c8672ec0ccb332e4ff86f2c5c0648a8efb5b739fbdf5a78a58a6f64f04f84943a9def450decba1a410a9e42eff70b06233ae09bb1ad0331af3ddacd2790f4528f1c53b9ca106c0936b9ed9d2084556093e9919f7110fcafdf9e2a4396e78d52ac99d94eb9831ce8a536f9da06bd29f5c83dac17cb015322a99c34d3251bfa4df72"
	s.Equal(expectedSignatureEncoded, signatureEncoded)
}

func (s *ChallengeTestSuite) TestSignChallengeForFakePrivateKey() {
	const privKeyPem = `-----BEGIN OPENSSH PRIVATE KEY-----

thisisjustabadfake==
-----END OPENSSH PRIVATE KEY-----`

	const nonceReceivedFromServer = "4beb89062bb2d09c01345ea58fa999e0f63773bbb0edc23bc726c638e3fe45889b992184face1a3390c3f44ec018df7a"

	nonceBytes, err := hex.DecodeString(nonceReceivedFromServer)
	s.NoError(err)

	signature, err := SignChallenge(nonceBytes, []byte(privKeyPem), nil)
	s.Error(err)
	s.Nil(signature)
	s.Equal("could not parse private key: ssh: invalid openssh private key format", err.Error())
}

func (s *ChallengeTestSuite) TestGetJSONWebTokenWithSuccess() {
	var err error

	challenge, err := GetChallenge(s.ctx, s.url, s.userEmail, s.publicKey)
	s.NoError(err)
	s.NotNil(challenge)

	//fmt.Println(challenge)
	//challenge, err := hex.DecodeString("9f8cf40d89d02e23040bb3890aa9042ee961cddaf3fa9c794686084ddfa13f062360347a7e908f4d606ee11ac6259e78")
	//s.NoError(err)

	//signature, err := SignChallenge(challenge, s.privateKey, s.privateKeyPassword)
	signature, err := SignChallenge(challenge, s.privateKey, nil)
	s.NoError(err)
	//fmt.Println(signature)

	jwt, err := GetJSONWebToken(s.ctx, s.url, challenge, signature)
	s.NoError(err)
	s.NotNil(jwt)
	s.NotEmpty(jwt)
	// s.Equal(3502, len(jwt)) // TODO: this is a bad test. The length changes, e.g. on user permission change
	// fmt.Println(jwt)
}

// TODO: key not in the database?
//func (s *TestSuite) TestGetJSONWebToken02(t *testing.T) {
//	const sshPublicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHXGOaJRFA44Bsiwrj0CbfZIsJFHSeqBl3rWLQXpXh1b flo@9c12c037-c379-43c4-bad9-3dd992cd9ade"
//	s.Equal(t, 121, len(sshPublicKey))
//
//	var err error
//	url := "https://app-gateway-svc-test-csharp.azurewebsites.net/graphql/"
//	userEmail := "testuser@ml-pa.com"
//
//	challenge, err := GetChallenge(url, userEmail, []byte(sshPublicKey))
//	s.NoError(t, err)
//	s.NotNil(t, challenge)
//
//	const privKeyPem = `-----BEGIN OPENSSH PRIVATE KEY-----
//b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
//QyNTUxOQAAACB1xjmiURQOOAbIsK49Am32SLCRR0nqgZd61i0F6V4dWwAAALAnA27UJwNu
//1AAAAAtzc2gtZWQyNTUxOQAAACB1xjmiURQOOAbIsK49Am32SLCRR0nqgZd61i0F6V4dWw
//AAAEDUgwg5LH1UTVGne28eSvO9qygdFYG3EzSqh+WelkEofnXGOaJRFA44Bsiwrj0CbfZI
//sJFHSeqBl3rWLQXpXh1bAAAAKGZsb0A5YzEyYzAzNy1jMzc5LTQzYzQtYmFkOS0zZGQ5OT
//JjZDlhZGUBAgMEBQ==
//-----END OPENSSH PRIVATE KEY-----`
//
//	signature, err := SignChallenge(challenge, []byte(privKeyPem), nil)
//	s.NoError(t, err)
//	s.NotNil(t, signature)
//
//	jwt, err := GetJSONWebToken(url, challenge, signature)
//	s.NoError(t, err)
//	s.NotNil(t, jwt)
//}

func (s *ChallengeTestSuite) TestGetJSONWebTokenForMissingChallenge() {
	var err error
	signature := []byte("test")
	jwt, err := GetJSONWebToken(s.ctx, s.url, nil, signature)
	s.Error(err)
	s.Equal("", jwt)
	s.Equal("input: Challenge is missing. Please provide a valid challenge.\n", err.Error())
}

func (s *ChallengeTestSuite) TestGetJSONWebTokenForInvalidChallenge() {
	var err error
	challenge := []byte("invalid")
	jwt, err := GetJSONWebToken(s.ctx, s.url, challenge, nil)
	s.Error(err)
	s.Equal("", jwt)
	s.Equal("input: Signed challenge is missing. Please provide a valid response.\n", err.Error())
}

func (s *ChallengeTestSuite) TestGetJSONWebTokenForMissingSignedChallenge() {
	var err error

	challenge, err := hex.DecodeString("9f8cf40d89d02e23040bb3890aa9042ee961cddaf3fa9c794686084ddfa13f062360347a7e908f4d606ee11ac6259e78")
	s.NoError(err)

	jwt, err := GetJSONWebToken(s.ctx, s.url, challenge, nil)
	s.Error(err)
	s.Equal("", jwt)
	s.Equal("input: Signed challenge is missing. Please provide a valid response.\n", err.Error())
}
