package env

import (
	"fmt"
	"m2cpcli/config"
	"m2cpcli/tools"
	"strings"

	"github.com/spf13/viper"
)

func StoreSession(method AuthenticationMethod, url, userEmail, privateKeyPath, jwt string) error {
	sanitizedUrl, err := config.SanitizeStoreUrl(url)
	if err != nil {
		return fmt.Errorf("could not sanitize store url \"%s\": %v", url, err)
	}
	storeId := UrlToStoreId(sanitizedUrl)

	viper.Set("store", sanitizedUrl)
	viper.Set("store-id", storeId)

	// clear legacy config
	viper.Set("ssh-user", nil)
	viper.Set("ssh-key", nil)

	viper.Set(storeId+".url", sanitizedUrl)
	viper.Set(storeId+".method", method.String())
	viper.Set(storeId+".ssh-user", userEmail)
	viper.Set(storeId+".ssh-key", privateKeyPath)
	viper.Set("jwt", jwt)
	err = viper.WriteConfig()
	if err != nil {
		return fmt.Errorf("could not store session: %s", err)
	}
	return nil
}

// StoreTenant persists the tenant identity resolved once after login (tenantId via the
// "me" query, alias/name via a tenant lookup keyed on that id), so subsequent commands
// can display it without repeating either network call.
func StoreTenant(tenantId, alias, name string) error {
	viper.Set("tenant-id", tenantId)
	viper.Set("tenant-alias", alias)
	viper.Set("tenant-name", name)
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("could not store tenant: %s", err)
	}
	return nil
}

// StorePermissions persists the roles/isSuperAdmin resolved once after login (via the
// "me" query), so subsequent commands (e.g. status) can display them without repeating
// that call. Scopes are not persisted here: "user roles list" is the dedicated command
// for scopes, and fetches them fresh itself.
func StorePermissions(roles []string, isSuperAdmin bool) error {
	viper.Set("permissions-roles", roles)
	viper.Set("permissions-is-super-admin", isSuperAdmin)
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("could not store permissions: %s", err)
	}
	return nil
}

// ClearTenantAndPermissions resets the tenant/permissions viper keys set by StoreTenant/
// StorePermissions to their empty values. Callers are responsible for calling
// viper.WriteConfig() themselves if the clear should be persisted immediately.
func ClearTenantAndPermissions() {
	viper.Set("tenant-id", "")
	viper.Set("tenant-alias", "")
	viper.Set("tenant-name", "")
	viper.Set("permissions-roles", nil)
	viper.Set("permissions-is-super-admin", nil)
}

func GetSshDetails(url string, sshUserArg, sshKeyArg string) (sshUser string, sshKey string, err error) {
	storeId := UrlToStoreId(url)

	// no ssh-key specified and nothing in config - use default
	if viper.GetString(storeId+".ssh-key") == "" && sshKeyArg == "" {
		sshKeyArg = "~/.ssh/id_rsa"
	}

	// no ssh-user specified and nothing in config - try to use from legacy config
	if viper.GetString(storeId+".ssh-user") == "" && sshUserArg == "" {
		sshUserArg = viper.GetString("ssh-user")
	}

	if sshKeyArg != "" {
		if sshKey, err = tools.Abspath(sshKeyArg); err != nil {
			return "", "", fmt.Errorf("could not find absolute path for \"%s\": %v", sshKeyArg, err)
		} else {
			viper.Set(storeId+".ssh-key", sshKey)
			if err = viper.WriteConfig(); err != nil {
				return "", "", fmt.Errorf("could not write configuration: %v", err)
			}
		}
	}

	if sshUserArg != "" {
		viper.Set(storeId+".ssh-user", sshUserArg)
		if err = viper.WriteConfig(); err != nil {
			return "", "", fmt.Errorf("could not write configuration: %v", err)
		}
	}

	sshUser = viper.GetString(storeId + ".ssh-user")
	sshKey = viper.GetString(storeId + ".ssh-key")
	return sshUser, sshKey, nil
}

func UrlToStoreId(url string) string {
	urlSafe := strings.TrimPrefix(url, "https://")
	urlSafe = strings.ReplaceAll(urlSafe, "graphql", "")
	urlSafe = strings.ReplaceAll(urlSafe, ".", "_")
	urlSafe = strings.ReplaceAll(urlSafe, "/", "")
	return "store_" + urlSafe
}

func InvalidateSessionJwt() error {
	viper.Set("jwt", "")
	ClearTenantAndPermissions()
	err := viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}
