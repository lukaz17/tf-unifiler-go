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

package cmd

type Args struct {
	File   *FileCmd   `arg:"subcommand:file" help:"Batch files processing"`
	Mirror *MirrorCmd `arg:"subcommand:mirror" help:"Create links for files and directories to save disk space for similar files"`
	Video  *VideoCmd  `arg:"subcommand:video" help:"Special operations related to video"`
}

type FileCmd struct {
	Delete *FileDeleteCmd `arg:"subcommand:delete" help:"Delete the files and add them to obsoletion lists for future cleanup"`
	Rename *FileRenameCmd `arg:"subcommand:rename" help:"Multi file rename"`
	Scan   *FileScanCmd   `arg:"subcommand:scan" help:"Scan the files and add them to collection"`
}

type FileDeleteCmd FileScanCmd

type FileRenameCmd struct {
	Files  []string `arg:"-f, --file" help:"Files and/or directories to rename"`
	Preset string   `arg:"-p,--preset" help:"Preset for creating new file name"`
}

type FileScanCmd struct {
	Collections []string `arg:"-c,--collection" help:"Name of collection of known files. If collection existed, files will be appended to that collection."`
	Files       []string `arg:"-f,--file" help:"Files and/or directories to delete."`
	Workspace   string   `arg:"-w,--workspace" help:"Custom workspace location that store the metadata."`
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

type VideoCmd struct {
	Info       *VideoInfoCmd       `arg:"subcommand:info" help:"Analyze video info using MediaInfo"`
	Screenshot *VideoScreenshotCmd `arg:"subcommand:screenshot" help:"Create screenshots for the video to overview of the content"`
}

type VideoInfoCmd struct {
	File   string `arg:"-f, --file" help:"Video file to generate info"`
	Output string `arg:"-o, --out" help:"File to store the info report"`
}

type VideoScreenshotCmd struct {
	File     string  `arg:"-f, --file" help:"Video file to generate info"`
	Interval float64 `arg:"-i, --interval" help:"Time in the second every subsequence screenshot will take"`
	Offset   float64 `arg:"-s, --offset" help:"Time in the second the first screenshot will take"`
	Limit    float64 `arg:"-l, --limit" help:"Time in the second the last screenshot will take"`
	Quality  int     `arg:"-q, --quality" help:"Quality factor for output screenshots"`
	Output   string  `arg:"-o, --out" help:"Directory to store the screenshots"`
}
