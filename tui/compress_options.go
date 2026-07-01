// Copyright (C) 2025 T-Force I/O
// This file is part of TFunifiler
//
// TFunifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TFunifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TFunifiler. If not, see <https://www.gnu.org/licenses/>.

package tui

import (
	"fmt"
	"slices"

	"github.com/tforce-io/tf-golib-extra/tftea-v2"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler/core/compression"
)

// CompressOptionsValue stores user-selected options.
type CompressOptionsValue struct {
	ArchiveType    string
	CompressLevel  compression.Level
	SolidMode      bool
	DictionarySize int
	ThreadNumber   int
	MoveToArchive  bool
	NormalizeAttrs bool
	SeparateInputs bool
}

// CompressOptionsModel contains internal state of the CompressOptions.
type CompressOptionsModel struct {
	archiveTypes    []string
	dictionarySizes []int
	threadNumbers   []int
	selectedIndices []int
}

// Return a new CompressOptions instance.
func NewCompressOptions() *CompressOptionsModel {
	archiveTypes := []string{"7z", "rar"}
	dictionarySizes := []int{32, 64, 96, 128, 256, 384, 512, 640, 768, 896, 1024}
	threadNumbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	return &CompressOptionsModel{
		archiveTypes:    archiveTypes,
		dictionarySizes: dictionarySizes,
		threadNumbers:   threadNumbers,
		selectedIndices: make([]int, 0),
	}
}

// Set Selected Options.
func (m *CompressOptionsModel) WithSelected(opts *CompressOptionsValue) *CompressOptionsModel {
	m.selectedIndices = []int{
		slices.Index(m.archiveTypes, opts.ArchiveType),
		int(opts.CompressLevel),
		opx.Ternary(opts.SolidMode, 1, 0),
		slices.Index(m.dictionarySizes, opts.DictionarySize),
		slices.Index(m.threadNumbers, opts.ThreadNumber),
		opx.Ternary(opts.MoveToArchive, 1, 0),
		opx.Ternary(opts.NormalizeAttrs, 1, 0),
		opx.Ternary(opts.SeparateInputs, 1, 0),
	}
	return m
}

// Display the CompressOptions to the terminal.
func (m *CompressOptionsModel) Run() (*CompressOptionsValue, error) {
	compressLevels := []string{"None", "Fast", "Normal", "High", "Ultra"}
	boolValues := []string{"No", "Yes"}
	toStringSlice := func(ints []int) []string {
		strs := make([]string, len(ints))
		for i, n := range ints {
			strs[i] = fmt.Sprintf("%d", n)
		}
		return strs
	}
	selectors := tftea.NewSelectPanel().
		WithLabel("Compress Options").
		WithSelect("Archive Type", m.archiveTypes).
		WithSelect("Compress Level", compressLevels).
		WithSelect("Solid Mode", boolValues).
		WithSelect("Dictionary Size", toStringSlice(m.dictionarySizes)).
		WithSelect("Thread Number", toStringSlice(m.threadNumbers)).
		WithSelect("Move to Archive", boolValues).
		WithSelect("Normalize Attrs", boolValues).
		WithSelect("Separate Inputs", boolValues).
		WithSelected(m.selectedIndices).
		WithHotkey(true).
		WithSelectWidth(20).
		WithOptionWidth(9)
	selected, err := selectors.Run()
	if err != nil {
		return nil, err
	}
	return &CompressOptionsValue{
		ArchiveType:    m.archiveTypes[selected[0]],
		CompressLevel:  compression.Level(selected[1]),
		SolidMode:      selected[2] == 1,
		DictionarySize: m.dictionarySizes[selected[3]],
		ThreadNumber:   m.threadNumbers[selected[4]],
		MoveToArchive:  selected[5] == 1,
		NormalizeAttrs: selected[6] == 1,
		SeparateInputs: selected[7] == 1,
	}, nil
}
