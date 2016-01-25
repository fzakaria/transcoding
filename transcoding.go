/**
  The purpose of this file is to scale and change the bitrate of a given file.

  Useful links that help understand what is going on:

  1. http://dranger.com/ffmpeg/tutorial01.html
**/
package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
)

func main() {
	var configFilename *string = flag.String("config", "", "Provide the TOML config file to start the server.")
	flag.Parse()
	var config Config
	if _, err := toml.DecodeFile(*configFilename, &config); err != nil {
		log.WithFields(log.Fields{
			"filename": *configFilename,
			"error":    err,
		}).Fatal("Failed to decode config file")
	}
	StartServer(config)
}
