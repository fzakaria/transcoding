package client

/*
 The following is the input structure to request a transcoding job from the server.
 ex.
 	{
		"input": {
			"bucket": "my-input-bucket",
			"key": "fake/filepath/movie.mp4"
		},
		"output": {
			"bucket": "my-output-bucket",
			"key": "fake/filepath/movie.mp4"
		},
		"type" : "320p"
	}
*/
type TranscodeInput struct {
	Input     S3File `json:"input"`
	Output    S3File `json:"output"`
	Type      string `json:"type"`
}

type S3File struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}
