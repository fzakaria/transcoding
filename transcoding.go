/**
  The purpose of this file is to scale and change the bitrate of a given file.

  Useful links that help understand what is going on:

  1. http://dranger.com/ffmpeg/tutorial01.html
**/
package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "transcoding"
	app.Usage = "Takes a given input file and options and runs them through ffmpeg"
	app.Version = "0.1"
	app.Authors = []cli.Author{{"Farid Zakaria", "farid.m.zakaria@gmail.com"}}
	app.Copyright = "© 2009 Farid Zakaria"

	var inputFilename string
	var outputFilename string
	var debug bool
	var port int

	app.Commands = []cli.Command{
		{
			Name:  "320p",
			Usage: "Default presets for transcoding to 320p.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "input, i",
					Usage:       "the input file to perform the transcoding on",
					Destination: &inputFilename,
				},
				cli.StringFlag{
					Name:        "output, o",
					Usage:       "the output file where to save the output of the transcoding",
					Destination: &outputFilename,
				},
				cli.BoolFlag{
					Name:        "debug, d",
					Usage:       "Whether we should display debug information.",
					Destination: &debug,
				},
			},
			Action: func(c *cli.Context) {
				if inputFilename == "" || outputFilename == "" {
					log.WithFields(log.Fields{
						"input":  inputFilename,
						"output": outputFilename,
					}).Fatal("Missing necessary arguments.")
				}
				if debug {
					log.SetLevel(log.DebugLevel)
				}
				converter := New320pConverter(inputFilename, outputFilename)
				converter.Transcode()
				log.Debug("Finished Transcoding.")
			},
		},
		{
			Name:  "server",
			Usage: "Starts a transcoding server",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:        "port, p",
					Usage:       "the HTTP port for the server to listen to",
					Destination: &port,
					Value:       8080,
					EnvVar:      "PORT",
				},
				cli.BoolFlag{
					Name:        "debug, d",
					Usage:       "Whether we should display debug information or run in production mode.",
					Destination: &debug,
				},
			},
			Action: func(c *cli.Context) {
				if debug {
					log.SetLevel(log.DebugLevel)
				}
				StartServer(port, debug)
			},
		},
	}

	app.RunAndExitOnError()
}
