package macho

import (
	"bytes"
	"debug/macho"
	"testing"

	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/pkg/macho/machotest"
	"github.com/worldiety/wdyapppack/requirements/fun/binary"
)

func TestInspectThin(t *testing.T) {
	data := machotest.Program(machotest.Options{Arch: "arm64", MinOS: "12.3", Signed: true,
		Libraries: []string{"/usr/lib/libSystem.B.dylib", "/opt/homebrew/lib/libfoo.dylib"}})
	info, err := Inspect(data)
	if err != nil {
		t.Fatal(err)
	}
	s := info.Slices[0]
	if info.Kind != Thin || s.Arch != "arm64" || s.MinOS != "12.3" || !s.Signed || len(s.Libraries) != 2 || s.FileType != "executable" {
		t.Fatalf("%+v", info)
	}
}

func TestDetect(t *testing.T) {
	cases := map[Kind][]byte{
		ELF:     machotest.ELF(),
		PE:      machotest.PE(),
		Script:  []byte("#!/bin/sh\necho"),
		Unknown: []byte("hello world"),
		Thin:    machotest.Program(machotest.Options{Arch: "x86_64"}),
	}
	for want, data := range cases {
		if got := Detect(data); got != want {
			t.Errorf("%s detected as %s", want, got)
		}
	}
}

func TestMergeKeepsEverySliceByteForByte(t *testing.T) {
	arm := machotest.Program(machotest.Options{Arch: "arm64", MinOS: "11.0", Signed: true})
	x86 := machotest.Program(machotest.Options{Arch: "x86_64", MinOS: "10.15"})
	fat, err := Merge([][]byte{arm, x86})
	if err != nil {
		t.Fatal(err)
	}

	// The standard library must read it back.
	ff, err := macho.NewFatFile(bytes.NewReader(fat))
	if err != nil {
		t.Fatal(err)
	}
	if len(ff.Arches) != 2 {
		t.Fatalf("%d arches", len(ff.Arches))
	}
	for _, a := range ff.Arches {
		if a.Offset%(1<<a.Align) != 0 {
			t.Errorf("slice %v at %d is not aligned to 2^%d", a.Cpu, a.Offset, a.Align)
		}
	}

	parts, err := Split(fat)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(parts[0], x86) || !bytes.Equal(parts[1], arm) {
		t.Fatal("a slice was changed while merging, which would break its signature")
	}

	info, _ := Inspect(fat)
	if info.Kind != Universal || len(info.Archs()) != 2 || !info.Slices[1].Signed {
		t.Fatalf("%+v", info)
	}

	if _, err := Merge([][]byte{arm, arm}); err == nil {
		t.Fatal("the same architecture was accepted twice")
	}
	if _, err := Merge([][]byte{arm, machotest.ELF()}); err == nil {
		t.Fatal("a Linux program was merged")
	}
	spec.Verified(t, binary.RBinaryUniversal)
}

func TestNormalizeArch(t *testing.T) {
	for in, want := range map[string]string{"amd64": "x86_64", "AMD64": "x86_64", "aarch64": "arm64", "arm64": "arm64"} {
		if got, ok := NormalizeArch(in); !ok || got != want {
			t.Errorf("%s → %s", in, got)
		}
	}
	if _, ok := NormalizeArch("ppc"); ok {
		t.Error("ppc accepted")
	}
}

func TestSystemLibrary(t *testing.T) {
	for p, want := range map[string]bool{
		"/usr/lib/libSystem.B.dylib":                       true,
		"/System/Library/Frameworks/Cocoa.framework/Cocoa": true,
		"@rpath/libfoo.dylib":                              true,
		"/opt/homebrew/lib/libfoo.dylib":                   false,
		"/usr/local/lib/libbar.dylib":                      false,
	} {
		if SystemLibrary(p) != want {
			t.Errorf("%s: want %v", p, want)
		}
	}
}
