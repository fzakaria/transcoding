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
}

type FfmpegConfig struct {
	Conversions []FfmpegConversion
}

type FfmpegConversion struct {
	Name             string
	Scale            string
	VideoKilobitRate uint `toml:"video_kilobit_rate"`
	AudioKilobitRate uint `toml:"audio_kilobit_rate"`
}
