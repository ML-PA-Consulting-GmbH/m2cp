package system

import (
	"fmt"
	"github.com/spf13/cobra"
	gql "m2cpcli/graphql"
	"os"
)

var assertionCmd = &cobra.Command{
	Use:   "assertion",
	Short: "Get store related assertions of the users tenant",
	RunE:  runAssertionCmd,
}

func init() {
	SystemCmd.AddCommand(assertionCmd)
}

func runAssertionCmd(cmd *cobra.Command, args []string) error {
	assertions, err := gql.StoreAssertions(cmd.Context())
	if err != nil {
		return err
	}

	cmd.SetOut(os.Stdout)
	for i, assertion := range assertions {
		if i < len(assertions)-1 {
			cmd.Print(fmt.Sprintf("%s\n", assertion.Assertion))
		} else {
			cmd.Print(assertion.Assertion)
		}
	}
	return nil
}
