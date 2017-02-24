package main

import (
	log "github.com/Sirupsen/logrus"

	"github.com/BarthV/epoxy/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("Execute failed")
	}
}
