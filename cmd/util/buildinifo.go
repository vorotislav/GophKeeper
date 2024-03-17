package util

import (
	"flag"
	"fmt"
)

var (
	buildVersion = "N/A" //nolint:gochecknoglobals
	buildDate    = "N/A" //nolint:gochecknoglobals
	buildCommit  = "N/A" //nolint:gochecknoglobals
)

func printVersion() {
	_, _ = fmt.Fprintf(
		flag.CommandLine.Output(),
		Version()+"\n",
	)
}

func Version() string {
	return fmt.Sprintf(
		"GophKeeper %s %s (%s)",
		buildVersion,
		buildDate,
		buildCommit,
	)
}
