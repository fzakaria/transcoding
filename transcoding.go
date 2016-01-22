package main

import "C"
import "github.com/giorgisio/goav/avutil"
import "github.com/giorgisio/goav/avformat"
import "github.com/giorgisio/goav/avcodec"
import "github.com/giorgisio/goav/avfilter"
import "log"
import "fmt"
import "errors"

func OpenInputFile(filename string) (ctxt *avformat.Context, err error) {
  var ctxtFormat *avformat.Context = nil

  // Open video file
  if avformat.AvformatOpenInput(&ctxtFormat, filename, nil, nil) != 0 {
    return nil,errors.New(fmt.Sprintf("Coudl not open the file %v", filename))
  }

  // Retrieve stream information
  //This function populates pFormatCtx->streams with the proper information.
  if ctxtFormat.AvformatFindStreamInfo(nil) < 0 {
    return nil, errors.New("Could not find the stream information")
  }

  //1 : manually go through each stream, make sure we have the codec for it
  //    and initialize it
  //https://www.ffmpeg.org/doxygen/2.7/avutil_8h_source.html#l00192
  for i := uint(0); i < ctxtFormat.NbStreams(); i++ {
    if ctxtFormat.Streams(i).Codec().CodecType() == avutil.AVMEDIA_TYPE_VIDEO ||
       ctxtFormat.Streams(i).Codec().CodecType() == avutil.AVMEDIA_TYPE_AUDIO {
        var videoCodecContext *avcodec.CodecContext = ctxtFormat.Streams(i).Codec()
        var videoCodec *avcodec.Codec = avcodec.AvcodecFindDecoder(videoCodecContext.CodecId())
        if videoCodecContext.AvcodecOpen2(videoCodec, nil) < 0 {
          return nil, errors.New(fmt.Sprintf("Failed to initialize the codec context for stream %v",i))
        }
    }
  }
  //We introduce a handy debugging function to show us what's inside
  // Dump information about file onto standard error
  ctxtFormat.AvDumpFormat(0, filename, 0)

  return ctxtFormat, nil
}

func OpenOutputFile(filename string) (ctxt *avformat.Context, err error) {
  var ctxtFormat *avformat.Context = nil
  //setting both oformat and formatname makes it guess format from extension
  if avformat.AvformatAllocOutputContext2(&ctxtFormat, nil, "", filename) < 0 {
    return nil, errors.New("Could not create a AVFormatContext for the output")
  }

  for i := uint(0); i < ctxtFormat.NbStreams(); i++ {
    var out_stream * avformat.Stream = ctxtFormat.AvformatNewStream(nil)
    if out_stream == nil {
      return nil, errors.New("Failed allocating stream.")
    }
    var in_stream * avformat.Stream = ctxtFormat.Streams(i)
    //Codec contains decoder and encoder.
    //in -> decoder -> encoder -> out
    var decodingContext *avcodec.CodecContext = in_stream.Codec()
    var encodingContext *avcodec.CodecContext = out_stream.Codec()
    if decodingContext.CodecType() == avutil.AVMEDIA_TYPE_VIDEO ||
       decodingContext.CodecType() == avutil.AVMEDIA_TYPE_AUDIO {
        /* in this example, we choose transcoding to same codec */
        var encoder *avcodec.Codec = avcodec.AvcodecFindDecoder(decodingContext.CodecId())
        if encoder == nil {
          return nil, errors.New("Necessary encoder not found.")
        }
        /* In this example, we transcode to same properties (picture size,
        * sample rate etc.). These properties can be changed for output
        * streams easily using filters */
        if decodingContext.CodecType() == avutil.AVMEDIA_TYPE_VIDEO {
          encodingContext.SetHeight(decodingContext.Height())
          encodingContext.SetWidth(decodingContext.Width())
          encodingContext.SetSampleAspectRatio(decodingContext.SampleAspectRatio())
          encodingContext.SetPixFmt(decodingContext.PixFmt())
          encodingContext.SetTimeBase(decodingContext.TimeBase())
        } else {
          encodingContext.SetSampleRate(decodingContext.SampleRate())
          encodingContext.SetChannelLayout(decodingContext.ChannelLayout())
          encodingContext.SetChannels(decodingContext.Channels())
          encodingContext.SetSampleFmt(decodingContext.SampleFmt())
          encodingContext.SetTimeBase( avutil.NewRational(1, encodingContext.SampleRate()) )
        }

       }
  }

  return ctxtFormat,nil
}

