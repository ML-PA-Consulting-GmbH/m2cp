package graphql

//func (s *TestSuite) TestEdgeDeviceInstallationStates() {
//	var err error
//	s.T().Skip("FIXME!")
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
//	ctx := context.Background()
//	deviceId := UUID("0c5b3881-3eaa-416d-dd32-08dbe508f6ce")
//	edis, err := EdgeDeviceInstallationStates(ctx, deviceId)
//	s.NoError(err)
//	s.NotNil(edis)
//
//	// TODO: proper assertions
//	//for _, e := range edis {
//	//	fmt.Printf("%v\n", e.SnapRevision.SnapDeclarationId)
//	//
//	//	decl, err := SnapDeclarationByDatabaseId(ctx, e.SnapRevision.SnapDeclarationId, true, true)
//	//	s.NoError(err)
//	//
//	//	fmt.Println(decl.SnapName)
//	//}
//}
