package config

import (
	"testing"

	"github.com/tforceaio/tf-unifiler-go/filesystem"
)

func TestFileConfig(t *testing.T) {
	prepareTests()

	cfg, err := BuildConfig("../../.tests/config/unifiler.yml")
	if err != nil {
		t.Error("Error get config from file", err)
	}
	if cfg.Path.FFMpegPath != "/usr/bin/ffmpeg" {
		t.Errorf("Wrong FFMpegPath. Expected '%s' Actual '%s'", "/usr/bin/ffmpeg", cfg.Path.FFMpegPath)
	}
	if cfg.Path.ImageMagickPath != "magick" {
		t.Errorf("Wrong ImageMagickPath. Expected '%s' Actual '%s'", "magick", cfg.Path.ImageMagickPath)
	}
	if cfg.Path.MediaInfoPath != "/opt/mediainfo/bin/mediainfo" {
		t.Errorf("Wrong MediaInfoPath. Expected '%s' Actual '%s'", "/opt/mediainfo/bin/mediainfo", cfg.Path.MediaInfoPath)
	}
	if cfg.Path.X264Path != "x264" {
		t.Errorf("Wrong X264Path. Expected '%s' Actual '%s'", "x264", cfg.Path.X264Path)
	}
	if cfg.Path.X265Path != "/usr/bin/x265" {
		t.Errorf("Wrong X265Path. Expected '%s' Actual '%s'", "/usr/bin/x265", cfg.Path.X265Path)
	}
}

func prepareTests() {
	contents := []string{
		"paths:",
		"  ffmpeg: /usr/bin/ffmpeg",
		"  mediainfo: /opt/mediainfo/bin/mediainfo",
		"  x265: /usr/bin/x265",
	}

	if !filesystem.IsExist("../../.tests/config") {
		filesystem.CreateDirectoryRecursive("../../.tests/config")
	}
	if !filesystem.IsExist("../../.tests/config/unifiler.yml") {
		filesystem.WriteLines("../../.tests/config/unifiler.yml", contents)
	}
}
