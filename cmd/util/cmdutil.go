package util

import (
	"flag"
	"os"
)

func ParseFlags() string {
	configPath := flag.String("config", "", "Path to configuration file")
	version := flag.Bool("version", false, "Show GophKeeper version")

	flag.Parse()

	if *version {
		printVersion()
		os.Exit(0)
	}

	return *configPath
}
