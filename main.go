package main

import (
	_ "embed"
	"strings"

	"github.com/tmuniversal/papercrypt/cmd"
)

// LicenseText is the license of the application as a string
//
//go:embed COPYING
var LicenseText string

// WordList is the eff.org large word list as a string
//
//go:embed "eff.org_files_2016_07_18_eff_large_wordlist.txt"
var WordList string

// Version is the current version of the application
//
//go:generate sh -c "scripts/get_version.sh > version.gen.txt"
//go:embed version.gen.txt
var Version string

// BuildDate is the date the application was built
//
//go:generate sh -c "scripts/get_date.sh > build_date.gen.txt"
//go:embed build_date.gen.txt
var BuildDate string

// GitCommit is the git commit hash the application was built from
//
//go:generate sh -c "scripts/get_git_commit.sh > git_commit.gen.txt"
//go:embed git_commit.gen.txt
var GitCommit string

// GitRef is the git ref the application was built from
//
//go:generate sh -c "scripts/get_git_ref.sh > git_ref.gen.txt"
//go:embed git_ref.gen.txt
var GitRef string

// GoVersion is the version of the Go compiler used to build the application
//
//go:generate sh -c "go version > go_version.gen.txt"
//go:embed go_version.gen.txt
var GoVersion string

// OsArch is the os/arch the application was built for
//
//go:generate sh -c "go env GOARCH > os_arch.gen.txt"
//go:embed os_arch.gen.txt
var OsArch string

// OsType is the os the application was built for
//
//go:generate sh -c "go env GOOS > os_type.gen.txt"
//go:embed os_type.gen.txt
var OsType string

func main() {
	cmd.LicenseText = &LicenseText
	cmd.WordListFile = &WordList

	cmd.VersionInfo = cmd.VersionDetails{
		Version:   strings.TrimSuffix(Version, "\n"),
		BuildDate: strings.TrimSuffix(BuildDate, "\n"),
		GitCommit: strings.TrimSuffix(GitCommit, "\n"),
		GitRef:    strings.TrimSuffix(GitRef, "\n"),
		GoVersion: strings.TrimSuffix(GoVersion, "\n"),
		OsArch:    strings.TrimSuffix(OsArch, "\n"),
		OsType:    strings.TrimSuffix(OsType, "\n"),
	}

	cmd.Execute()
}
