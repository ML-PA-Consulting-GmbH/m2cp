package graphql

//func (s *TestSuite) TestStoreAssertions() {
//	assertions, err := StoreAssertions(s.ctx)
//	s.NoError(err)
//	s.NotNil(assertions)
//	s.NotEmpty(assertions)
//}

//func (s *TestSuite) TestSnapAssertionsBySnapDatabaseIdAndRevision() {
//	s.T().Skip("FIXME!")
//	snapDatabaseId := UUID("c0a23730-040d-4b98-659d-08db2482b564") // rabbitmq-server-snap
//	revision := int32(1)
//	assertions, err := SnapAssertionsBySnapDatabaseIdAndRevision(s.ctx, snapDatabaseId, revision)
//	s.NoError(err)
//	s.NotNil(assertions)
//	s.NotEmpty(assertions)
//}

//func (s *TestSuite) TestSnapAssertionsBySnapDatabaseIdAndRevisionsInvalidRevision() {
//	s.T().Skip("FIXME!")
//	snapDatabaseId := UUID("c0a23730-040d-4b98-659d-08db2482b564") // rabbitmq-server-snap
//	revision := int32(999)
//	assertions, err := SnapAssertionsBySnapDatabaseIdAndRevision(s.ctx, snapDatabaseId, revision)
//	s.Error(err)
//	s.Contains(err.Error(), "revision")
//	s.Nil(assertions)
//}

//func (s *TestSuite) TestSnapDeclarationByDatabaseIdAndRevisionsInvalidSnapDbId() {
//	snapDatabaseId := UUID("c0a23730-040d-4b98-659d-08db2482b5ff") // rabbitmq-server-snap
//	revision := int32(999)
//	assertions, err := SnapAssertionsBySnapDatabaseIdAndRevision(s.ctx, snapDatabaseId, revision)
//	s.Error(err)
//	s.Contains(err.Error(), "declaration")
//	s.Nil(assertions)
//}
