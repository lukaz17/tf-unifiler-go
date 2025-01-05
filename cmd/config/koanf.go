// Copyright (C) 2024 T-Force I/O
// This file is part of TF Unifiler
//
// TF Unifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TF Unifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TF Unifiler. If not, see <https://www.gnu.org/licenses/>.

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
