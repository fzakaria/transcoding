package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	el "github.com/deoxxa/echo-logrus"
	"github.com/fzakaria/transcoding/client"
	"github.com/fzakaria/transcoding/ffmpeg"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/thoas/stats"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
)

func TranscodeJsonPost(conversions map[string]FfmpegConversion) echo.HandlerFunc {
	fn := func(c *echo.Context) error {
		input := &client.TranscodeInput{}
		if err := c.Bind(input); err != nil {
			return err //return Unsupported Media Type or BadRequest
		}
		return nil
	}
	return fn
}

/*
func TranscodeSQS(config AwsConfig, conversions map[string]FfmpegConversion) sqs.Handler {
	fn := func(msg *string) error {
		log.WithFields(log.Fields{
			"msg": *msg,
		}).Info("Received a SQS message.")
		var s3Event s3e.Event
		if err := json.Unmarshal([]byte(*msg), &s3Event); err != nil {
			log.WithFields(log.Fields{
				"msg": *msg,
			}).Warn("Received a SQS message we don't know how to handle. Consuming it.")
			return nil
		}

		svc := s3.New(session.New(&aws.Config{Region: aws.String(config.Region)}))
		for _, record := range s3Event.Records {

			if !strings.HasPrefix(record.EventName, "ObjectCreated") {
				log.WithFields(log.Fields{
					"type": record.EventName,
				}).Debug("Ignoring non-object created messages..")
				continue
			}

			getObjectParams := &s3.GetObjectInput{
				Bucket: aws.String(record.S3.Bucket.Name),
				Key:    aws.String(record.S3.Object.Key),
			}
			resp, err := svc.GetObject(getObjectParams)
			if err != nil {
				log.WithFields(log.Fields{
					"error":   err.Error(),
					"code":    err.(awserr.Error).Code(),
					"message": err.(awserr.Error).Message(),
				}).Warn("Issue occured fetching object.")
				return err
			}

			input, err := ioutil.TempFile("", "s3Input")
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Error creating s3 input temporary file.")
				return err
			}
			defer os.Remove(input.Name())
			//Copy over the buffer to the file
			_, errCopy := io.Copy(input, resp.Body)
			if errCopy != nil {
				log.WithFields(log.Fields{
					"error": errCopy,
				}).Error("Error copying object to temporary input file")
				return errCopy
			}

			//TODO: perform multiple conversions based on metadata
			conversion, _ := conversions["320p"]
			output, err := ioutil.TempFile("", "output")
			if err != nil {
				return err
			}
			defer os.Remove(output.Name())
			converter := ffmpeg.NewConverter(input.Name(), output.Name(), conversion.Scale,
				conversion.VideoKilobitRate, conversion.AudioKilobitRate)

			if err := converter.Transcode(); err != nil {
				return err
			}

			fi, _ := output.Stat()
			//Begin Upload
			putObjectParams := &s3.PutObjectInput{
				Bucket:        aws.String(record.S3.Bucket.Name),
				Key:           aws.String("320p" + record.S3.Object.Key),
				ContentType:   aws.String("video/mp4"),
				ContentLength: aws.Int64(fi.Size()),
				Body:          output,
			}

			_, err1 := svc.PutObject(putObjectParams)
			if err1 != nil {
				log.WithFields(log.Fields{
					"error": err1,
				}).Error("Error putting s3 ouypuy temporary file.")
				return err1
			}

		}
		return nil
	}
	return sqs.HandlerFunc(fn)
}
*/

func TranscodeGet(c *echo.Context) error {
	return c.File("./public/views/transcode.html", "", false)
}

func TranscodePost(conversions map[string]FfmpegConversion) echo.HandlerFunc {
	fn := func(c *echo.Context) error {
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

		conversion, exists := conversions[c.Form("type")]
		if !exists {
			return c.String(http.StatusBadRequest, "Not a valid transcoding type.")
		}

		converter := ffmpeg.NewConverter(input.Name(), output.Name(), conversion.Scale,
			conversion.VideoKilobitRate, conversion.AudioKilobitRate)

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

		if err := c.File(output.Name(), "output.mp4", true); err != nil {
			c.String(http.StatusInternalServerError, "Error sending file.")
			return err
		}

		return nil
	}
	return fn
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
	// https://github.com/thoas/stats
	s := stats.New()

	// Middleware
	e.Use(
		el.New(),
		mw.Recover(),
		mw.Gzip(),
		s.Handler,
	)

	/*
	*    Admin routes
	*   The following are some high level administration routes.
	*/
	admin := e.Group("/admin")
	admin.Get("", func(c *echo.Context) error {
		return c.File("./public/views/admin.html", "", false)
	})
	//ping-pong
	admin.Get("/ping", func(c *echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
	admin.Get("/stats", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, s.Data())
	})
	// Route to see the configuration we are using
	admin.Get("/config", configHandler(config))
	//pprof
	admin.Get("/pprof", http.HandlerFunc(pprof.Index))
	admin.Get("/pprof/heap", pprof.Handler("heap").ServeHTTP)
	admin.Get("/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	admin.Get("/pprof/block", pprof.Handler("block").ServeHTTP)
	admin.Get("/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	admin.Get("/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	admin.Get("/pprof/profile", http.HandlerFunc(pprof.Profile))
	admin.Get("/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	admin.Get("/pprof/trace", http.HandlerFunc(pprof.Trace))

	/*
	*   View routes
	*   The following are the view routes
	*/
	e.Get("/transcode", TranscodeGet)
	e.Post("/transcode", TranscodePost(config.Ffmpeg.Conversions))


	/*
	*   API routes
	*   The following are the API routes
	*/
	g := e.Group("/api")
	g.Post("/transcode", TranscodeJsonPost(config.Ffmpeg.Conversions))

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
