package main

import (
	"os"

	log "github.com/charmbracelet/log" //charmbracelet logs are so cool
	"github.com/itspacchu/kubeloki/cmd"
)

var logger *log.Logger

func main() {
	logger = log.NewWithOptions(os.Stdout, log.Options{
		Prefix:          " kubeloki::",
		ReportTimestamp: true,
		// ReportCaller:    true,
	})

	logger.Info("Starting")
	if err := cmd.GetKubeDetails(logger); err != nil {
		logger.Fatal(err)
	}
}
