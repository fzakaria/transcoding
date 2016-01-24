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
		<form enctype="multipart/form-data" action="/transcode" method="POST">
		    <input type="file" name="input" multiple="multiple"/>
		    <input type="submit" value="Transcode" />
		</form>
	</body>
	</html>
	`
	return c.HTML(http.StatusOK, html)
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

	converter := New320pConverter(input.Name(), output.Name())

	converter.Transcode()

	c.Response().Header().Set(echo.ContentType, "video/mp4")
	fi, err := output.Stat()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error retrieving size of file.")
		return err
	}
	c.Response().Header().Set(echo.ContentLength,  strconv.FormatInt(fi.Size(), 10))

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
