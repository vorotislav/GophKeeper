package main

import "flag"

func parseFlag() string {
	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "path to config file")

	flag.Parse()

	return configFile
}
