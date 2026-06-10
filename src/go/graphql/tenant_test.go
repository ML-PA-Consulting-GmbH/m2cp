package graphql

//type TenantTestSuite struct {
//	suite.Suite
//	e                  *RuntimeEnvironment
//	url                string
//	userEmail          string
//	publicKey          []byte
//	privateKey         []byte
//	privateKeyPassword []byte
//	jwt                string
//}
//
//func (s *TenantTestSuite) SetupTest() {
//	var err error
//	s.e, err = NewRuntimeEnvironment("m2cp")
//	assert.NoError(s.T(), err)
//
//	//var err error
//	s.url = "https://app-gateway-svc-test-csharp.azurewebsites.net/graphql/"
//
//	s.userEmail = "testuser@ml-pa.com"
//
//	s.publicKey = []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDBmefSwoqMqG+9wLNq+O5nI/MP4Sjq/7ArGRlX/UlDG4kQSbaB9yF9OknPQYGVzDAOfuKfEITiYq2yyrvAvdyc/iAE/YHHb5LzRm1JeRGaE6m+RAha5rb7BCwSjGxg9wA+3VSTuRm7I31gpdgYA/n2G1yl2cew7XpP9LFT8XhJ5KBEaMqFBikjOxBTTrAj7fBG2NGid1oSC1GhpNWbiOvwKKUoB4gLj4vRmFXFeFjl+fWXX3Zqcr8X6ennTjPeaqHpVWJY5i3WE2YaJtfIOEum2gnSlmN4dnnNiu6DAfodV9k1NAC6ZqzZ/19+qBWlRrBjfWOaI5pRzvRfv0sKOhGzSlKZ2WEKg3ZXWbHOHtKyM5pUIzvLZ+A6GHCzBVV8qvKs/RVYE/VQzHFhGrVeKAGgj5vZWuOZ2UetarTnduF1I2Uqb2gjaukWFHeSee9d3DWrFO9fvJmPcNyjbDq6N5QXEeSz03W38JIoWOl40d84cSC0G5kVogKXhKj3eKvK/CM= flo@MLPA-NB105")
//	assert.Equal(s.T(), 567, len(s.publicKey))
//
//	s.privateKey = []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
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
//f6dVuIzZNzZNhG8aSPw6Cbfmn+c=
//-----END OPENSSH PRIVATE KEY-----`)
//	assert.Equal(s.T(), 2654, len(s.privateKey))
//
//	s.privateKeyPassword = []byte("test1234")
//
//	challenge, err := graphql.GetChallenge(s.url, s.userEmail, s.publicKey)
//	assert.NoError(s.T(), err)
//	assert.NotNil(s.T(), challenge)
//
//	signature, err := graphql.SignChallenge(challenge, s.privateKey, s.privateKeyPassword)
//	assert.NoError(s.T(), err)
//	assert.NotNil(s.T(), signature)
//
//	s.jwt, err = graphql.GetJSONWebToken(s.url, challenge, signature)
//	assert.NoError(s.T(), err)
//	assert.NotNil(s.T(), s.jwt)
//
//	err = s.e.SetGraphQlBackend(s.url, s.jwt)
//	assert.NoError(s.T(), err)
//}
//
//func (s *TenantTestSuite) TearDownTest() {
//	defer goleak.VerifyNone(s.T())
//
//	var err error
//	err = graphql.InvalidateJSONWebToken(s.url, s.jwt)
//	assert.NoError(s.T(), err)
//}
//
//func TestTenantTestSuite(t *testing.T) {
//	suite.Run(t, new(TenantTestSuite))
//}
//
//func (s *TenantTestSuite) TestTenantNameAndAliasByIdWithSuccess() {
//	tenantId := "226df810-b146-459e-bc64-defdfc5e6062"
//	tenantName, tenantAlias, err := TenantNameAndAlias(s.e, tenantId)
//	expectedName := "ML!PA Consulting GmbH"
//	expectedAlias := "mlpa"
//	assert.NoError(s.T(), err)
//	assert.Equal(s.T(), expectedName, tenantName)
//	assert.Equal(s.T(), expectedAlias, tenantAlias)
//}

//func TestTenantIdByNameWithTheOldBackend(t *testing.T) {
//	url := "https://app-snapstore-backend-st02-dev.azurewebsites.net/graphql"
//	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiU25hcFN0b3JlIiwia2V5IjoiQWRNWlRNY2lhOSIsImtleTIiOiJ3RkxJbFh2RDN5In0.LFi8sR6mheRTFeU6wk7ySJ0QDd0YKvJXcUuFdiXek6I"
//
//	var err error
//
//	httpClient := &http.Client{
//		Transport: &transport{
//			headers: map[string]string{
//				"Token": token,
//			},
//		},
//	}
//	graphqlClient := graphql.NewClient(url, httpClient)
//
//	var query struct {
//		Tenants []struct {
//			Id string
//		} `graphql:"brands(where: {brandName: {eq: $name}})"`
//	}
//	name := "test"
//	variables := map[string]any{
//		"name": string(name),
//	}
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	err = graphqlClient.Query(ctx, &query, variables)
//	assert.NoError(t, err)
//
//	actualId := string(query.Tenants[0].Id)
//	expectedId := "7275f4de-6136-4aa2-9fcb-67d0174b7c8b"
//	assert.Equal(t, expectedId, actualId)
//}
