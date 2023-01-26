package main

import (
	"github.com/BrainBuzzer/php-ch/cmd/ch"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	ch.Execute()
}
