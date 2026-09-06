package agent

import (
	"path/filepath"
	"regexp"
	"strings"
)

func sensitiveName(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	ext := strings.ToLower(filepath.Ext(base))
	switch ext {
	case ".pem", ".key", ".p12", ".pfx", ".crt", ".cer", ".cert", ".csr", ".der", ".p7b", ".p7c", ".p8", ".keystore", ".jks", ".kdbx", ".sqlite", ".sqlite3", ".db":
		return true
	}
	for _, part := range strings.Split(strings.ToLower(filepath.ToSlash(path)), "/") {
		if part == ".ssh" || part == ".aws" || part == ".azure" || part == ".gnupg" {
			return true
		}
	}
	for _, word := range []string{"credential", "secret", "private_key", "private-key", "id_rsa", "id_ed25519", "id_ecdsa", ".netrc", ".npmrc", ".pypirc", ".git-credentials", "token"} {
		if strings.Contains(base, word) {
			return true
		}
	}
	return false
}

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`-----BEGIN (?:[A-Z ]*PRIVATE KEY(?: BLOCK)?|CERTIFICATE)-----`),
	regexp.MustCompile(`(?i)(?:password|passwd|secret|api[_-]?key|access[_-]?key|auth[_-]?token|bearer|credential)["']?\s*[=:]\s*["']?[^\s"'{}$][^\r\n]{3,}`),
	regexp.MustCompile(`\b(?:AKIA|ASIA)[A-Z0-9]{16}\b|\bgh[pousr]_[A-Za-z0-9]{20,}\b|\bgithub_pat_[A-Za-z0-9_]{20,}\b|\bsk-[A-Za-z0-9_-]{20,}\b`),
	regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{12,}\b|\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b`),
	regexp.MustCompile(`(?i)(?:postgres(?:ql)?|mysql|mongodb(?:\+srv)?|redis|https?)://[^\s/@:]+:[^\s/@]+@`),
}

func suspiciousContent(data []byte) bool {
	for _, p := range secretPatterns {
		if p.Match(data) {
			return true
		}
	}
	return false
}