/**
  The purpose of this file is to scale and change the bitrate of a given file.

  Useful links that help understand what is going on:

  1. http://dranger.com/ffmpeg/tutorial01.html
**/
func main() {

	filename := "sample.mp4"
  output := "modified.mp4"

	var (
		inputCtxtFormat *avformat.Context
    outputCtxtFormat *avformat.Context
	)

	// Register all formats and codecs
	avformat.AvRegisterAll()

	avfilter.AvfilterRegisterAll()

  inputCtxtFormat,err := OpenInputFile(filename)

  if err != nil {
    log.Fatal(err)
    return
  }


	//2: use AvFindBestStream
	//https://www.ffmpeg.org/doxygen/2.5/group__lavf__decoding.html#gaa6fa468c922ff5c60a6021dcac09aff9
	var videoCodec2 *avcodec.Codec = nil
	var streamIndex int = avformat.AvFindBestStream(ctxtFormat, avutil.AVMEDIA_TYPE_VIDEO, -1, -1, &videoCodec2, 0)
	if streamIndex < 0 {
		log.Println("Did not automatically find best video stream.")
		return
	}

	log.Println("Option 2: Found video stream at index:", streamIndex)

	if videoCodec2 == nil {
		log.Println("AvFindBestStream did not find a suitable codec.")
		return
	}

	if videoCodecContext.AvcodecOpen2(videoCodec, nil) < 0 {
		log.Println("Failed to intiialize the video codec context.")
		return
	}

	//Let us initialize our filters now
	//https://ffmpeg.org/ffmpeg-filters.html#buffer
	var bufferSrcFilter *avfilter.Filter = avfilter.AvfilterGetByName("buffer")
	//https://ffmpeg.org/ffmpeg-filters.html#buffersink
	var bufferSinkFilter *avfilter.Filter = avfilter.AvfilterGetByName("buffersink")

	var outputs *avfilter.Input = avfilter.AvfilterInoutAlloc()
	var inputs *avfilter.Input = avfilter.AvfilterInoutAlloc()

	var timeBase avutil.Rational = videoStream.TimeBase()

	log.Println("Found Time Base:", timeBase)

	/* buffer video source: the decoded frames from the decoder will be inserted here. */
	filterDescription := fmt.Sprintf("video_size=%dx%d:pix_fmt=%d:time_base=%d/%d:pixel_aspect=%d/%d",
		videoCodecContext.Width(), videoCodecContext.Height(), videoCodecContext.PixFmt(),
		timeBase.Num(), timeBase.Den(),
		videoCodecContext.SampleAspectRatio().Num(), videoCodecContext.SampleAspectRatio().Den())

	log.Println("Initializing filter with description:", filterDescription)

	var graph *avfilter.Graph = avfilter.AvfilterGraphAlloc()

	var bufferSrcFilterContext *avfilter.Context = nil
	if avfilter.AvfilterGraphCreateFilter(&bufferSrcFilterContext, bufferSrcFilter, "in", filterDescription,
		0, graph) < 0 {
		log.Println("Error creating the in filter.")
		return
	}

	/* buffer video sink: to terminate the filter chain. */
	var bufferSinkFilterContext *avfilter.Context = nil
	if avfilter.AvfilterGraphCreateFilter(&bufferSinkFilterContext, bufferSinkFilter, "out", "",
		0, graph) < 0 {
		log.Println("Error creating the out filter.")
		return
	}

	/*
	 * Set the endpoints for the filter graph. The filter_graph will
	 * be linked to the graph described by filters_descr.
	 */

	/*
	 * The buffer source output must be connected to the input pad of
	 * the first filter described by filters_descr; since the first
	 * filter input label is not specified, it is set to "in" by
	 * default.
	 */
	outputs.SetName("in")
	outputs.SetContext(bufferSrcFilterContext)
	outputs.SetPadIndex(0)
	outputs.SetNext(nil)

	/*
	 * The buffer sink input must be connected to the output pad of
	 * the last filter described by filters_descr; since the last
	 * filter output label is not specified, it is set to "out" by
	 * default.
	 */
	inputs.SetName("out")
	inputs.SetContext(bufferSinkFilterContext)
	inputs.SetPadIndex(0)
	inputs.SetNext(nil)

	if graph.AvfilterGraphParsePtr("scale=78:24,transpose=cclock", &inputs, &outputs, 0) < 0 {
		log.Println("Error parsing the filter description and adding to graph.")
		return
	}

	if graph.AvfilterGraphConfig(0) < 0 {
		log.Println("Validity of graph incorrect.")
		return
	}

	//free the memory
	avfilter.AvfilterInoutFree(&inputs)
	avfilter.AvfilterInoutFree(&outputs)
	log.Println("Filter graph initialized.")

	var graphDump string = graph.AvfilterGraphDump("")
	log.Println("Graph dump:\n", graphDump)

	/*
	   Now the magic part! Let us go through each frame and apply the filters
	*/
	// Allocate video frame
	var packet avcodec.Packet = avcodec.Packet{}
	for ctxtFormat.AvReadFrame(&packet) >= 0 {
		// Is this a packet from the video stream? If not lets skip it.
		if packet.StreamIndex() == streamIndex {
			//Multiple packets can compose a frame.
			//We keep consuming packets until they tell us a whole frame finished.
			var frameFinished int = 0
      var frame * avutil.Frame = avutil.AvFrameAlloc()
			videoCodecContext.AvcodecDecodeVideo2(frame, &frameFinished, &packet)
			if frameFinished > 0 {
				//lets push the finished decoded frame into the filter graph.
        bufferSrcFilterContext.AvBuffersrcAddFrame(frame)
			}

      frame.AvFrameFree()
		}

		packet.AvFreePacket()
	}

}
