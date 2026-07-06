package user

import (
	"encoding/json"
	"fmt"
	"m2cpcli/config"
	"m2cpcli/env"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/tools/console"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "prints the current session status",
	Args:  cobra.ExactArgs(0),
	RunE:  runStatusCmd,
}

func init() {
	UserCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("jwt", false, "show the JWT token")
}

type StatusResult struct {
	Status         string                        `json:"status"`
	Session        map[string]interface{}        `json:"session"`
	ConfigFilePath string                        `json:"configFilePath"`
	Stores         map[string]StatusStoreDetails `json:"stores"`
	//CurrentTenant  gql.TenantAlias             `json:"tenant"`
}

type StatusStoreDetails struct {
	Url     string `json:"url"`
	Alias   string `json:"alias"`
	Method  string `json:"method"`
	SshKey  string `json:"ssh-key"`
	SshUser string `json:"ssh-user"`
}

func runStatusCmd(cmd *cobra.Command, args []string) error {

	result := StatusResult{
		Stores: make(map[string]StatusStoreDetails),
	}

	aliasDefinition, err := config.GetAliases()
	if err != nil {
		return err
	}

	jwtString := viper.GetString("jwt")
	jwt, err := env.NewJsonWebToken(jwtString)
	if err != nil {
		return err
	}
	if jwt.IsValid() {
		result.Status = "logged in"
	} else {
		result.Status = "not logged in"
		err = env.InvalidateSessionJwt()
		if err != nil {
			return err
		}
	}

	result.ConfigFilePath = viper.ConfigFileUsed()
	result.Session = viper.AllSettings() // Note: has to happen after `env.InvalidateSessionJwt()`
	if jwt.IsValid() {
		// Add some information from different sources
		timestamp := time.Unix(int64(jwt.ExpirationTime), 0)
		result.Session["jwtExpirationTime"] = timestamp.Format(time.RFC3339)

		// Tenant/permissions info comes from the already-persisted session (see login.go).
		if tenantId := viper.GetString("tenant-id"); tenantId != "" {
			result.Session["tenant"] = &gql.Tenant{
				Id:         gql.UUID(tenantId),
				Alias:      viper.GetString("tenant-alias"),
				TenantName: viper.GetString("tenant-name"),
			}
		}
		// Scopes are deliberately not shown here: "user roles list" is the dedicated
		// command for scopes (and roles), and fetches them fresh itself.
		result.Session["roles"] = viper.GetStringSlice("permissions-roles")
		result.Session["isSuperAdmin"] = viper.GetBool("permissions-is-super-admin")
	}

	// if jwt flag is set
	if cmd.Flag("jwt").Changed {
		result.Session["jwtShow"] = jwtString
	} else {
		result.Session["jwtShow"] = "hidden, use --jwt flag to show"
	}

	for key, value := range result.Session {

		if len(key) > 5 && key[0:6] == "store_" {
			storeDetails := toStoreDetails(value)
			storeDetails.Alias = aliasDefinition.GetAlias(storeDetails.Url)
			result.Stores[storeDetails.Url] = storeDetails
		}
	}

	return format.PrintFormattedOutput(cmd, result, userStatusFormatter)
}

func userStatusFormatter(userStatus StatusResult) (string, error) {
	outputStr := ""

	// Generate the fleet tree

	fleetTree := format.NewTree(console.Colorize(console.Green, "User Status"))
	fleetTree.AddLeaf("Status: " + userStatus.Status)
	fleetTree.AddLeaf(fmt.Sprintf("Store: %s", userStatus.Session["store"]))
	fleetTree.AddLeaf(fmt.Sprintf("Login Expiration: %v", userStatus.Session["jwtExpirationTime"]))
	if userStatus.Session["tenant"] != nil {
		tenant := userStatus.Session["tenant"].(*gql.Tenant)
		fleetTree.AddLeaf("Tenant Alias: " + tenant.Alias)
		fleetTree.AddLeaf("Tenant ID: " + string(tenant.Id))
	}
	if roles, ok := userStatus.Session["roles"].([]string); ok && len(roles) > 0 {
		fleetTree.AddLeaf("Roles: " + strings.Join(roles, ", "))
	}
	if isSuperAdmin, ok := userStatus.Session["isSuperAdmin"].(bool); ok && isSuperAdmin {
		fleetTree.AddLeaf("Super Admin: true")
	}
	fleetTree.AddLeaf("Status File: " + userStatus.ConfigFilePath)
	fleetTree.AddLeaf("JWT: " + userStatus.Session["jwtShow"].(string))

	for _, storeDetails := range userStatus.Stores {

		modelNode := fleetTree.NewChild("Store")
		modelNode.AddLeaf("URL: " + storeDetails.Url)
		modelNode.AddLeaf("Alias: " + storeDetails.Alias)
		modelNode.AddLeaf("Method: " + storeDetails.Method)
		modelNode.AddLeaf("SSH Key: " + storeDetails.SshKey)
		modelNode.AddLeaf("SSH User: " + storeDetails.SshUser)
	}
	outputStr += fleetTree.String()
	return outputStr, nil
}

func toStoreDetails(value any) StatusStoreDetails {
	j, _ := json.Marshal(value)
	storeDetails := StatusStoreDetails{}
	_ = json.Unmarshal(j, &storeDetails)
	return storeDetails
}
