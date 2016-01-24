package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"strings"
	"os/exec"
)

type FfmpegConverter struct {
	inputFilename    string
	outputFilename   string
	videoScale       string
	videoKilobitRate uint
	audioKilobitRate uint
}

func NewFfmpegConverter(input, output, videoScale string, videoKilobitRate, audioKilobitRate uint) *FfmpegConverter {
	return &FfmpegConverter{input, output, videoScale, videoKilobitRate, audioKilobitRate}
}

func (c *FfmpegConverter) Transcode() {
	cmd := c.Command()
	log.WithFields(log.Fields{
		cmd: cmd,
	}).Debug("Transcoding file.")
	//we need to split up the command for os.exec
	parts := strings.Fields(cmd)
	head, parts := parts[0], parts[1:]
	if err := exec.Command(head, parts...).Run(); err != nil {
		log.WithFields(log.Fields{
		"error": err,
		}).Panic("Error during the transcoding.")
	}
}

/*
 * This is a two pass command for transcoding a file to H.264 and using AAC as the audio codec.
 * Two passes is used to get to a specific filesize/bit-rate (Constant-Bitrate-Encoding)
 * Helpful links to understand what is going on:
 * https://trac.ffmpeg.org/wiki/Encode/H.264
 * https://www.virag.si/2012/01/web-video-encoding-tutorial-with-ffmpeg-0-9/
 */
func (c *FfmpegConverter) Command() string {
	commandName := "ffmpeg"
	buffsize := c.videoKilobitRate * 2
	firstPass := fmt.Sprintf(
		"%v -y -i %v -codec:v libx264 -profile:v high -preset slow -b:v %vk -maxrate %vk -vf scale=%v -threads 0 -pass 1 -c:a libfdk_aac -b:a %vk -f mp4 /dev/null",
		commandName, c.inputFilename, c.videoKilobitRate, buffsize, c.videoScale, c.audioKilobitRate)
	secondPass := fmt.Sprintf(
		"%v -i %v -codec:v libx264 -profile:v high -preset slow -b:v %vk -maxrate %vk -vf scale=%v -threads 0 -pass 2 -codec:a libfdk_aac -b:a %vk -f mp4 %v",
		commandName, c.inputFilename, c.videoKilobitRate, buffsize, c.videoScale, c.audioKilobitRate, c.outputFilename)

	return firstPass + " && \\ \n" + secondPass
}
