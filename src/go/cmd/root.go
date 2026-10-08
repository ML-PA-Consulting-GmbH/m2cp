package cmd

import (
	"context"
	"errors"
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

func prepare(cmd *cobra.Command, _ []string) error {
	var err error
	flagStateJson := cmd.Flag("state").Value.String()
	state.ConfigStateFileName, err = tools.Abspath(flagStateJson)
	if err != nil {
		return err
	}
	err = initConfiguration()
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
	err = os.MkdirAll(dir, state.DirMode)
	cobra.CheckErr(err)

	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		// The state file holds a bearer JWT, so create it owner-only (0600).
		// os.Create would use 0644 (world-readable)
		// os.OpenFile applies the restrictive mode at creation, avoiding a create-then-chmod race.
		file, err2 := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, state.FileMode)
		cobra.CheckErr(err2)
		defer file.Close()
	}
}

func initConfiguration() error {
	// split filename into path, filename and extension
	configFileNameFull := filepath.Base(state.ConfigStateFileName)
	configFileExtension := filepath.Ext(configFileNameFull)
	if configFileExtension != ".json" {
		return errors.New("bad value for --state, configuration file must have '.json' extension")
	}
	configFileName := strings.TrimSuffix(configFileNameFull, configFileExtension)
	configFileDir := filepath.Dir(state.ConfigStateFileName)

	viper.SetConfigName(configFileName) // name of config file (without extension)
	viper.SetConfigType("json")         // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(configFileDir)

	// The state file holds a bearer JWT. Make viper write it owner-only (its default is 0644),
	// and tighten any pre-existing file that an older CLI created world-readable
	viper.SetConfigPermissions(state.FileMode)
	state.SecurePermissions(state.ConfigStateFileName)

	if err := viper.ReadInConfig(); err != nil {
		// Only create config file if it doesn't exist, not for other errors
		var notFoundErr viper.ConfigFileNotFoundError
		if errors.As(err, &notFoundErr) {
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
