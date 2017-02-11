package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/barthv/epoxy/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("Execute failed")
	}
}
