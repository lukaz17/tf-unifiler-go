package config

type RootConfig struct {
	Path *PathConfig `koanf:"paths"`
}

type PathConfig struct {
	FFMpegPath      string `koanf:"ffmpeg"`
	ImageMagickPath string `koanf:"imagemagick"`
	MediaInfoPath   string `koanf:"mediainfo"`
	X264Path        string `koanf:"x264"`
	X265Path        string `koanf:"x265"`
}
