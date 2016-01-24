package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os/exec"
	"strings"
	"os"
	"io/ioutil"
)

type FfmpegConverter struct {
	inputFilename    string
	outputFilename   string
	videoScale       string
	videoKilobitRate uint
	audioKilobitRate uint
}

func New720pConverter(input, output string) *FfmpegConverter {
	return NewFfmpegConverter(input, output, "-1:720", uint(1000), uint(128))
}

func New576pConverter(input, output string) *FfmpegConverter {
	return NewFfmpegConverter(input, output, "-1:576", uint(850), uint(128))
}

func New480pConverter(input, output string) *FfmpegConverter {
	return NewFfmpegConverter(input, output, "-1:480", uint(500), uint(128))
}

func New360pConverter(input, output string) *FfmpegConverter {
	return NewFfmpegConverter(input, output, "-1:360", uint(300), uint(128))
}

func New320pConverter(input, output string) *FfmpegConverter {
	return NewFfmpegConverter(input, output, "-1:320", uint(180), uint(128))
}

func NewFfmpegConverter(input, output, videoScale string, videoKilobitRate, audioKilobitRate uint) *FfmpegConverter {
	return &FfmpegConverter{input, output, videoScale, videoKilobitRate, audioKilobitRate}
}

func (c *FfmpegConverter) Transcode() (error) {
	ffmpegCmd := func(fullCommand string) (error) {
		log.WithFields(log.Fields{
			"cmd": fullCommand,
		}).Debug("Transcoding file.")
		//we need to split up the command for os.exec
		parts := strings.Fields(fullCommand)
		head, parts := parts[0], parts[1:]
		cmd := exec.Command(head, parts...)
		cmd.Stdout = log.StandardLogger().Out
		cmd.Stderr = log.StandardLogger().Out
		if err := cmd.Run(); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Error during the transcoding.")
			return err
		}
		return nil
	}
	passLogFile, err := ioutil.TempFile("", "ffmpeg2pass")
	if err != nil {
		log.WithFields(log.Fields{
				"error": err,
		}).Error("Error creating ffmpeg2pass log file.")
		return err
	}
	defer os.Remove(passLogFile.Name())

	if err := ffmpegCmd(c.Pass1(passLogFile.Name())); err != nil {
		log.WithFields(log.Fields{
				"error": err,
		}).Error("Error during first pass.")
		return err
	}
	if err := ffmpegCmd(c.Pass2(passLogFile.Name())); err != nil {
		log.WithFields(log.Fields{
				"error": err,
		}).Error("Error during second pass.")
		return err
	}
	return nil
}

/*
 * This is a two pass command for transcoding a file to H.264 and using AAC as the audio codec.
 * Two passes is used to get to a specific filesize/bit-rate (Constant-Bitrate-Encoding)
 * Helpful links to understand what is going on:
 * https://trac.ffmpeg.org/wiki/Encode/H.264
 * https://www.virag.si/2012/01/web-video-encoding-tutorial-with-ffmpeg-0-9/
 */
func (c *FfmpegConverter) Pass1(passlog string) string {
	commandName := "ffmpeg"
	buffsize := c.videoKilobitRate * 2
	firstPass := fmt.Sprintf(
		"%v -y -i %v -passlogfile %v -codec:v libx264 -profile:v high -preset slow -b:v %vk -bufsize %vk -vf scale=%v -threads 0 -pass 1 -c:a libfdk_aac -b:a %vk -f mp4 /dev/null",
		commandName, c.inputFilename, passlog, c.videoKilobitRate, buffsize, c.videoScale, c.audioKilobitRate)
	return firstPass
}

func (c *FfmpegConverter) Pass2(passlog string) string {
	commandName := "ffmpeg"
	buffsize := c.videoKilobitRate * 2
	secondPass := fmt.Sprintf(
		"%v -y -i %v -passlogfile %v -codec:v libx264 -profile:v high -preset slow -b:v %vk -bufsize %vk -vf scale=%v -threads 0 -pass 2 -codec:a libfdk_aac -b:a %vk -f mp4 %v",
		commandName, c.inputFilename, passlog, c.videoKilobitRate, buffsize, c.videoScale, c.audioKilobitRate, c.outputFilename)

	return secondPass
}
