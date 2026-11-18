// Package machotest writes small but well formed Mach-O files for tests.
//
// They contain no code and cannot run. They carry exactly what wdyapppack
// reads: the processor, the file type, the minimum macOS version, the
// libraries loaded and whether a signature is present. That keeps the tests
// independent of a compiler and of macOS.
package machotest

import (
	"encoding/binary"
	"strconv"
	"strings"
)

// Options shape the file.
type Options struct {
	// Arch is arm64 or x86_64.
	Arch string
	// MinOS is e.g. "12.0"; "" leaves the build version out.
	MinOS string
	// Libraries are written as load commands.
	Libraries []string
	// Signed adds a code signature load command.
	Signed bool
	// Dylib makes it a dynamic library rather than an executable.
	Dylib bool
}

// Program returns a Mach-O file.
func Program(o Options) []byte {
	le := binary.LittleEndian
	cpu, sub := uint32(0x0100000c), uint32(0)
	if o.Arch == "x86_64" {
		cpu, sub = 0x01000007, 3
	}
	ftype := uint32(2)
	if o.Dylib {
		ftype = 6
	}

	var cmds [][]byte
	if o.MinOS != "" {
		c := make([]byte, 24)
		le.PutUint32(c, 0x32)
		le.PutUint32(c[4:], 24)
		le.PutUint32(c[8:], 1) // platform macOS
		le.PutUint32(c[12:], encode(o.MinOS))
		le.PutUint32(c[16:], encode("15.0"))
		cmds = append(cmds, c)
	}
	for _, lib := range o.Libraries {
		name := append([]byte(lib), 0)
		size := (24 + len(name) + 7) &^ 7
		c := make([]byte, size)
		le.PutUint32(c, 0xc)
		le.PutUint32(c[4:], uint32(size))
		le.PutUint32(c[8:], 24)
		le.PutUint32(c[16:], 0x10000)
		le.PutUint32(c[20:], 0x10000)
		copy(c[24:], name)
		cmds = append(cmds, c)
	}
	sigAt := -1
	if o.Signed {
		c := make([]byte, 16)
		le.PutUint32(c, 0x1d)
		le.PutUint32(c[4:], 16)
		sigAt = len(cmds)
		cmds = append(cmds, c)
	}

	sizeofcmds := 0
	for _, c := range cmds {
		sizeofcmds += len(c)
	}
	head := make([]byte, 32)
	le.PutUint32(head, 0xfeedfacf)
	le.PutUint32(head[4:], cpu)
	le.PutUint32(head[8:], sub)
	le.PutUint32(head[12:], ftype)
	le.PutUint32(head[16:], uint32(len(cmds)))
	le.PutUint32(head[20:], uint32(sizeofcmds))

	out := append([]byte{}, head...)
	for _, c := range cmds {
		out = append(out, c...)
	}
	if sigAt >= 0 {
		// Point the signature at a few bytes after the commands.
		pos := 32
		for i := 0; i < sigAt; i++ {
			pos += len(cmds[i])
		}
		le.PutUint32(out[pos+8:], uint32(len(out)))
		le.PutUint32(out[pos+12:], 16)
		out = append(out, make([]byte, 16)...)
	}
	// Some padding so that the file looks like it has a body.
	return append(out, make([]byte, 64)...)
}

func encode(v string) uint32 {
	parts := strings.Split(v, ".")
	var n [3]uint32
	for i := 0; i < len(parts) && i < 3; i++ {
		x, _ := strconv.Atoi(parts[i])
		n[i] = uint32(x)
	}
	return n[0]<<16 | n[1]<<8 | n[2]
}

// ELF returns the start of a Linux program, enough to be recognised.
func ELF() []byte { return append([]byte("\x7fELF\x02\x01\x01"), make([]byte, 57)...) }

// PE returns the start of a Windows program, enough to be recognised.
func PE() []byte { return append([]byte("MZ"), make([]byte, 62)...) }
