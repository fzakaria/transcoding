# Transcoding CLI tool

This tool is a thin wrapper around ffmpeg (msut be on $PATH) with the hopes of making it more programmatable.

You can provide arbitrary video filter descriptions to modify the input file in question.

## Choosing an appriorate filter description

A common scenario is re-encoding a video for streaming over the web. In such a scenario you generally want to transcode to a lower resolution and bitrate.

### Resolution
Resolution is the "sharpness" of the video in question. 
**Filesize is not determined by resolution**

> filesize (in MB) = (bitrate in Mbit/s * 8) * (video length in seconds)
*But a larger resolution will require more bitrate for it to keep the same level of "quality"*

According to this [post](https://www.virag.si/2012/01/web-video-encoding-tutorial-with-ffmpeg-0-9/) here are some general resolution/bitrate guidelines.

https://www.virag.si/2012/01/web-video-encoding-tutorial-with-ffmpeg-0-9/


Resolution    | Bitrate       | Approx. File size of 10 minutes
------------- | ------------- | -------------------------------
320p (mobile) | 180 kbit/s    | ~13 MB
360p          | 300 kbit/s    |	~22MB
480p          |	500 kbit/s    | ~37MB
576p (PAL)    | 850 kbit/s    | ~63MB
720p          | 1000 kbit/s   | ~75 MB


## Sample Commands

./transcoding 320p -i sample.mp4 -o output.mp4  
