package system

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tredoe/osutil/user/crypt/sha512_crypt"
	"golang.org/x/term"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/helper"
	"math/rand"
	"os"
	"time"
)

var generateSystemUser = &cobra.Command{
	Use:   "generate-system-user",
	Short: "Generate a system user assertion for bundling with an dev image",
	Args:  cobra.ExactArgs(0),
	RunE:  runGenerateSystemUser,
}

var models []string

func init() {
	SystemCmd.AddCommand(generateSystemUser)
	generateSystemUser.Flags().StringSliceVarP(&models, "model", "m", []string{}, "allowed model for the system user (can be used multiple times)")
	generateSystemUser.Flags().StringP("email", "e", "debug@ml-pa.com", "change the email address of the system user")
	generateSystemUser.Flags().StringP("username", "u", "debug", "change the username of the system user")
	generateSystemUser.Flags().StringP("name", "n", "Debug User", "change the name of the system user")
	generateSystemUser.Flags().StringP("password", "p", "", "password for the system user")
}

func runGenerateSystemUser(cmd *cobra.Command, args []string) error {
	var err error
	if len(models) == 0 {
		return fmt.Errorf("at least one model must be provided")
	}
	for _, model := range models {
		_, err = helper.GetModelAssertionByModelNameAndLatestRevision(cmd.Context(), model)
		if err != nil {
			return err
		}
	}

	email, _ := cmd.Flags().GetString("email")
	username, _ := cmd.Flags().GetString("username")
	name, _ := cmd.Flags().GetString("name")
	password, _ := cmd.Flags().GetString("password")

	if password == "" {
		passwordBytes, err := requestUserToInputPasswordFromStd()
		if err != nil {
			return err
		}
		password = string(passwordBytes)
	}

	hashedPassword := encryptPassword(password)

	createSystemUserAssertionInput := gql.SystemUserAssertionCreateInput{
		Models:         models,
		Username:       username,
		Name:           name,
		Email:          email,
		HashedPassword: hashedPassword,
	}

	assertion, err := gql.CreateSystemUserAssertion(cmd.Context(), createSystemUserAssertionInput)
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, assertion, customGenerateSystemUserFormatter)
}

// From: https://stackoverflow.com/questions/62089578/using-go-to-generate-hashed-password-strings-suitable-for-etc-shadow
// There does not seem to be a built-in function in the Go standard library to generate a hashed password string suitable for UNIX passwords
// I checked some results of this against the python crypt library (which is a wrapper around the C library) and it could
// successfully verify the passwords.
// Alternatively, we could call the c crypt function directly using cgo, but that would be more complex.
func encryptPassword(userPassword string) string {
	// Generate a random string for use in the salt
	const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := make([]byte, 8)
	for i := range s {
		s[i] = charset[seededRand.Intn(len(charset))]
	}
	salt := []byte(fmt.Sprintf("$6$%s", s))
	// use salt to hash user-supplied password
	c := sha512_crypt.New()
	hash, err := c.Generate([]byte(userPassword), salt)
	if err != nil {
		fmt.Printf("error hashing user's supplied password: %s\n", err)
		os.Exit(1)
	}
	return string(hash)
}

func requestUserToInputPasswordFromStd() ([]byte, error) {
	fmt.Print("Please enter the password for the system user: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	fmt.Print("\nPlease confirm the password: ")
	confirmPassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return nil, err
	}

	if string(password) != string(confirmPassword) {
		return nil, fmt.Errorf("passwords do not match")
	}
	return password, nil
}

func customGenerateSystemUserFormatter(res *gql.AssertionPlain) (string, error) {
	return fmt.Sprintf("%s", res.Assertion), nil
}
