package system

import (
	"context"
	_ "embed"
	"fmt"
	"m2cpcli/backend"
	gql "m2cpcli/graphql"
	"m2cpcli/helper"
	"m2cpcli/structs"
	"m2cpcli/tools"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var generateConstantsGoCmd = &cobra.Command{
	Use:   "generate-constants-go [modelName revision]",
	Short: "Generate Go constants",
	Args:  cobra.ExactArgs(2),
	RunE:  runGenerateConstantsGoCmd,
}

//go:embed template/constants.go.template
var constantsGoTemplate string

func init() {
	SystemCmd.AddCommand(generateConstantsGoCmd)
}

func runGenerateConstantsGoCmd(cmd *cobra.Command, args []string) error {
	modelName := args[0]
	if args[1] == "latest" {
		return fmt.Errorf("latest as revision not supported, please specify a revision number")
	}
	if !tools.IsValidRevision(args[1]) {
		return fmt.Errorf("invalid revision")
	}
	revision, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	modelAssertion, err := helper.GetModelAssertionByModelNameAndRevision(cmd.Context(), modelName, revision)
	if err != nil {
		return err
	}

	modelAssertionParsed, err := helper.ParseModelAssertion(modelAssertion.AssertionBody)
	if err != nil {
		return err
	}

	replacementDict := map[string]string{}
	_, tenantName, err := backend.GetStoreOwner(cmd.Context())
	if err != nil {
		return err
	}
	replacementDict["account-id"] = tenantName

	err = setUrlReplacements(cmd.Context(), replacementDict)
	if err != nil {
		return err
	}
	err = setAssertionReplacements(cmd.Context(), modelAssertion, replacementDict)
	if err != nil {
		return err
	}
	err = setSnapIdReplacements(cmd.Context(), replacementDict, *modelAssertionParsed)
	if err != nil {
		return err
	}

	constantsFile := generateConstantsFile(replacementDict)
	fmt.Println(constantsFile)

	// Check that all tokens in the template are replaced
	if strings.Contains(constantsFile, "{{") || strings.Contains(constantsFile, "}}") {
		return fmt.Errorf("not all tokens in the template are replaced")
	}
	return nil
}

func setAssertionReplacements(ctx context.Context, model *gql.Assertion, dict map[string]string) error {
	// Retrieve store assertions. We have to filter manually for the assertion we require for each field
	assertions, err := gql.StoreAssertions(ctx)
	if err != nil {
		panic(err)
	}

	var authority string
	// Get signing key infos
	rootAccountKey, err := helper.FindAssertion(assertions, "account-key", "name: snapstore-key-root")
	if err != nil {
		return err
	}
	rootKeySha3, err := helper.GetAssertionFieldValue(rootAccountKey, "public-key-sha3-384")
	if err != nil {
		return err
	}
	authority, err = helper.GetAssertionFieldValue(rootAccountKey, "authority-id")
	if err != nil {
		return fmt.Errorf("store owner not found in root account key assertion")
	}

	modelAccountKey, err := helper.FindAssertion(assertions, "account-key", "name: snapstore-key-model")
	if err != nil {
		return err
	}
	modelKeySha3, err := helper.GetAssertionFieldValue(modelAccountKey, "public-key-sha3-384")
	if err != nil {
		return err
	}

	// Get account assertions
	authorityAccount, err := helper.FindAssertion(assertions, "account", fmt.Sprintf("display-name: %s", authority))
	if err != nil {
		return err
	}

	// Account
	dict["canonical-account"] = authorityAccount
	dict["generic-account"] = authorityAccount

	dict["staging-trusted-account"] = authorityAccount
	dict["staging-generic-account"] = authorityAccount

	// Account-key
	dict["canonical-root-account-key"] = rootAccountKey
	dict["repair-root-account-key"] = rootAccountKey
	dict["staging-repair-root-account-key"] = rootAccountKey
	dict["staging-root-account-key"] = rootAccountKey

	dict["staging-generic-models-account-key"] = modelAccountKey
	dict["generic-models-account-key"] = modelAccountKey

	// Account-key SHA3
	dict["repair-root-account-key-public-key-sha3"] = rootKeySha3
	dict["canonical-account-sign-key-sha3"] = rootKeySha3

	dict["generic-models-account-key-public-key-sha3"] = modelKeySha3

	// Model
	dict["generic-classic-model"] = model.AssertionBody
	dict["staging-generic-classic-model"] = model.AssertionBody

	return nil
}

func setSnapIdReplacements(ctx context.Context, dict map[string]string, parsed structs.AssertionModel) error {
	for _, snapName := range []string{"snapd", "core", "core18", "core20", "core22", "core24", "core26"} {
		app, err := backend.GetAppByNameAndArch(ctx, snapName, parsed.Architecture)
		if err != nil {
			if snapName == "core" || snapName == "core18" || snapName == "core26" {
				dict["prod-id-"+snapName] = "00000000000000000000000000000000"
				continue
			} else {
				return fmt.Errorf("failed to get snap '%s' info from backend: %s", snapName, err)
			}

		}
		if app.AppSnap == nil {
			return fmt.Errorf("failed to get snap '%s' info from backend - missing snap details", snapName)
		}
		dict["prod-id-"+snapName] = app.AppSnap.SnapId
	}
	return nil
}

func setUrlReplacements(ctx context.Context, dict map[string]string) error {
	// Get user session for the store url
	userSession := viper.AllSettings()

	// Get the store url
	storeUrl := userSession["store"].(string)
	protocol := "https://"
	if strings.HasPrefix(storeUrl, "http://") {
		protocol = "http://"
	}
	rawStoreUrl := strings.TrimPrefix(storeUrl, protocol)
	rawStoreUrl = strings.TrimSuffix(rawStoreUrl, "/graphql")
	rawStoreUrl = strings.TrimSuffix(rawStoreUrl, "/")

	// If the appstore is running locally, we need to replace localhost with the default gateway IP
	// adress of the Docker bridge to be able to access it from the container where the generated constants file will be used.
	// This is only the case if the backend is run for development
	if strings.Contains(rawStoreUrl, "localhost") {
		rawStoreUrl = strings.ReplaceAll(rawStoreUrl, "localhost", "172.17.0.1")
	}

	dict["base-url-snapcraft-dashboard"] = protocol + rawStoreUrl + "/"
	dict["base-url-snapcraft-api"] = protocol + rawStoreUrl + "/"
	dict["base-url-snapcraft-api-v2"] = protocol + rawStoreUrl + "/v2/"
	dict["auth-location"] = rawStoreUrl

	return nil
}

func generateConstantsFile(replacements map[string]string) string {
	template := constantsGoTemplate
	for key, value := range replacements {
		replacementKey := fmt.Sprintf("{{ %s }}", key)
		template = strings.ReplaceAll(template, replacementKey, value)
	}
	return template
}
