package tests

/* deprecated - we're using a test suite for this now
func InitUseUserConfig(t *testing.T) {
	var err error
	// defer goleak.VerifyNone(t)

	configFilename := "state.json"
	viper.SetConfigName(configFilename)
	viper.SetConfigType("json")
	home, err := os.UserHomeDir()
	assert.NoError(t, err)

	configDirName := fmt.Sprintf(".%s", "m2cp")
	configDirPath := filepath.Join(home, configDirName)
	viper.AddConfigPath(configDirPath)
	err = viper.ReadInConfig()
	assert.NoError(t, err)
}
*/
