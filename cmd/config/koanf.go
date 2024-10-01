package config

import (
	"os"
	"path"
	"runtime"
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
	return &config, err
}

func InitKoanf() (*RootConfig, error) {
	if cfg != nil {
		return cfg, nil
	}
	isPortable := IsPortable()
	configFile := "unifiler.yml"
	if isPortable {
		exec, _ := os.Executable()
		exec, _ = filesystem.GetAbsPath(exec)
		configFile = path.Join(path.Dir(exec), "unifiler.yml")
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		home := os.Getenv("HOME")
		configFile = path.Join(home, ".config", "unifiler", "unifiler.yml")
	} else if runtime.GOOS == "windows" {
		appData := filesystem.NormalizePath(os.Getenv("APPDATA"))
		configFile = path.Join(appData, "Unifiler", "unifiler.yml")
	}
	var err error
	cfg, err = BuildConfig(configFile)
	if err != nil {
		return cfg, err
	}

	cfg.ConfigDir = path.Dir(configFile)
	cfg.ConfigFile = configFile
	cfg.IsPortable = isPortable
	return cfg, nil
}

func IsPortable() bool {
	exec, _ := os.Executable()
	exec, _ = filesystem.GetAbsPath(exec)
	portableFile := path.Join(path.Dir(exec), "unifiler.portable")
	return filesystem.IsFileExist(portableFile)
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
