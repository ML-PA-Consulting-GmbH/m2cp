package ssh

import (
	"github.com/spf13/cobra"
)

var closeCmd = &cobra.Command{
	Use:        "close [deviceName|deviceSerial port]",
	Short:      "Deprecated: Shut down reverse ssh tunnel for device",
	Run:        runCloseCmd,
	Deprecated: "closing is no longer required",
}

func init() {
	sshCmd.AddCommand(closeCmd)
}

func runCloseCmd(_ *cobra.Command, _ []string) {}
