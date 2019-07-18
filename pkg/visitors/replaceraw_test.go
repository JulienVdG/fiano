// Copyright 2018 the LinuxBoot Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package visitors

import (
	"reflect"
	"testing"

	"github.com/linuxboot/fiano/pkg/uefi"
)

func TestReplaceRaw(t *testing.T) {
	f := parseImage(t)

	// Apply the visitor.
	replace := &ReplaceRaw{
		Predicate: FindFileGUIDPredicate(*testRawGUID),
		NewRaw:    []byte("rawbanana"),
	}
	if err := replace.Run(f); err != nil {
		t.Fatal(err)
	}

	// We expect one match.
	if len(replace.Matches) != 1 {
		t.Fatalf("got %d matches; expected 1", len(replace.Matches))
	}

	// Find the section and make sure it contains the expected data.
	results := find(t, f, testRawGUID)
	if len(results) != 1 {
		t.Fatalf("got %d matches; expected 1", len(results))
	}
	want := []byte{'r', 'a', 'w', 'b', 'a', 'n', 'a', 'n', 'a'}
	file, ok := results[0].(*uefi.File)
	if !ok {
		t.Fatalf("did not match a file, got type :%T", file)
	}
	got := file.Buf()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v; got %v", want, got)
	}
}

func TestReplaceRawErrors(t *testing.T) {
	f := parseImage(t)

	var tests = []struct {
		name   string
		newRaw []byte
		match  string
		err    string
	}{
		{"No Matches", []byte("rawbanana"), "no-match-string",
			"no matches found for replacement"},
		{"No Raw Matches", []byte("rawbanana"), testGUID.String(),
			"no matches found for replacement"},
		/*		{"Multiple Matches", []byte("MZbanana"), ".*",
				"multiple matches found! There can be only one. Use find to list all matches"},*/
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Apply the visitor.
			pred, err := FindFilePredicate(test.match)
			if err != nil {
				t.Fatal(err)
			}
			replace := &ReplaceRaw{
				Predicate: pred,
				NewRaw:    test.newRaw,
			}
			err = replace.Run(f)
			if err == nil {
				t.Fatalf("Expected Error (%v), got nil", test.err)
			} else if err.Error() != test.err {
				t.Fatalf("Mismatched Error: Expected %v, got %v", test.err, err.Error())
			}
		})
	}
}
