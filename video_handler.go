package main

import (
	"fmt"
	"path"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tforceaio/tf-unifiler-go/cmd"
	"github.com/tforceaio/tf-unifiler-go/extension/generic"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
	"github.com/tforceaio/tf-unifiler-go/filesystem/exec"
	"github.com/tforceaio/tf-unifiler-go/x/nullable"
)

type VideoModule struct {
	logger zerolog.Logger
}

func (m *VideoModule) Video(args *cmd.VideoCmd) {
	if args.Info != nil {
		m.VideoInfo(args.Info)
	} else if args.Screenshot != nil {
		m.VideoScreenshot(args.Screenshot)
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

func (m *VideoModule) VideoScreenshot(args *cmd.VideoScreenshotCmd) {
	if args.File == "" {
		m.logger.Error().Msg("No input file")
		return
	} else if !filesystem.IsFileExist(args.File) {
		m.logger.Error().Str("path", args.File).Msg("Video file not found")
		return
	}
	if args.Output == "" {
		m.logger.Warn().Msg("Output directory is not set, the screenshots will be saved in same directory with media file")
	} else {
		if filesystem.IsFileExist(args.Output) {
			m.logger.Error().Msg("A file with same name with output directory existed")
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
			return
		}
	}
	if args.Interval == 0 {
		m.logger.Warn().Msg("Interval is not set, default value will be used")
	}
	if args.Quality == 0 {
		m.logger.Warn().Msg("Interval is not set, default value will be used")
	}
	m.logger.Info().
		Str("file", args.File).
		Str("output", args.Output).
		Msgf("Taking screenshot for video file")

	inputFile, _ := filesystem.CreateEntry(args.File)
	miOptions := &exec.MediaInfoOptions{
		InputFile:    inputFile.AbsolutePath,
		OutputFormat: "JSON",
	}
	stdout, err := exec.Run("mediainfo", exec.NewMediaInfoArgs(miOptions))
	if err != nil {
		m.logger.Err(err).Msg("Error analyzing video file information")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}
	fileMI, _ := exec.DecodeMediaInfoJson(stdout)

	duration, err := strconv.ParseFloat(fileMI.Media.GeneralTracks[0].Duration, 64)
	if err != nil {
		m.logger.Err(err).Msg("Invalid video file duration")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}

	outputRoot := generic.TernaryAssign(args.Output == "", path.Dir(inputFile.AbsolutePath), args.Output)
	if !filesystem.IsDirectoryExist(outputRoot) {
		err = filesystem.CreateDirectoryRecursive(outputRoot)
		if err != nil {
			m.logger.Err(err).Msg("Error creating output directory")
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
			return
		}
	}
	outputFilenameFormat := path.Join(outputRoot, inputFile.Name+"_%s"+".jpg")

	offsetDef, intervalDef := m.DefaultScreenshotParameter(duration)
	offset := generic.TernaryAssign(args.Offset == 0, offsetDef, args.Offset)
	interval := generic.TernaryAssign(args.Interval == 0, intervalDef, args.Interval)
	quality := generic.TernaryAssign(args.Quality == 0, 1, args.Quality)
	for t := offset; t < duration; t = t + interval {
		outFile := fmt.Sprintf(outputFilenameFormat, m.ConvertSecondToTimeCode(t))
		ffmOptions := &exec.FFmpegArgsOptions{
			InputFile:      inputFile.AbsolutePath,
			InputStartTime: nullable.FromInt(int(t)),

			OutputFile:       outFile,
			OutputFrameCount: nullable.FromInt(1),
			QualityFactor:    nullable.FromInt(quality),
			OverwiteOutput:   true,
		}
		_, err := exec.Run("ffmpeg", exec.NewFFmpegArgs(ffmOptions))
		if err != nil {
			m.logger.Err(err).Msg("Error taking video file screenshot")
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
			return
		}
		log.Info().Float64("time", t).Str("output", outFile).Msg("Created screenshot successfully")
	}
}

func (mod *VideoModule) ConvertSecondToTimeCode(sec float64) string {
	msec := int64(sec * 1000)
	h := msec / int64(3600000)
	msec = msec % int64(3600000)
	m := msec / int64(60000)
	msec = msec % int64(60000)
	s := msec / int64(1000)
	ms := msec % int64(1000)

	return fmt.Sprintf("%d_%02d_%02d_%03d", h, m, s, ms)
}

func (mod *VideoModule) DefaultScreenshotParameter(sec float64) (float64, float64) {
	defaults := []struct {
		duration float64
		offset   float64
		interval float64
	}{
		{120, 1, 2.5},      // 0 -> 47
		{420, 1.3, 4.3},    // 27 -> 97
		{1200, 1.7, 7.1},   // 59 -> 168
		{3600, 2.3, 12.3},  // 97 -> 292
		{10800, 2.7, 12.7}, // 283 -> 850
	}
	for _, d := range defaults {
		if sec <= d.duration {
			return d.offset, d.interval
		}
	}
	return 3.4, 17.1 // 631 -> max
}
