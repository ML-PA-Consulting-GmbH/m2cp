package graphql

//
//func (s *TestSuite) TestSnapDeclarationByDatabaseId() {
//	var err error
//	// defer goleak.VerifyNone(t)
//
//	configFilename := "state.json"
//	viper.SetConfigName(configFilename)
//	viper.SetConfigType("json")
//	home, err := os.UserHomeDir()
//	s.NoError(err)
//
//	configDirName := fmt.Sprintf(".%s", "m2cp")
//	configDirPath := filepath.Join(home, configDirName)
//	viper.AddConfigPath(configDirPath)
//	err = viper.ReadInConfig()
//	s.NoError(err)
//
//	// The code above is needed to initialize the JWT for the request
//
//	ctx := context.Background()
//	snapDatabaseId := UUID("c0a23730-040d-4b98-659d-08db2482b564") // rabbitmq-server-snap
//	decl, err := SnapDeclarationByDatabaseId(ctx, snapDatabaseId, true, true)
//	s.NoError(err)
//	s.NotNil(decl)
//
//	// TODO: proper assertions
//	// fmt.Printf("%v\n", decl)
//}
