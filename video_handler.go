package main

import (
	"github.com/rs/zerolog"
	"github.com/tforceaio/tf-unifiler-go/cmd"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
	"github.com/tforceaio/tf-unifiler-go/filesystem/exec"
)

type VideoModule struct {
	logger zerolog.Logger
}

func (m *VideoModule) Video(args *cmd.VideoCmd) {
	if args.Info != nil {
		m.VideoInfo(args.Info)
	} else {
		m.logger.Error().Msg("Invalid arguments")
	}
}

func (m *VideoModule) VideoInfo(args *cmd.VideoInfoCmd) {
	if args.File == "" {
		m.logger.Error().Msg("No input file")
		return
	} else if !filesystem.IsFileExist(args.File) {
		m.logger.Error().Str("path", args.File).Msg("Video file not found")
		return
	}
	m.logger.Info().
		Str("file", args.File).
		Msgf("Analyzing video file information")

	inputFile, _ := filesystem.GetAbsPath(args.File)
	miFile := inputFile + ".mediainfo.json"
	miOptions := &exec.MediaInfoOptions{
		InputFile:    inputFile,
		OutputFormat: "JSON",
		OutputFile:   miFile,
	}

	stdout, err := exec.Run("mediainfo", exec.NewMediaInfoArgs(miOptions))
	if err != nil {
		m.logger.Err(err).Msg("Error analyzing video file information")
	}

	m.logger.Info().Str("path", inputFile).Str("info", string(stdout)).Msg("Analyzed video file information")
	m.logger.Info().Str("path", miFile).Msg("Log file created")
}
