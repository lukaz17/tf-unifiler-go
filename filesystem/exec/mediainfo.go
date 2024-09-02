package exec

import (
	"fmt"
)

type MediaInfoArgs struct {
	Options *MediaInfoOptions
}

type MediaInfoOptions struct {
	InputFile    string
	OutputFormat string
	OutputFile   string
}

func (args MediaInfoArgs) Compile() []string {
	results := []string{}
	if args.Options.OutputFormat != "" {
		results = append(results, fmt.Sprintf("--output=%s", args.Options.OutputFormat))
	}
	if args.Options.OutputFile != "" {
		results = append(results, fmt.Sprintf("--logfile=%s", args.Options.OutputFile))
	}
	results = append(results, args.Options.InputFile)
	return results
}

func NewMediaInfoArgs(options *MediaInfoOptions) MediaInfoArgs {
	return MediaInfoArgs{options}
}
