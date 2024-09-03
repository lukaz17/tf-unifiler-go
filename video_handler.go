package main

import (
	"fmt"
	"math"
	"math/big"
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
		m.logger.Warn().Msg("Quality is not set, default value will be used")
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
	limit := generic.TernaryAssign(args.Limit == 0, duration, math.Min(duration, args.Limit))
	limitMs := big.NewInt(int64(limit * float64(1000)))

	outputRoot := generic.TernaryAssign(args.Output == "", path.Dir(inputFile.AbsolutePath), args.Output)
	if !filesystem.IsDirectoryExist(outputRoot) {
		err = filesystem.CreateDirectoryRecursive(outputRoot)
		if err != nil {
			m.logger.Err(err).Msg("Error creating output directory")
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
			return
		}
	}

	isHDR := fileMI.Media.VideoTracks[0].HDRFormat != ""
	// Convert from BT2020 HDR to BT709 using ffmpeg
	// Reference https://web.archive.org/web/20190722004804/https://stevens.li/guides/video/converting-hdr-to-sdr-with-ffmpeg/
	vfHDR := "zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=tonemap=hable:desat=0,zscale=t=bt709:m=bt709:r=tv,format=yuv420p"
	if isHDR {
		m.logger.Info().Str("param", vfHDR).Msg("The video is HDR, Unifiler will attempt to apply colorspace conversion")
	}
	offsetDef, intervalDef := m.DefaultScreenshotParameter(limitMs)
	offset := generic.TernaryAssign(args.Offset == 0, offsetDef, big.NewInt(int64(args.Offset*1000)))
	interval := generic.TernaryAssign(args.Interval == 0, intervalDef, big.NewInt(int64(args.Interval*1000)))
	quality := generic.TernaryAssign(args.Quality == 0, 1, args.Quality)
	outputFilenameFormat := generic.TernaryAssign(quality == 1, path.Join(outputRoot, inputFile.Name+"_%s"+".jpg"), path.Join(outputRoot, inputFile.Name+"_%s_q%d"+".jpg"))
	for t := offset; t.Cmp(limitMs) <= 0; t = new(big.Int).Add(t, interval) {
		outFile := generic.TernaryAssign(quality == 1, fmt.Sprintf(outputFilenameFormat, m.ConvertSecondToTimeCode(t)), fmt.Sprintf(outputFilenameFormat, m.ConvertSecondToTimeCode(t), quality))
		ffmOptions := &exec.FFmpegArgsOptions{
			InputFile:      inputFile.AbsolutePath,
			InputStartTime: nullable.FromInt(int(t.Int64()) / 1000),

			OutputFile:       outFile,
			OutputFrameCount: nullable.FromInt(1),
			QualityFactor:    nullable.FromInt(quality),
			OverwiteOutput:   true,
		}
		if isHDR {
			ffmOptions.VideoFilter = vfHDR
		}

		_, err := exec.Run("ffmpeg", exec.NewFFmpegArgs(ffmOptions))
		if err != nil {
			m.logger.Err(err).Msg("Error taking video file screenshot")
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
			return
		}
		log.Info().Float64("time", float64(t.Int64())/float64(1000)).Str("output", outFile).Msg("Created screenshot successfully")
	}
}

func (mod *VideoModule) ConvertSecondToTimeCode(msec *big.Int) string {
	h := new(big.Int).Div(msec, big.NewInt(3600000))
	msec = new(big.Int).Mod(msec, big.NewInt(3600000))
	m := new(big.Int).Div(msec, big.NewInt(60000))
	msec = new(big.Int).Mod(msec, big.NewInt(60000))
	s := new(big.Int).Div(msec, big.NewInt(1000))
	ms := new(big.Int).Mod(msec, big.NewInt(1000))

	return fmt.Sprintf("%d_%02d_%02d_%03d", h.Int64(), m.Int64(), s.Int64(), ms.Int64())
}

func (mod *VideoModule) DefaultScreenshotParameter(msec *big.Int) (*big.Int, *big.Int) {
	defaults := []struct {
		duration *big.Int
		offset   *big.Int
		interval *big.Int
	}{
		{big.NewInt(120), big.NewInt(1000), big.NewInt(2500)},       // 0 -> 47
		{big.NewInt(420000), big.NewInt(1300), big.NewInt(4300)},    // 27 -> 97
		{big.NewInt(1200000), big.NewInt(1700), big.NewInt(7100)},   // 59 -> 168
		{big.NewInt(3600000), big.NewInt(2300), big.NewInt(12300)},  // 97 -> 292
		{big.NewInt(10800000), big.NewInt(2700), big.NewInt(12700)}, // 283 -> 850
	}
	for _, d := range defaults {
		if msec.Cmp(d.duration) < 0 {
			return d.offset, d.interval
		}
	}
	return big.NewInt(3400), big.NewInt(17100) // 631 -> max
}
