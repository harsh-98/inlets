package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	Version   string
	GitCommit string
)

func init() {
	inletsCmd.Flags().BoolP("version", "v", false, "print the version information")
}

// inletsCmd represents the base command when called without any sub commands.
var inletsCmd = &cobra.Command{
	Use:   "inlets",
	Short: "Expose your local endpoints to the Internet.",
	Long: `
Inlets combines a reverse proxy and websocket tunnels to expose your internal and development
endpoints to the public Internet via an exit-node.

An exit-node may be a 5-10 USD VPS or any other computer with an IPv4 IP address.

See: https://github.com/harsh-98/inlets for more information.`,
	Run: parseBaseCommand,
}

func parseBaseCommand(_ *cobra.Command, _ []string) {
	if len(Version) == 0 {
		fmt.Println("Version: dev")
	} else {
		fmt.Println("Version:", Version)
	}
	fmt.Println("Git Commit:", GitCommit)
	os.Exit(0)
}

// Execute adds all child commands to the root command(InletsCmd) and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the InletsCmd.
func Execute(version, gitCommit string) error {

	// Get Version and GitCommit values from main.go.
	Version = version
	GitCommit = gitCommit

	if err := inletsCmd.Execute(); err != nil {
		return err
	}
	return nil
}
