package cmd

type Args struct {
	Hash   *HashCmd   `arg:"subcommand:hash" help:"Compute or verify checksum"`
	Mirror *MirrorCmd `arg:"subcommand:mirror" help:"Create links for files and directories to save disk space for similar files"`
}

type HashCmd struct {
	Create *HashCreateCmd `arg:"subcommand:create" help:"Compute hash for files and directories"`
}

type HashCreateCmd struct {
	Algorithms []string `arg:"-a,--algo" help:"Hash algorithms to use, multiple supported. Valid algorithms: md4, md5, ripemd160, sha1, sha224, sha256, sha384, sha512"`
	Files      []string `arg:"-f,--file" help:"Files and/or directories to compute hashes"`
	Output     string   `arg:"-o,--out" help:"File or directory to store the result"`
}

type MirrorCmd struct {
	Export *MirrorExportCmd `arg:"subcommand:export" help:"Generates file structure using cached files and checksum file"`
	Import *MirrorImportCmd `arg:"subcommand:import" help:"Import to cache directory using pre computed checksum file"`
	Scan   *MirrorScanCmd   `arg:"subcommand:scan" help:"Scan files and/or directories and create hardlink to cache directory"`
}

type MirrorExportCmd struct {
	Cache      string `arg:"-c,--cache" help:"Directory to store the cached files. Must be in the same physical partition as files for hardlinks to work"`
	Checksum   string `arg:"-f,--file" help:"Checksum file contains the file structures for export"`
	TargetRoot string `arg:"-r,--root" htlp:"Target root directory that will contains the file structure"`
}

type MirrorImportCmd struct {
	Cache      string `arg:"-c,--cache" help:"Directory to store the cached files. Must be in the same physical partition as files for hardlinks to work"`
	Checksum   string `arg:"-f,--file" help:"Checksum file contains the file structures for export"`
	TargetRoot string `arg:"-r,--root" htlp:"Target root directory that will contains the file structure"`
}

type MirrorScanCmd struct {
	Cache string   `arg:"-c,--cache" help:"Directory to store the cached files. Must be in the same physical partition as files for hardlinks to work"`
	Files []string `arg:"-f,--file" help:"Files and/or directories to import to cache"`
}
