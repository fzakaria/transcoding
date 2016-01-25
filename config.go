package main

type Config struct {
	Aws    AwsConfig
	Server ServerConfig
	Ffmpeg FfmpegConfig
}

type ServerConfig struct {
	Port  int
	Debug bool
}

type AwsConfig struct {
	Region string
	QueueUrl string `toml:"queue_url"`
}

type FfmpegConfig struct {
	Conversions map[string]FfmpegConversion
}

type FfmpegConversion struct {
	Scale            string
	VideoKilobitRate uint `toml:"video_kilobit_rate"`
	AudioKilobitRate uint `toml:"audio_kilobit_rate"`
}
