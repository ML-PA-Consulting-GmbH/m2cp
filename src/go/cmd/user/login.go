package user

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"m2cpcli/auth"
	"m2cpcli/backend"
	"m2cpcli/config"
	"m2cpcli/env"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/version"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"gopkg.in/ini.v1"

	"m2cpcli/state"
	"m2cpcli/tools"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var loginCmd = &cobra.Command{
	Use:   "login [alias]",
	Short: "Log into a m2cp backend",
	Long: `Log into a m2cp backend. The default authentication method is browser based authentication. Alternatively, SSH authentication can be used.
A call without parameters will try to re-use parameters from the last login, as stored in the configuration file.`,
	Args: validateLoginArgs,
	//ArgAliases: []string{"user"}, // TODO: how to use them?
	Example: `When logging in the first time, you need to provide at least a backend URL:
  $ m2cp user login --store <url>

The --method "browser" is the the implicit default. 
Alternatively, you can use --method "ssh" and provide your email or your SSH username (as configured in your user account) for logging in:

  $ m2cp user login --method ssh --ssh-user testuser@ml-pa.com --store <url>
  $ m2cp user login --method ssh --ssh-user testuser --store <url>

Next time, all parameters can be omitted, because they will be fetched from the "` + state.DefaultStatePath + `" configuration file, e.g.:

  $ m2cp user login

If you previously logged in with SSH, but now want to switch to browser based authentication, provide the --method flag:

  $ m2cp user login --method browser 

Nevertheless, parameters can be selectively overwritten, by giving parameters, e.g. to use a different
snapstore with the previous authentication details:

  $ m2cp user login --store https://example.com/graphql

For convenience it is possible to store an alias for a store URL:

  $ m2cp user login --store https://example.com/graphql --alias ex1

Once stored, you can use it to log in like this:

  $ m2cp user login ex1`,
	RunE: runLoginCmd,
}

func validateLoginArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.RangeArgs(0, 1)(cmd, args); err != nil {
		return err
	}

	if len(args) > 0 {
		alias := args[0]

		aliasDefinition, err := config.GetAliases()
		if err != nil {
			return err
		}

		_, found := aliasDefinition.GetUrl(alias)
		if found != nil {
			return fmt.Errorf("URL definition for alias \"%s\" does not exist", alias)
		}
	}
	return nil
}

func init() {
	UserCmd.AddCommand(loginCmd)
	// TODO: there is no default value for any customer!
	loginCmd.Flags().String("store", "", "URL of store, normally ends with \"/graphql\"")
	loginCmd.Flags().String("ssh-user", "", "username of email address to use for login")
	loginCmd.Flags().String("ssh-key", "", "path to private ssl key to use for login (~/.ssh/id_rsa if nothing else defined)")
	loginCmd.Flags().String("alias", "", "store an alias for the URL")
	loginCmd.Flags().String("method", "browser", fmt.Sprintf("the authentication method in {%s}",
		env.ListingOfKnownAuthenticationMethods()))
	loginCmd.Flags().Bool("ignore-update", false, "skip checking for updates (Windows only)")

	var err error
	err = viper.BindPFlag("store", loginCmd.Flags().Lookup("store"))
	cobra.CheckErr(err)

	// do not bind to the alias flag

	err = viper.BindPFlag("method", loginCmd.Flags().Lookup("method"))
	cobra.CheckErr(err)
}

func publicKeyPath(privateKeyPath string) string {
	return fmt.Sprintf("%s.pub", privateKeyPath)
}

func isPassphraseProtected(privKeyPem []byte) bool {
	_, err := ssh.ParsePrivateKey(privKeyPem)
	return err != nil && strings.Contains(err.Error(), "private key is passphrase protected")
}

func authenticate(ctx context.Context, privateKeyPath, url, userEmail string) (string, error) {
	pub, err := tools.ReadLocalFile(publicKeyPath(privateKeyPath))
	if err != nil {
		return "", err
	}

	challenge, err := auth.GetChallenge(ctx, url, userEmail, pub)
	if err != nil {
		return "", err
	}

	privKeyPem, err := tools.ReadLocalFile(privateKeyPath)
	if err != nil {
		return "", err
	}
	var privKeyPemPassword []byte
	if isPassphraseProtected(privKeyPem) {
		passwd, err := tools.ConsoleInputPassword(fmt.Sprintf("Enter passphrase for ssh key \"%s\":", privateKeyPath))
		privKeyPemPassword = []byte(passwd)
		if err != nil {
			return "", err
		}
	} else {
		privKeyPemPassword = nil
	}

	signature, err := auth.SignChallenge(challenge, privKeyPem, privKeyPemPassword)
	if err != nil {
		return "", err
	}
	jwt, err := auth.GetJSONWebToken(ctx, url, challenge, signature)
	if err != nil {
		return "", err
	}
	return jwt, nil
}

