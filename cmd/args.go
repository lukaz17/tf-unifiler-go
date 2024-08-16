package cmd

type Args struct {
	Hash *HashCmd `arg:"subcommand:hash" help:"Compute or verify checksum"`
}

type HashCmd struct {
	Create *HashCreateCmd `arg:"subcommand:create" help:"Compute hash for files and directories"`
}

type HashCreateCmd struct {
	Algorithms []string `arg:"-a,--algo" help:"Hash algorithms to use, multiple supported. Valid algorithms: md5, sha1, sha256"`
	Files      []string `arg:"-f,--file" help:"Files and/or directories to compute hashes"`
	Output     string   `arg:"-o,--out" help:"File or directory to store the result"`
}
