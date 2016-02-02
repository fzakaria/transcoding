# Transcoding Server tool

## Sample Commands

```
brew install ffmpeg
go get github.com/fzakaria/transcoding
go install github.com/fzakaria/transcoding
#Assumes $GOPATH/bin is on your $PATH 
transcoding --config ./configs/prod-us-east-1.toml
```

## Server
A basic server implementation is included that offers the facility to transcode uploaded multipart files and some additional admin urls.

```                               
GET /transcode                               
POST /transcode 
#Some admin routes
GET /admin 
GET /admin/ping                                 
GET /admin/pprof/                            
GET /admin/pprof/heap                        
GET /admin/pprof/goroutine                   
GET /admin/pprof/block                       
GET /admin/pprof/threadcreate                
GET /admin/pprof/cmdline                     
GET /admin/pprof/profile                     
GET /admin/pprof/symbol                      
GET /admin/stats    
GET /admin/config
```

###Transcoding
A useful route is the `[POST|GET] /transcode` one, which provided a file and conversion type will return the resulting MP4 file. A sample form is provided at the `GET` route, however you can also make use of CLI tools.

```
brew install http 
http -f POST http://localhost:8080/transcode input@~/Downloads/sample.mp4 type=480p  > output.mp4
```

## Docker
To make bootstrapping easier for variety of platforms. A Dockerfile is provided which will run the server in a [docker](https://www.docker.com/) container.

```
#The following commands assumes you are in the package
docker build -t transcode-server
docker run -p 8080:8080 transcode-server   
#You can now access the server at localhost:8080
#or if you ar on mac osx `docker-machine ip default`
```

## Choosing an appriorate filter description

A common scenario is re-encoding a video for streaming over the web. In such a scenario you generally want to transcode to a lower resolution and bitrate.

### Resolution
Resolution is the "sharpness" of the video in question. 
**Filesize is not determined by resolution**

> filesize (in MB) = (bitrate in Mbit/s * 8) * (video length in seconds)
*But a larger resolution will require more bitrate for it to keep the same level of "quality"*

According to this [post](https://www.virag.si/2012/01/web-video-encoding-tutorial-with-ffmpeg-0-9/) here are some general resolution/bitrate guidelines.

Resolution    | Bitrate       | Approx. File size of 10 minutes
------------- | ------------- | -------------------------------
320p (mobile) | 180 kbit/s    | ~13 MB
360p          | 300 kbit/s    |	~22MB
480p          |	500 kbit/s    | ~37MB
576p (PAL)    | 850 kbit/s    | ~63MB
720p          | 1000 kbit/s   | ~75 MB
