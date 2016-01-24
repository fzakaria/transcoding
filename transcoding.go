/**
  The purpose of this file is to scale and change the bitrate of a given file.

  Useful links that help understand what is going on:

  1. http://dranger.com/ffmpeg/tutorial01.html
**/
package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"os"
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

	flags := []cli.Flag{
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
	}

	app.Commands = []cli.Command{
		{
			Name:  "320p",
			Usage: "Default presets for transcoding to 320p.",
			Flags: flags,
			Action: func(c *cli.Context) {
				if inputFilename == "" || outputFilename == "" {
					fmt.Fprintln(os.Stderr, "You must specify an input/output filename.")
					log.WithFields(log.Fields{
						"input":  inputFilename,
						"output": outputFilename,
					}).Fatal("Missing necessary arguments.")
				}
				scale := "-1:380"
				videoKilobitRate := uint(180)
				audioKilobitRate := uint(128)
				converter := NewFfmpegConverter(inputFilename, outputFilename, scale, videoKilobitRate, audioKilobitRate)
				converter.Transcode()
			},
		},
	}

	app.RunAndExitOnError()
}
