package cmd

import (
	"context"
	"fmt"
	"m2cpcli/format"
	"m2cpcli/state"
	"m2cpcli/tools"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

//var cfgFile string
//var jwt string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:               "m2cp",
	Short:             "M2CP command line tool for developers",
	PersistentPreRunE: prepare,
	SilenceErrors:     true, // Assure, nothing is printed to stdout when we expect JSON!
}

func prepare(cmd *cobra.Command, args []string) error {
	err := initConfiguration(cmd)
	if err != nil {
		return err
	}
	cmd.SilenceUsage = true
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute(ctx context.Context) {
	err := RootCmd.ExecuteContext(ctx)
	if err != nil {
		type ErrorMessage struct {
			Error string `json:"error" yaml:"error"`
		}
		var errorMessage ErrorMessage
		errorMessage.Error = err.Error()

		formatter := func(msg ErrorMessage) (string, error) {
			res := fmt.Sprintf("Error: %s", msg.Error)
			return res, nil
		}

		msgErr := format.PrintFormattedOutput(RootCmd, errorMessage, formatter)
		if msgErr != nil {
			RootCmd.PrintErrln(msgErr.Error())
		}
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&format.JsonOutputMode, "json", false, "print output as JSON")

	state.FixLegacyStateFile()

	RootCmd.PersistentFlags().String("state", state.DefaultStatePath, "experts: provide alternative path for state.json file")
	_ = RootCmd.PersistentFlags().MarkHidden("state")

	RootCmd.CompletionOptions.DisableDefaultCmd = true // disable the "completion" feature
}

func createFile(filePath string) {
	var err error
	dir := filepath.Dir(filePath)
	err = os.MkdirAll(dir, 0755)
	cobra.CheckErr(err)

	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		file, err2 := os.Create(filePath)
		cobra.CheckErr(err2)
		defer file.Close()
	}
}

func initConfiguration(cmd *cobra.Command) error {
	var err error
	flagStateJson := cmd.Flag("state").Value.String()
	state.ConfigStateFileName, err = tools.Abspath(flagStateJson)

	// split filename into path, filename and extension
	configFileNameFull := filepath.Base(state.ConfigStateFileName)
	configFileExtension := filepath.Ext(configFileNameFull)
	if configFileExtension != ".json" {
		return fmt.Errorf("bad value for --state, configuration file must have '.json' extension")
	}
	configFileName := strings.TrimSuffix(configFileNameFull, configFileExtension)
	configFileDir := filepath.Dir(state.ConfigStateFileName)

	viper.SetConfigName(configFileName) // name of config file (without extension)
	viper.SetConfigType("json")         // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(configFileDir)

	if err = viper.ReadInConfig(); err != nil {
		// Only create config file if it doesn't exist, not for other errors
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			createFile(state.ConfigStateFileName)
			if err := viper.WriteConfig(); err != nil {
				// If we still can't write config, it's not critical for version command
				// Just continue without config
				return nil
			}
		} else {
			// For other errors (permissions, parse errors, etc.), fail
			cobra.CheckErr(err)
		}
	}
	return nil
}
