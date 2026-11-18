// Package distribution is the bounded context that hands a bundle to other
// people: it signs it, has it notarized by Apple, and wraps it into a disk
// image or a zip archive.
//
// Signing, notarization and disk images are Apple's formats and Apple's
// service, so this context does them through Apple's own tools, behind ports.
// It decides what to ask for and in which order, checks what it can before
// anything is sent, and explains what it cannot do here.
package distribution

import (
	"fmt"
	"io/fs"
	"time"

	"github.com/worldiety/wdyapppack/foundation/problem"
)

// SignSettings is the "sign" section of the configuration file.
type SignSettings struct {
	// Identity is a Developer ID certificate's name or fingerprint, or "-"
	// for an ad hoc signature.
	Identity        string `json:"identity,omitempty"`
	Entitlements    string `json:"entitlements,omitempty"`
	HardenedRuntime *bool  `json:"hardenedRuntime,omitempty"`
}

// NotarizeSettings is the "notarize" section of the configuration file.
//
// It holds the way to reach Apple's notary service, never a password: that
// comes from the keychain or from the environment variable named here.
type NotarizeSettings struct {
	KeychainProfile string `json:"keychainProfile,omitempty"`
	Key             string `json:"key,omitempty"`
	KeyID           string `json:"keyId,omitempty"`
	Issuer          string `json:"issuer,omitempty"`
	AppleID         string `json:"appleId,omitempty"`
	TeamID          string `json:"teamId,omitempty"`
	PasswordEnv     string `json:"passwordEnv,omitempty"`
	Staple          *bool  `json:"staple,omitempty"`
}

// DMGSettings is the "dmg" section of the configuration file.
type DMGSettings struct {
	VolumeName string `json:"volumeName,omitempty"`
	// Sign signs the disk image with the identity of the sign section.
	Sign bool `json:"sign,omitempty"`
}

// Settings are the sections of the configuration file this context reads.
type Settings struct {
	Sign     SignSettings     `json:"sign"`
	Notarize NotarizeSettings `json:"notarize"`
	DMG      DMGSettings      `json:"dmg"`
}

// Identity is a code signing certificate in the keychain.
type Identity struct {
	// Hash is the SHA-1 fingerprint.
	Hash string `json:"hash"`
	Name string `json:"name"`
}

// SignOptions say how one piece of code is signed.
type SignOptions struct {
	// Identity is "-" for ad hoc, else a fingerprint.
	Identity        string
	HardenedRuntime bool
	Timestamp       bool
	Entitlements    string
}

// Signature describes the signature of a piece of code.
type Signature struct {
	Signed          bool     `json:"signed"`
	AdHoc           bool     `json:"adHoc"`
	Authority       []string `json:"authority,omitempty"`
	TeamID          string   `json:"teamId,omitempty"`
	HardenedRuntime bool     `json:"hardenedRuntime"`
	Timestamp       bool     `json:"timestamp"`
}

// Credentials reach Apple's notary service in one of three ways.
type Credentials struct {
	// KeychainProfile names credentials stored with
	// xcrun notarytool store-credentials.
	KeychainProfile string
	// Key, KeyID and Issuer name an App Store Connect API key.
	Key, KeyID, Issuer string
	// AppleID, TeamID and Password are an Apple ID with an app specific
	// password.
	AppleID, TeamID, Password string
}

// Submission is the answer of the notary service.
type Submission struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// DiskImageSpec describes a disk image to create.
type DiskImageSpec struct {
	// App is the bundle to put into it.
	App string
	// Out is the image file.
	Out string
	// Volume is the name shown when the image is opened.
	Volume string
	Force  bool
}

// FileInfo describes a path.
type FileInfo struct {
	Path   string
	Exists bool
	Dir    bool
	Mode   fs.FileMode
}

// ZipOptions control an archive.
type ZipOptions struct {
	Force   bool
	ModTime time.Time
}

// Unavailable is what an adapter reports when it cannot work here: the
// operating system is not macOS, or a tool is missing.
type Unavailable struct {
	// Tool is the program that is needed, e.g. "codesign".
	Tool string
	// OS is set when the operating system is the reason.
	OS string
	// Install says how to get the tool, when that is known.
	Install string
}

func (u *Unavailable) Error() string {
	if u.OS != "" {
		return fmt.Sprintf("%s exists only on macOS, and this is %s", u.Tool, u.OS)
	}
	return u.Tool + " was not found"
}

// Report carries the notes of a use case.
type Report struct {
	Notes []problem.Message `json:"notes,omitempty"`
}
