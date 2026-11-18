package bundle

import (
	"fmt"
	"slices"

	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/pkg/macho"
)

// program is a program read and recognised.
type program struct {
	path string
	data []byte
	info macho.Info
}

// readInput reads a file named as input, reporting a missing or unreadable one.
func readInput(files Files, path, role string, c *problem.Collector) ([]byte, bool) {
	st, err := files.Stat(path)
	if err == nil && !st.Exists {
		c.Error(problem.Failed, msg("WAP-FILE-NOT-FOUND", fmt.Sprintf("The %s %s does not exist.", role, path)))
		return nil, false
	}
	if err == nil && st.Dir {
		c.Error(problem.Failed, msg("WAP-FILE-UNREADABLE", fmt.Sprintf("The %s %s is a directory, not a file.", role, path)))
		return nil, false
	}
	data, err := files.Read(path)
	if err != nil {
		c.Error(problem.Failed, msg("WAP-FILE-UNREADABLE", fmt.Sprintf("The %s %s cannot be read: %v.", role, path, err)))
		return nil, false
	}
	return data, true
}

// loadProgram reads a program and checks that it is an executable for macOS,
// for the architecture stated if one is.
func loadProgram(files Files, b Binary, role string, c *problem.Collector) (program, bool) {
	stated := ""
	if b.Arch != "" {
		a, ok := macho.NormalizeArch(b.Arch)
		if !ok {
			c.Error(problem.Usage, msg("WAP-ARCH-UNKNOWN", fmt.Sprintf("%q, stated for %s, is not an architecture wdyapppack knows.", b.Arch, b.Path)))
			return program{}, false
		}
		stated = a
	}

	data, ok := readInput(files, b.Path, role, c)
	if !ok {
		return program{}, false
	}
	info, err := macho.Inspect(data)
	if err != nil || !info.IsMachO() {
		c.Error(problem.Failed, notMachO(b.Path, info.Kind, err))
		return program{}, false
	}

	good := true
	for _, s := range info.Slices {
		if s.FileType != "executable" {
			c.Error(problem.Failed, msg("WAP-BINARY-NOT-EXECUTABLE",
				fmt.Sprintf("The %s %s is a Mach-O %s for %s, not an executable.", role, b.Path, s.FileType, s.Arch)))
			good = false
		}
	}
	if stated != "" && !slices.Contains(info.Archs(), stated) {
		c.Error(problem.Usage, msg("WAP-ARCH-MISMATCH",
			fmt.Sprintf("%s is stated to be for %s but is built for %s.", b.Path, stated, joinArchs(info.Archs()))))
		good = false
	}
	return program{path: b.Path, data: data, info: info}, good
}

func notMachO(path string, kind macho.Kind, err error) problem.Message {
	switch kind {
	case macho.ELF:
		return msg("WAP-BINARY-NOT-MACHO", fmt.Sprintf("%s is a Linux program (ELF), not a macOS program.", path))
	case macho.PE:
		return msg("WAP-BINARY-NOT-MACHO", fmt.Sprintf("%s is a Windows program (PE), not a macOS program.", path))
	case macho.Script:
		return msgHow("WAP-BINARY-NOT-MACHO", fmt.Sprintf("%s is a script, not a compiled macOS program.", path),
			"Package the compiled program. A script cannot be signed and notarized like a program, so wdyapppack does not package one.")
	}
	if err != nil {
		return msg("WAP-BINARY-NOT-MACHO", fmt.Sprintf("%s looks like a Mach-O file but cannot be read as one: %v.", path, err))
	}
	return msg("WAP-BINARY-NOT-MACHO", fmt.Sprintf("%s is not a program wdyapppack recognises.", path))
}

func joinArchs(a []string) string {
	if len(a) == 0 {
		return "no known architecture"
	}
	out := a[0]
	for i := 1; i < len(a); i++ {
		if i == len(a)-1 {
			out += " and " + a[i]
		} else {
			out += ", " + a[i]
		}
	}
	return out
}

// programNotes records what does not prevent a bundle but will bite later.
func programNotes(p program, c *problem.Collector) {
	reported := map[string]bool{}
	for _, s := range p.info.Slices {
		if (s.Arch == "arm64" || s.Arch == "arm64e") && !s.Signed {
			c.Note(msg("WAP-ARM64-UNSIGNED", fmt.Sprintf("%s for %s carries no signature, so macOS will refuse to start it.", p.path, s.Arch)))
		}
		for _, lib := range s.Libraries {
			if macho.SystemLibrary(lib) || reported[lib] {
				continue
			}
			reported[lib] = true
			c.Note(msg("WAP-DYLIB-EXTERNAL", fmt.Sprintf("%s loads %s, which is not part of macOS.", p.path, lib)))
		}
	}
}

// assemble turns the programs into the one executable of the bundle: a single
// program as it is, several merged into a universal one.
func assemble(progs []program, c *problem.Collector) (data []byte, archs []string, minOS string, ok bool) {
	for _, p := range progs {
		for _, s := range p.info.Slices {
			if minOS == "" || compareVersions(s.MinOS, minOS) > 0 {
				minOS = s.MinOS
			}
		}
	}
	if len(progs) == 1 {
		return progs[0].data, progs[0].info.Archs(), minOS, true
	}

	var parts [][]byte
	from := map[string]string{}
	for _, p := range progs {
		thin, err := macho.Split(p.data)
		if err != nil {
			c.Error(problem.Failed, notMachO(p.path, p.info.Kind, err))
			return nil, nil, "", false
		}
		for i, s := range p.info.Slices {
			if prev, dup := from[s.Arch]; dup {
				c.Error(problem.Usage, msg("WAP-ARCH-DUPLICATE",
					fmt.Sprintf("Both %s and %s are built for %s.", prev, p.path, s.Arch)))
				return nil, nil, "", false
			}
			from[s.Arch] = p.path
			parts = append(parts, thin[i])
			archs = append(archs, s.Arch)
		}
	}
	merged, err := macho.Merge(parts)
	if err != nil {
		c.Error(problem.Failed, msg("WAP-BINARY-NOT-MACHO", fmt.Sprintf("The programs cannot be merged: %v.", err)))
		return nil, nil, "", false
	}
	slices.Sort(archs)
	return merged, archs, minOS, true
}
