package distribution

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/worldiety/wdyapppack/foundation/problem"
)

// openKeychain prepares the keychain to sign from. Without a certificate it is
// the user's own; with one, a temporary keychain holding it. The function
// returned removes what was created and reports a failure to do so as a note.
func openKeychain(ws Workspace, signer Signer, cert *Certificate, notes *problem.Collector) (Keychain, func(), error) {
	if cert == nil || (cert.File == "" && strings.TrimSpace(cert.Base64) == "") {
		return Keychain{}, func() {}, nil
	}
	source := EnvCertificate
	if cert.File != "" {
		source = cert.File
	}
	if !cert.HasPassword {
		return Keychain{}, nil, problem.Misuse(msg("WAP-CERTIFICATE-PASSWORD",
			fmt.Sprintf("A certificate is given in %s, but no password: neither --certificate-password nor %s is set.", source, EnvCertificatePassword)))
	}
	var p12 []byte
	if cert.File != "" {
		data, err := ws.Read(cert.File)
		if err != nil {
			return Keychain{}, nil, problem.Fail(msg("WAP-CERTIFICATE-INVALID", fmt.Sprintf("%s cannot be read: %v.", cert.File, err)))
		}
		p12 = data
	} else {
		data, err := decodeBase64(cert.Base64)
		if err != nil {
			return Keychain{}, nil, problem.Misuse(msg("WAP-CERTIFICATE-INVALID",
				fmt.Sprintf("%s is not base64: %v.", EnvCertificate, err)))
		}
		p12 = data
	}
	kc, closeKC, err := signer.OpenKeychain(p12, cert.Password)
	if err != nil {
		return Keychain{}, nil, problem.Fail(msg("WAP-CERTIFICATE-INVALID",
			fmt.Sprintf("The certificate in %s cannot be imported: %v", source, strings.TrimSpace(err.Error()))))
	}
	return kc, func() {
		if err := closeKC(); err != nil {
			notes.Note(msg("WAP-KEYCHAIN-CLEANUP", fmt.Sprintf("The temporary keychain %s could not be deleted: %v", kc.Path, err)))
		}
	}, nil
}

// decodeBase64 reads base64 as CI secrets hold it: with or without padding,
// with line breaks wherever the tool that encoded it put them.
func decodeBase64(s string) ([]byte, error) {
	s = strings.Join(strings.Fields(s), "")
	if data, err := base64.StdEncoding.DecodeString(s); err == nil {
		return data, nil
	}
	return base64.RawStdEncoding.DecodeString(strings.TrimRight(s, "="))
}

// apiKeyFile turns an API key from the environment into the .p8 text
// notarytool reads.
func apiKeyText(k APIKey) ([]byte, error) {
	data := []byte(k.Data)
	if !strings.Contains(k.Data, "-----BEGIN") {
		decoded, err := decodeBase64(k.Data)
		if err != nil {
			return nil, fmt.Errorf("it is neither the text of a .p8 file nor base64: %v", err)
		}
		data = decoded
	}
	if !strings.Contains(string(data), "PRIVATE KEY") {
		return nil, fmt.Errorf("it does not contain a private key")
	}
	return data, nil
}
