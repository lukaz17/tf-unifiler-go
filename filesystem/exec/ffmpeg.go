package exec

import (
	"strconv"

	"github.com/tforceaio/tf-unifiler-go/x/nullable"
)

type FFmpegArgs struct {
	Options *FFmpegArgsOptions
}

type FFmpegArgsOptions struct {
	InputFile      string
	InputStartTime nullable.Int

	OutputFile       string
	OutputFrameCount nullable.Int
	OutputStartTime  nullable.Int
	QualityFactor    nullable.Int
	VideoFilter      string
	OverwiteOutput   bool
}

func (args FFmpegArgs) Compile() []string {
	results := []string{}
	if args.Options.InputStartTime.IsValid {
		results = append(results, "-ss", strconv.Itoa(args.Options.InputStartTime.RealValue))
	}
	if args.Options.InputFile != "" {
		results = append(results, "-i", args.Options.InputFile)
	}
	if args.Options.OutputStartTime.IsValid {
		results = append(results, "-ss", strconv.Itoa(args.Options.OutputStartTime.RealValue))
	}
	if args.Options.OutputFrameCount.IsValid {
		results = append(results, "-frames", strconv.Itoa(args.Options.OutputFrameCount.RealValue))
	}
	if args.Options.QualityFactor.IsValid {
		results = append(results, "-q", strconv.Itoa(args.Options.QualityFactor.RealValue))
	}
	if args.Options.VideoFilter != "" {
		results = append(results, "-vf", args.Options.VideoFilter)
	}
	if args.Options.OverwiteOutput {
		results = append(results, "-y")
	}
	if args.Options.OutputFile != "" {
		results = append(results, args.Options.OutputFile)
	}
	return results
}

func NewFFmpegArgs(options *FFmpegArgsOptions) FFmpegArgs {
	return FFmpegArgs{options}
}
