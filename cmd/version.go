package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type VersionDetails struct {
	Version   string
	BuildDate string
	GitCommit string
	GitRef    string
	GoVersion string
	OsArch    string
	OsType    string
}

var VersionInfo VersionDetails

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the version of the application",
	Long:  `Shows the version of the application, as well as the build date, git commit hash, git ref, Go version, os/arch and os type.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("PaperCrypt Version %s,\nbuilt on %s, from commit %s,\nfor %s/%s (Go %s)\n", VersionInfo.Version, VersionInfo.BuildDate, VersionInfo.GitCommit, VersionInfo.OsType, VersionInfo.OsArch, VersionInfo.GoVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
