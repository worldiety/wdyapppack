// Package bundle is the bounded context that creates macOS application
// bundles: it turns a prebuilt program, an image and some metadata into a
// .app directory, and prepares those inputs — icons, universal programs, the
// configuration file.
//
// It knows nothing about how files are stored or how it is called. It reads
// and writes through the Files port and reports problems as messages that say
// what is wrong, why and how to fix it.
package bundle

import (
	"cmp"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/worldiety/wdyapppack/foundation/problem"
)

// Config is the configuration file: everything a build needs, in the form the
// user writes it.
//
// The sections sign, notarize and dmg belong to the distribution commands.
// They are carried here unread, so that one file describes the whole way from
// program to download, and so that a field unknown to every command is still
// refused.
type Config struct {
	Schema            string            `json:"$schema,omitempty"`
	Name              string            `json:"name,omitempty"`
	DisplayName       string            `json:"displayName,omitempty"`
	Identifier        string            `json:"identifier,omitempty"`
	Version           string            `json:"version,omitempty"`
	BuildNumber       string            `json:"buildNumber,omitempty"`
	Binaries          []Binary          `json:"binaries,omitempty"`
	ExecutableName    string            `json:"executableName,omitempty"`
	Icon              *Icon             `json:"icon,omitempty"`
	MinimumMacOS      string            `json:"minimumMacOS,omitempty"`
	Category          string            `json:"category,omitempty"`
	Copyright         string            `json:"copyright,omitempty"`
	Agent             bool              `json:"agent,omitempty"`
	BackgroundOnly    bool              `json:"backgroundOnly,omitempty"`
	Resources         []Resource        `json:"resources,omitempty"`
	Frameworks        []string          `json:"frameworks,omitempty"`
	Helpers           []Helper          `json:"helpers,omitempty"`
	URLSchemes        []string          `json:"urlSchemes,omitempty"`
	UsageDescriptions map[string]string `json:"usageDescriptions,omitempty"`
	Environment       map[string]string `json:"environment,omitempty"`
	Plist             map[string]any    `json:"plist,omitempty"`
	PlistFile         string            `json:"plistFile,omitempty"`
	Out               string            `json:"out,omitempty"`

	Sign     json.RawMessage `json:"sign,omitempty"`
	Notarize json.RawMessage `json:"notarize,omitempty"`
	DMG      json.RawMessage `json:"dmg,omitempty"`
}

// BundlePath is where build writes the bundle this configuration describes:
// <out>/<name>.app.
func (c Config) BundlePath() string {
	return filepath.Join(cmp.Or(c.Out, "dist"), c.Name+".app")
}

// Binary is one program to package.
type Binary struct {
	// Arch states the architecture the program is for. Optional: it is read
	// from the program and only checked against it when given.
	Arch string `json:"arch,omitempty"`
	Path string `json:"path"`
}

// Helper is a further program copied into Contents/Helpers: one program, or
// one per architecture, merged into a universal one like the main program.
type Helper struct {
	// Name is the file name inside Contents/Helpers. When empty, it is the
	// file name of the first program without an architecture suffix.
	Name     string   `json:"name,omitempty"`
	Binaries []Binary `json:"binaries"`
}

// FileName is the name the helper is written under in Contents/Helpers.
func (h Helper) FileName() string {
	if h.Name != "" || len(h.Binaries) == 0 {
		return h.Name
	}
	return trimArchSuffix(filepath.Base(h.Binaries[0].Path))
}

// Icon names the image the icon is made from.
type Icon struct {
	Path string `json:"path"`
	// Shape is "none" or "squircle".
	Shape string `json:"shape,omitempty"`
}

// Resource is a file or directory copied into Contents/Resources.
type Resource struct {
	From string `json:"from"`
	// To is the path below Contents/Resources; the name of From when empty.
	To string `json:"to,omitempty"`
}

// Node is one entry of a tree to write.
type Node struct {
	// Path is relative to the root of the tree, with forward slashes.
	Path string
	// Dir marks a directory.
	Dir bool
	// Mode is the permission of a file or directory.
	Mode fs.FileMode
	// Data is the content of a file, unless From is set.
	Data []byte
	// From names a file whose content is copied, so that large files need not
	// be held in memory.
	From string
	// Link is the target of a symbolic link.
	Link string
}

// FileInfo describes a path.
type FileInfo struct {
	// Path is the path as asked for, or relative to the listed directory.
	Path   string
	Exists bool
	Dir    bool
	Link   string
	Mode   fs.FileMode
	Size   int64
}

// WriteOptions control how something is written.
type WriteOptions struct {
	// Force allows replacing what exists.
	Force bool
	// ModTime, when not zero, is given to every file and directory written.
	ModTime time.Time
}

// Report is what every use case of this context hands back besides its
// result: the notes worth knowing.
type Report struct {
	Notes []problem.Message `json:"notes,omitempty"`
}