func updateAliasDefinition(cmd *cobra.Command) error {
	if !cmd.Flags().Changed("store") {
		return fmt.Errorf("--store flag is required when --alias flag is set")
	}
	alias, err := cmd.Flags().GetString("alias")
	if err != nil {
		return err
	}
	storeUrl, err := cmd.Flags().GetString("store")
	if err != nil {
		return err
	}
	defaultStoreUrlAliasDefinitionFilepath, err := config.DefaultStoreUrlAliasDefinitionFilepath()
	if err != nil {
		return err
	}
	err = config.CreateFileIfNotExisting(defaultStoreUrlAliasDefinitionFilepath)
	if err != nil {
		return err
	}
	aliasDefinition := config.NewStoreUrlAliasDefinition()
	err = aliasDefinition.ReadFromJson(defaultStoreUrlAliasDefinitionFilepath)
	if err != nil {
		return err
	}
	err = aliasDefinition.AddOrOverwrite(alias, storeUrl)
	if err != nil {
		return err
	}
	err = aliasDefinition.Write(defaultStoreUrlAliasDefinitionFilepath)
	if err != nil {
		return err
	}
	return nil
}

type loginResultType struct {
	Message        string                 `json:"message"`
	ConfigFilePath string                 `json:"configFilePath"`
	Session        map[string]interface{} `json:"session"`
	ClientUpdate   struct {
		Current           string `json:"current"`
		Available         string `json:"available"`
		UpdateRecommended bool   `json:"updateRecommended"`
		DownloadUrl       string `json:"downloadUrl"`
	} `json:"clientUpdate"`
}

func runLoginCmd(cmd *cobra.Command, args []string) error {
	var loginResult loginResultType

	// Clear any existing JWT and tenant/permissions info to prevent server faults when
	// logging in with a different backend, and to avoid persisting stale info (from a
	// previous account/tenant) alongside a new JWT if this login fails before that info
	// is re-resolved below.
	viper.Set("jwt", "")
	env.ClearTenantAndPermissions()

	method := viper.GetString("method")
	authenticationMethod, err := env.ParseAuthenticationMethod(method)
	if err != nil {
		return fmt.Errorf("%s: please use one of {%s}",
			err, env.ListingOfKnownAuthenticationMethods())
	}

	url := viper.GetString("store")
	sanitizedUrl, err := config.SanitizeStoreUrl(url)
	if err != nil {
		return fmt.Errorf("could not sanitize store url \"%s\": %v", url, err)
	}
	viper.Set("store", sanitizedUrl)
	if sanitizedUrl == "" {
		return fmt.Errorf("the store URL must not be empty: please give a value for \"--store\"")
	}

	if cmd.Flags().Changed("store") && cmd.Flags().Changed("alias") {
		err = updateAliasDefinition(cmd)
		if err != nil {
			return err
		}
	} else {
		if len(args) > 0 {
			alias := args[0]

			defaultStoreUrlAliasDefinitionFilepath, err := config.DefaultStoreUrlAliasDefinitionFilepath()
			if err != nil {
				return err
			}
			err = config.CreateFileIfNotExisting(defaultStoreUrlAliasDefinitionFilepath)
			if err != nil {
				return err
			}
			aliasDefinition := config.NewStoreUrlAliasDefinition()
			err = aliasDefinition.ReadFromJson(defaultStoreUrlAliasDefinitionFilepath)
			if err != nil {
				return err
			}

			url, err = aliasDefinition.GetUrl(alias)
			if err != nil {
				return err
			}
		}
	}

	sshUserArg, err := cmd.Flags().GetString("ssh-user")
	if err != nil {
		return fmt.Errorf("failed to get ssh-user: %s", err)
	}
	sshKeyArg, err := cmd.Flags().GetString("ssh-key")
	if err != nil {
		return fmt.Errorf("failed to get ssh-key: %s", err)
	}
	// TODO: GetSshDetails does also change the config!
	sshUser, sshKey, err := env.GetSshDetails(url, sshUserArg, sshKeyArg)
	if err != nil {
		return fmt.Errorf("failed to get SSH details: %s", err)
	}

	var jwt string
	switch authenticationMethod {
	case env.BrowserAuthentication:
		token, err := auth.LoginWithBrowser(cmd.Context(), url)
		if err != nil {
			return err
		}
		jwt = *token
	case env.SshAuthentication:
		if sshUser == "" {
			return fmt.Errorf("please specify --ssh-user")
		}
		if jwt, err = authenticate(cmd.Context(), sshKey, url, sshUser); err != nil {
			return err
		}
	}

	if err = env.StoreSession(authenticationMethod, sanitizedUrl, sshUser, sshKey, jwt); err != nil {
		return err
	}

	// The JWT no longer necessarily carries a tenant_id/permissions claims (e.g. tokens
	// issued by the external authentication provider used for browser-based login), so
	// they are resolved via the "me" query, which works regardless of auth method.
	// The one store whose backend lacks "me" falls back to the token's own claims (see
	// backend.GetMeWithFallback). Fetched once here (along with the tenant's alias/name)
	// and persisted, so status/roles-list don't need to repeat any of these calls.
	me, err := backend.GetMeWithFallback(cmd.Context())
	if err != nil {
		return fmt.Errorf("could not fetch current user information: %s", err)
	}
	if me.TenantId == "" {
		return fmt.Errorf("could not determine tenant id: \"me\" query returned an empty tenantId")
	}
	tenant, err := gql.TenantById(cmd.Context(), gql.UUID(me.TenantId))
	if err != nil {
		return fmt.Errorf("could not retrieve tenant metadata: %s", err)
	}
	if err = env.StoreTenant(me.TenantId, tenant.Alias, tenant.TenantName); err != nil {
		return err
	}

	var roles []string
	isSuperAdmin := false
	if me.Permissions != nil {
		roles = me.Permissions.Roles
		if me.Permissions.IsSuperAdmin != nil {
			isSuperAdmin = *me.Permissions.IsSuperAdmin
		}
	}
	if err = env.StorePermissions(roles, isSuperAdmin); err != nil {
		return err
	}

	loginResult.Message = "logged in successfully"
	loginResult.ConfigFilePath = viper.ConfigFileUsed()
	loginResult.Session = viper.AllSettings()
	loginResult.ClientUpdate.UpdateRecommended = false

	ignoreUpdate, _ := cmd.Flags().GetBool("ignore-update")
	if !ignoreUpdate && runtime.GOOS == "windows" {
		current, available, _ := checkForUpdates()

		if current != nil && available != nil {
			loginResult.ClientUpdate.Current = current.String()
			loginResult.ClientUpdate.Available = available.String()
		}

		loginResult.ClientUpdate.DownloadUrl = "https://m2cp:m2cp@m2cptools.ml-pa.com/downloads/m2cp-setup-latest.exe"
		loginResult.ClientUpdate.UpdateRecommended = available != nil && current != nil && available.GreaterThan(current)
	}

	return format.PrintFormattedOutput(cmd, loginResult, formatLoginOutput)
}

