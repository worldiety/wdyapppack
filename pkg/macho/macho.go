// Package macho tells what kind of program a file is and merges programs for
// different processors into one universal program.
//
// Reading is done by the standard library's debug/macho. What it leaves out —
// the minimum macOS version, whether a signature is present — sits in load
// commands it hands over as raw bytes, and is decoded here. Writing a
// universal ("fat") file is a header and a table followed by the programs at
// aligned offsets; it needs nothing beyond encoding/binary.
package macho

import (
	"bytes"
	"cmp"
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Kind is what a file turned out to be.
type Kind string

const (
	Thin      Kind = "mach-o"
	Universal Kind = "universal mach-o"
	ELF       Kind = "elf"
	PE        Kind = "pe"
	Script    Kind = "script"
	Unknown   Kind = "unknown"
)

// Slice is one program inside a file: the whole file for a thin Mach-O, one
// architecture of a universal one.
type Slice struct {
	// Arch is the conventional name: arm64, x86_64, arm64e, i386.
	Arch     string `json:"arch"`
	CPU      uint32 `json:"-"`
	SubCPU   uint32 `json:"-"`
	Offset   int64  `json:"-"`
	Size     int64  `json:"size"`
	FileType string `json:"fileType"`
	// MinOS is the oldest macOS the program was built for, "" if unstated.
	MinOS string `json:"minimumMacOS,omitempty"`
	SDK   string `json:"sdk,omitempty"`
	// Signed reports whether a code signature is present at all. Whether it
	// is valid is codesign's question, not this package's.
	Signed bool `json:"signed"`
	// Libraries are the dynamic libraries the program loads.
	Libraries []string `json:"libraries,omitempty"`
}

// Info describes a file.
type Info struct {
	Kind   Kind    `json:"kind"`
	Slices []Slice `json:"slices,omitempty"`
}

// Archs lists the architectures of the slices.
func (i Info) Archs() []string {
	out := make([]string, 0, len(i.Slices))
	for _, s := range i.Slices {
		out = append(out, s.Arch)
	}
	return out
}

// IsMachO reports whether the file is a macOS program of either form.
func (i Info) IsMachO() bool { return i.Kind == Thin || i.Kind == Universal }

const (
	magic64   = 0xfeedfacf
	magic32   = 0xfeedface
	magicFat  = 0xcafebabe
	lcBuildV  = 0x32
	lcMinMac  = 0x24
	lcCodeSig = 0x1d
)

// Detect classifies a file by its first bytes without parsing it.
func Detect(data []byte) Kind {
	if len(data) < 4 {
		return Unknown
	}
	le := binary.LittleEndian.Uint32(data)
	be := binary.BigEndian.Uint32(data)
	switch {
	case le == magic64 || le == magic32 || be == magic64 || be == magic32:
		return Thin
	case be == magicFat:
		// Java class files share this magic; their second word is a version
		// far above any plausible number of architectures.
		if len(data) >= 8 && binary.BigEndian.Uint32(data[4:]) < 32 {
			return Universal
		}
	case bytes.HasPrefix(data, []byte("\x7fELF")):
		return ELF
	case bytes.HasPrefix(data, []byte("MZ")):
		return PE
	case bytes.HasPrefix(data, []byte("#!")):
		return Script
	}
	return Unknown
}

// Inspect describes a file.
func Inspect(data []byte) (Info, error) {
	kind := Detect(data)
	switch kind {
	case Thin:
		s, err := inspectThin(data)
		if err != nil {
			return Info{Kind: kind}, err
		}
		s.Size = int64(len(data))
		return Info{Kind: kind, Slices: []Slice{s}}, nil
	case Universal:
		ff, err := macho.NewFatFile(bytes.NewReader(data))
		if err != nil {
			return Info{Kind: kind}, fmt.Errorf("universal header: %w", err)
		}
		defer ff.Close()
		info := Info{Kind: kind}
		for _, a := range ff.Arches {
			end := int64(a.Offset) + int64(a.Size)
			if end > int64(len(data)) {
				return info, errors.New("a slice extends beyond the end of the file")
			}
			s, err := inspectThin(data[a.Offset:end])
			if err != nil {
				return info, err
			}
			s.Offset, s.Size = int64(a.Offset), int64(a.Size)
			info.Slices = append(info.Slices, s)
		}
		return info, nil
	}
	return Info{Kind: kind}, nil
}

func inspectThin(data []byte) (Slice, error) {
	f, err := macho.NewFile(bytes.NewReader(data))
	if err != nil {
		return Slice{}, fmt.Errorf("mach-o header: %w", err)
	}
	defer f.Close()

	s := Slice{
		Arch:     ArchName(uint32(f.Cpu), f.SubCpu),
		CPU:      uint32(f.Cpu),
		SubCPU:   f.SubCpu,
		FileType: fileType(f.Type),
	}
	for _, l := range f.Loads {
		raw := l.Raw()
		if len(raw) < 8 {
			continue
		}
		switch f.ByteOrder.Uint32(raw) {
		case lcBuildV:
			if len(raw) >= 20 {
				s.MinOS = version(f.ByteOrder.Uint32(raw[12:]))
				s.SDK = version(f.ByteOrder.Uint32(raw[16:]))
			}
		case lcMinMac:
			if len(raw) >= 16 && s.MinOS == "" {
				s.MinOS = version(f.ByteOrder.Uint32(raw[8:]))
				s.SDK = version(f.ByteOrder.Uint32(raw[12:]))
			}
		case lcCodeSig:
			s.Signed = true
		}
	}
	libs, err := f.ImportedLibraries()
	if err == nil {
		s.Libraries = libs
	}
	return s, nil
}

// version decodes the xxxx.yy.zz nibble form of a load command.
func version(v uint32) string {
	major, minor, patch := v>>16, (v>>8)&0xff, v&0xff
	if patch != 0 {
		return fmt.Sprintf("%d.%d.%d", major, minor, patch)
	}
	return fmt.Sprintf("%d.%d", major, minor)
}

func fileType(t macho.Type) string {
	switch t {
	case macho.TypeExec:
		return "executable"
	case macho.TypeDylib:
		return "dynamic library"
	case macho.TypeBundle:
		return "loadable bundle"
	case macho.TypeObj:
		return "object file"
	}
	return fmt.Sprintf("type %d", t)
}

// CPU types.
const (
	CPUX86_64 = 0x01000007
	CPUARM64  = 0x0100000c
	CPUI386   = 0x00000007
)

// ArchName returns the conventional name of a processor type.
func ArchName(cpu, sub uint32) string {
	switch cpu {
	case CPUARM64:
		if sub&0xff == 2 {
			return "arm64e"
		}
		return "arm64"
	case CPUX86_64:
		return "x86_64"
	case CPUI386:
		return "i386"
	}
	return fmt.Sprintf("cpu-%#x", cpu)
}

// NormalizeArch accepts the spellings people use for an architecture and
// returns the conventional one: amd64 is x86_64, aarch64 is arm64.
func NormalizeArch(s string) (string, bool) {
	switch strings.ToLower(s) {
	case "arm64", "aarch64":
		return "arm64", true
	case "amd64", "x86_64", "x86-64", "intel":
		return "x86_64", true
	case "arm64e":
		return "arm64e", true
	}
	return "", false
}

// SystemLibrary reports whether a library ships with macOS or is reached
// relative to the program, which is how a bundle carries its own. Anything
// else — /opt/homebrew, /usr/local — exists only on machines that installed it.
func SystemLibrary(path string) bool {
	for _, p := range []string{"/usr/lib/", "/System/", "@rpath/", "@executable_path/", "@loader_path/"} {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// Split returns the thin programs of a file: the file itself when it is thin,
// its slices when it is universal.
func Split(data []byte) ([][]byte, error) {
	info, err := Inspect(data)
	if err != nil {
		return nil, err
	}
	switch info.Kind {
	case Thin:
		return [][]byte{data}, nil
	case Universal:
		out := make([][]byte, 0, len(info.Slices))
		for _, s := range info.Slices {
			out = append(out, data[s.Offset:s.Offset+s.Size])
		}
		return out, nil
	}
	return nil, fmt.Errorf("not a Mach-O program but %s", info.Kind)
}

// alignment is the power of two a slice is placed at: the page size of the
// processor, as lipo does.
func alignment(cpu uint32) uint32 {
	if cpu == CPUARM64 {
		return 14
	}
	return 12
}

// Merge combines thin programs for different architectures into one universal
// program. The programs are copied byte for byte, so their signatures stay
// valid.
func Merge(programs [][]byte) ([]byte, error) {
	if len(programs) < 2 {
		return nil, errors.New("a universal program needs at least two programs")
	}
	type part struct {
		data     []byte
		cpu, sub uint32
		align    uint32
		arch     string
	}
	parts := make([]part, 0, len(programs))
	seen := map[string]bool{}
	for i, p := range programs {
		if Detect(p) != Thin {
			return nil, fmt.Errorf("program %d is not a thin Mach-O program", i+1)
		}
		s, err := inspectThin(p)
		if err != nil {
			return nil, fmt.Errorf("program %d: %w", i+1, err)
		}
		if seen[s.Arch] {
			return nil, fmt.Errorf("architecture %s is given twice", s.Arch)
		}
		seen[s.Arch] = true
		parts = append(parts, part{data: p, cpu: s.CPU, sub: s.SubCPU, align: alignment(s.CPU), arch: s.Arch})
	}
	slices.SortFunc(parts, func(a, b part) int {
		return cmp.Or(cmp.Compare(a.cpu, b.cpu), cmp.Compare(a.sub, b.sub))
	})

	var out bytes.Buffer
	head := make([]byte, 8+20*len(parts))
	binary.BigEndian.PutUint32(head, magicFat)
	binary.BigEndian.PutUint32(head[4:], uint32(len(parts)))
	offset := uint64(len(head))
	offsets := make([]uint64, len(parts))
	for i, p := range parts {
		a := uint64(1) << p.align
		offset = (offset + a - 1) &^ (a - 1)
		offsets[i] = offset
		if offset+uint64(len(p.data)) > 1<<32-1 {
			return nil, errors.New("the programs are too large for a 32 bit universal header")
		}
		e := head[8+20*i:]
		binary.BigEndian.PutUint32(e, p.cpu)
		binary.BigEndian.PutUint32(e[4:], p.sub)
		binary.BigEndian.PutUint32(e[8:], uint32(offset))
		binary.BigEndian.PutUint32(e[12:], uint32(len(p.data)))
		binary.BigEndian.PutUint32(e[16:], p.align)
		offset += uint64(len(p.data))
	}
	out.Write(head)
	for i, p := range parts {
		out.Write(make([]byte, int(offsets[i])-out.Len()))
		out.Write(p.data)
	}
	return out.Bytes(), nil
}
