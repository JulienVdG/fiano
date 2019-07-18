// Copyright 2018 the LinuxBoot Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package visitors

import (
	"errors"
	"io/ioutil"

	"github.com/linuxboot/fiano/pkg/uefi"
)

// ReplaceRaw replaces raw file content with NewRaw for all files matching Predicate.
type ReplaceRaw struct {
	// Input
	Predicate func(f uefi.Firmware) bool
	NewRaw    []byte

	// Output
	Matches []uefi.Firmware
}

// Run wraps Visit and performs some setup and teardown tasks.
func (v *ReplaceRaw) Run(f uefi.Firmware) error {
	// Run "find" to generate a list of matches to replace.
	find := Find{
		Predicate: FindAndPredicate(
			v.Predicate,
			FindFileTypePredicate(uefi.FVFileTypeRaw),
		),
	}
	if err := find.Run(f); err != nil {
		return err
	}

	// Use this list of matches for replacing sections.
	v.Matches = find.Matches
	if len(find.Matches) == 0 {
		return errors.New("no matches found for replacement")
	}
	/* TODO: do we want to limit to one match?
	if len(find.Matches) > 1 {
			return errors.New("multiple matches found! There can be only one. Use find to list all matches")
		}
	*/
	for _, m := range v.Matches {
		if err := m.Apply(v); err != nil {
			return err
		}
	}
	return nil
}

// Visit applies the Extract visitor to any Firmware type.
func (v *ReplaceRaw) Visit(f uefi.Firmware) error {
	switch f := f.(type) {

	case *uefi.File:
		if f.Header.Type == uefi.FVFileTypeRaw {
			// Rebuild the file
			fileData := v.NewRaw
			dLen := uint64(len(fileData))
			f.SetSize(uefi.FileHeaderMinLength+dLen, true)
			if err := f.ChecksumAndAssemble(fileData); err != nil {
				return err
			}
		}
		return f.ApplyChildren(v)

	default:
		// Must be applied to a File to have any effect.
		return nil
	}
}

func init() {
	RegisterCLI("replace_raw", "replace a raw file content given a GUID and new file", 2, func(args []string) (uefi.Visitor, error) {
		pred, err := FindFilePredicate(args[0])
		if err != nil {
			return nil, err
		}

		filename := args[1]
		newRaw, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		// Find all the matching files and replace their inner PE32s.
		return &ReplaceRaw{
			Predicate: pred,
			NewRaw:    newRaw,
		}, nil
	})
}