func formatLoginOutput(result loginResultType) (string, error) {
	out := strings.Builder{}

	if result.Message != "" {
		out.WriteString(result.Message + "\n")
	}

	if result.ClientUpdate.UpdateRecommended && runtime.GOOS == "windows" {
		fmt.Printf("A new version of the m2cp installer is available!\nCurrent:   %s\nAvailable: %s\n\n\n", result.ClientUpdate.Current, result.ClientUpdate.Available)
		err := installLatestWindowsVersion(result.ClientUpdate.DownloadUrl)
		if err != nil {
			fmt.Println(err.Error())
			out.WriteString("Please upgrade manually with:\n\n" + result.ClientUpdate.DownloadUrl + "\n")
		} else {
			fmt.Println("The installer has started. Please follow the installation steps")
			fmt.Println("After setup is complete, run the 'm2cp' command to start using the command line tool")

			os.Exit(0)
		}
	}

	return out.String(), nil
}

func checkForUpdates() (current, available *semver.Version, err error) {
	if current, err = semver.NewVersion(version.Version); err != nil {
		return
	}
	if available, err = getLatestArtifactVersion(); err != nil {
		return
	}

	if available == nil || current == nil || !available.GreaterThan(current) {
		return
	}

	return current, available, nil
}

func getLatestArtifactVersion() (version *semver.Version, err error) {
	var resp *http.Response
	var versionServer *semver.Version

	// TODO: with this, the Linux build depends on the Windows build to work correctly!
	if resp, err = http.Get("https://m2cp:m2cp@m2cptools.ml-pa.com/downloads/latest-version.txt"); err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()

	scanner := bufio.NewScanner(resp.Body)
	if scanner.Scan() {
		firstLine := scanner.Text()
		if versionServer, err = semver.NewVersion(firstLine); err != nil {
			return
		}
		return versionServer, nil
	}

	return nil, fmt.Errorf("could not read version from response")
}

func getDebianVersion() (version *semver.Version, err error) {
	// check the installed version
	cmd := exec.Command("dpkg", "-s", "mlpa-m2cp-dev")
	var output []byte
	if output, err = cmd.Output(); err != nil {
		return
	}
	var cfg *ini.File
	if cfg, err = ini.Load(output); err != nil {
		return
	}
	versionInstalledRaw := cfg.Section("").Key("Version").String()
	var versionInstalled *semver.Version
	if versionInstalled, err = semver.NewVersion(versionInstalledRaw); err != nil {
		return
	}
	mod, _ := versionInstalled.SetPrerelease("")
	return &mod, nil
}

func installLatestWindowsVersion(downloadUrl string) error {
	installerPath := filepath.Join(os.TempDir(), "m2cp-setup-latest.exe")
	resp, err := http.Get(downloadUrl)
	if err != nil {
		return fmt.Errorf("failed to download installer: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download installer: received status code %d", resp.StatusCode)
	}

	out, err := os.Create(installerPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	if _, err = io.Copy(out, resp.Body); err != nil {
		_ = out.Close() // Close before returning in case of an error
		return fmt.Errorf("failed to write installer to file: %w", err)
	}

	if err = out.Close(); err != nil { // Ensure file is closed before executing
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Now start the installer process
	cmd := exec.Command(installerPath)
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("failed to start installer: %w", err)
	}

	return nil
}
