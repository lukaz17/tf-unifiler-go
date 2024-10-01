package config

import (
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
)

var cfg *RootConfig

func BuildConfig(f string) (*RootConfig, error) {
	k := defaultConfig()
	if filesystem.IsFileExist(f) {
		k, _ = configFromYaml(k, f)
	}
	k, _ = configFromEnv(k)

	var config RootConfig
	err := k.Unmarshal("", &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func defaultConfig() *koanf.Koanf {
	var k = koanf.New(".")

	k.Load(
		structs.Provider(RootConfig{
			Path: &PathConfig{
				FFMpegPath:      "ffmpeg",
				ImageMagickPath: "magick",
				MediaInfoPath:   "mediainfo",
				X264Path:        "x264",
				X265Path:        "x265",
			},
		}, "koanf"),
		nil,
	)

	return k
}

func configFromEnv(k *koanf.Koanf) (*koanf.Koanf, error) {
	err := k.Load(env.Provider("UNIFILER_", ".", func(s string) string {
		return strings.Replace(
			strings.ToLower(
				strings.TrimPrefix(s, "UNIFILER_")), "_", ".", -1)
	}), nil)
	if err != nil {
		return k, err
	}
	return k, nil
}

func configFromYaml(k *koanf.Koanf, f string) (*koanf.Koanf, error) {
	err := k.Load(file.Provider(f), yaml.Parser())
	if err != nil {
		return k, err
	}
	return k, nil
}
