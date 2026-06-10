package roles

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"
	"m2cpcli/tools/console"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list [user-email]",
	Short: "List user roles of current or given user",
	Long:  "Output a list of roles assigned to the logged in or given user.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runListCmd,
}

type listOutput struct {
	CurrentUser *structs.User
	Claims      jwtTokenClaims
	OtherUser   *structs.User
}

type jwtTokenClaims struct {
	Roles  []string `json:"role_values"`
	Scopes []string `json:"scopes"`
	Tenant string   `json:"tenant_alias"`
}

func init() {
	rolesCmd.AddCommand(listCmd)
}

func runListCmd(cmd *cobra.Command, args []string) (err error) {
	output := listOutput{}

	var user *structs.User

	if len(args) > 0 && args[0] != "" {
		if user, err = backend.GetUserByEmail(cmd.Context(), args[0]); err != nil {
			return err
		}
		output.OtherUser = user

		output.Claims = jwtTokenClaims{
			Tenant: output.OtherUser.TenantName,
			Roles:  []string{},
			Scopes: []string{},
		}

		userRoles, err := backend.GetUserRolesByUserIdAndTenantId(cmd.Context(), user.Id, user.TenantId)
		if err != nil {
			return err
		}

		seen := make(map[string]struct{})

		for _, bridge := range userRoles.UserBridgeRoles.Items {
			output.Claims.Roles = append(output.Claims.Roles, bridge.Role.Name)

			for _, scope := range bridge.Role.RoleBridgeScopes {
				if _, exists := seen[scope.Scope.Value]; !exists {
					output.Claims.Scopes = append(output.Claims.Scopes, scope.Scope.Value)
					seen[scope.Scope.Value] = struct{}{}
				}

			}
		}
	} else {

		// Get roles for the current user

		jwtToken := viper.GetString("jwt")

		if jwtToken == "" {
			return fmt.Errorf("no JWT token found. Please login first")
		}

		user, err = backend.GetCurrentUser(cmd.Context())
		if err != nil {
			return err
		}

		parts := strings.Split(jwtToken, ".")
		payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
		var claims jwtTokenClaims

		err = claims.unmarshalJson(payload)
		if err != nil {
			return err
		}

		output.CurrentUser = user
		output.Claims = claims
	}

	if err = format.PrintFormattedOutput(cmd, output, listOutputFormatter); err != nil {
		return err
	}

	return nil
}

func (c *jwtTokenClaims) unmarshalJson(data []byte) error {
	type Alias jwtTokenClaims
	aux := &struct {
		Roles  interface{} `json:"role_values"`
		Scopes interface{} `json:"scopes"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.Roles.(type) {
	case string:
		c.Roles = []string{v}
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				c.Roles = append(c.Roles, str)
			}
		}
	default:
		return fmt.Errorf("invalid type for role_values")
	}

	switch v := aux.Scopes.(type) {
	case string:
		c.Scopes = []string{v}
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				c.Scopes = append(c.Scopes, str)
			}
		}
	default:
		return fmt.Errorf("invalid type for scopes")
	}

	return nil
}

func listOutputFormatter(output listOutput) (string, error) {
	outputStr := ""

	sort.Strings(output.Claims.Roles)
	sort.Strings(output.Claims.Scopes)

	if output.CurrentUser != nil {
		outputStr += fmt.Sprintf("\n%s  ", console.Colorize(console.Green, "Current User"))
		outputStr += fmt.Sprintf("\n  ├─ Name:   %s", output.CurrentUser.Name)
		outputStr += fmt.Sprintf("\n  ├─ Email:  %s", output.CurrentUser.Email)
		outputStr += fmt.Sprintf("\n  ├─ Tenant: %s", output.Claims.Tenant)
	} else {
		outputStr += fmt.Sprintf("\n%s  ", console.Colorize(console.Green, "User"))
		outputStr += fmt.Sprintf("\n  ├─ Name:   %s", output.OtherUser.Name)
		outputStr += fmt.Sprintf("\n  ├─ Email:  %s", output.OtherUser.Email)
		outputStr += fmt.Sprintf("\n  ├─ Tenant: %s", output.OtherUser.TenantName)
	}

	outputStr += fmt.Sprintf("\n  ├─ Roles")
	for idx, role := range output.Claims.Roles {
		if idx < len(output.Claims.Roles)-1 {
			outputStr += fmt.Sprintf("\n  │    ├─ %s", role)
		} else {
			outputStr += fmt.Sprintf("\n  │    └─ %s", role)
		}
	}

	outputStr += fmt.Sprintf("\n  └─ Scopes")

	for idx, scope := range output.Claims.Scopes {
		if idx < len(output.Claims.Scopes)-1 {
			outputStr += fmt.Sprintf("\n       ├─ %s", scope)
		} else {
			outputStr += fmt.Sprintf("\n       └─ %s", scope)
		}
	}

	return outputStr, nil
}
