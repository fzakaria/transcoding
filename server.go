package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
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
	return c.File("./public/views/transcode.html", "", false)
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

	if err := converter.Transcode(); err != nil {
		c.String(http.StatusInternalServerError, "Error transcoding the file.")
		return err
	}

	c.Response().Header().Set(echo.ContentType, "video/mp4")
	fi, err := output.Stat()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error retrieving size of file.")
		return err
	}
	c.Response().Header().Set(echo.ContentLength, strconv.FormatInt(fi.Size(), 10))

	if err := c.File(output.Name(), "output.mp4", false); err != nil {
		c.String(http.StatusInternalServerError, "Error sending file.")
		return err
	}

	return nil
}

func configHandler(config Config) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		encoder := toml.NewEncoder(w)
		if err := encoder.Encode(config); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
	return http.HandlerFunc(fn)
}

func StartServer(config Config) {
	port := config.Server.Port
	debug := config.Server.Debug
	hostname := fmt.Sprintf(":%v", port)

	// Echo instance
	e := echo.New()
	//auto creating an index page for the directory
	e.AutoIndex(true)
	//enable some helpful debug settings
	if debug {
		log.SetLevel(log.DebugLevel)
		e.SetDebug(debug)
	}

	// Middleware
	e.Use(el.New())
	e.Use(mw.Recover())
	e.Use(mw.Gzip())

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

	// Route for some basic statics
	// https://github.com/thoas/stats
	s := stats.New()
	e.Use(s.Handler)
	e.Get("/stats", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, s.Data())
	})

	// Route to see the configuration we are using
	e.Get("/config", configHandler(config))

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
