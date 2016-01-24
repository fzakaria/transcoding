package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	el "github.com/deoxxa/echo-logrus"
	"github.com/echo-contrib/echopprof"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/robfig/cron"
	"github.com/thoas/stats"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func InitializeCron() {
	log.Info("Initializing jobs")
	cron := cron.New()
	cron.Start()
}

func TranscodeGet(c *echo.Context) error {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>Transcoder</title>
	</head>
	<body>
		<table>
			<tr>
				<th>Resolution</th>
				<th>Bitrate</th>
				<th>Approx. File size of 10 min video</th>
			</tr>
			<tr>
				<td>320p (mobile)</td><td>180 kbit/sec</td><td>~13MB</td>
			</tr>
			<tr>
				<td>360p</td><td>300 kbit/sec</td><td>~22MB</td>
			</tr>
			<tr>
				<td>480p</td><td>500 kbit/sec</td><td>~37MB</td>
			</tr>
			<tr>
				<td>576p(PAL)</td><td>850 kbit/sec</td><td>~63MB</td>
			</tr>
			<tr>
				<td>720p</td><td>1000 kbit/sec</td><td>~75MB</td>
			</tr>
		</table>
		<form enctype="multipart/form-data" action="/transcode" method="POST">
		 <div>
		    <input type="file" name="input" multiple="multiple"/>
		    <select name="type">
			  <option value="320p">320p</option>
			  <option value="360p">360p</option>
			  <option value="480p">480p</option>
			  <option value="576p">576p</option>
			  <option value="720p">720p</option>
			</select>
		</div>
		<div>
		    <input type="submit" value="Transcode" />
		</div>
		</form>
	</body>
	</html>
	`
	return c.HTML(http.StatusOK, html)
}

var TypeTranscoderMap map[string](func(s1, s2 string) *FfmpegConverter) = map[string](func(s1, s2 string) *FfmpegConverter){
	"320p": New320pConverter,
	"360p": New360pConverter,
	"480p": New480pConverter,
	"576p": New576pConverter,
	"720p": New720pConverter,
}

func TranscodePost(c *echo.Context) error {
	//The 0 here is important because it forces the file
	//to be written to disk, causing us to cast it to os.File
	c.Request().ParseMultipartForm(0)
	mf, _, err := c.Request().FormFile("input")
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing input file.")
		return err
	}
	input := mf.(*os.File)
	defer os.Remove(input.Name())

	output, err := ioutil.TempFile("", "output")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error creating output file.")
		return err
	}
	defer os.Remove(output.Name())

	transcodeFunc, exists := TypeTranscoderMap[c.Form("type")]
	if !exists {
		return c.String(http.StatusBadRequest, "Not a valid transcoding type.")
	}

	converter := transcodeFunc(input.Name(), output.Name())

	converter.Transcode()

	c.Response().Header().Set(echo.ContentType, "video/mp4")
	fi, err := output.Stat()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error retrieving size of file.")
		return err
	}
	c.Response().Header().Set(echo.ContentLength, strconv.FormatInt(fi.Size(), 10))

	if err := c.File(output.Name(), "output.mp4", true); err != nil {
		c.String(http.StatusInternalServerError, "Error sending file.")
		return err
	}

	return nil
}

func StartServer(port int, debug bool) {
	hostname := fmt.Sprintf(":%v", port)

	// Echo instance
	e := echo.New()
	//auto creating an index page for the directory
	e.AutoIndex(true)
	e.SetDebug(debug)

	// Middleware
	e.Use(el.New())
	e.Use(mw.Recover())
	//e.Use(mw.Gzip())

	// Routes
	e.Get("/ping", func(c *echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	//One off transcoding
	e.Get("/transcode", TranscodeGet)
	e.Post("/transcode", TranscodePost)

	// automatically add routers for net/http/pprof
	// e.g. /debug/pprof, /debug/pprof/heap, etc.
	echopprof.Wrapper(e)

	// https://github.com/thoas/stats
	s := stats.New()
	e.Use(s.Handler)
	// Route
	e.Get("/stats", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, s.Data())
	})

	// Start server
	log.WithFields(log.Fields{
		"port":  port,
		"debug": debug,
	}).Info("Starting the server...")
	for _, route := range e.Routes() {
		log.Info(route.Method + " " + route.Path)
	}
	e.Run(hostname)
}
