package project

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"path"
	"regexp"
	"strings"
)

var scpLikeURLPattern = regexp.MustCompile(`^(?:[^@]+@)?([^:]+):/?(.+)$`)

func HashIdentifier(identifier string) string {
	sum := sha256.Sum256([]byte(identifier))
	return hex.EncodeToString(sum[:])[:16]
}

func CanonicalizeIdentifier(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if canonical, ok := canonicalizeURL(raw); ok {
		return canonical
	}
	if canonical, ok := canonicalizeSCPLikeURL(raw); ok {
		return canonical
	}
	return strings.TrimSuffix(raw, ".git")
}

func canonicalizeURL(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", false
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.User = nil
	u.Path = strings.TrimSuffix(u.Path, ".git")
	u.Path = path.Clean(u.Path)
	if u.Path == "." {
		u.Path = ""
	}
	u.Path = normalizeRepoPath(u.Host, u.Path)
	return u.String(), true
}

func canonicalizeSCPLikeURL(raw string) (string, bool) {
	match := scpLikeURLPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if len(match) != 3 {
		return "", false
	}
	host := strings.ToLower(match[1])
	repoPath := strings.TrimPrefix(match[2], "/")
	repoPath = strings.TrimSuffix(repoPath, ".git")
	repoPath = path.Clean(repoPath)
	repoPath = strings.TrimPrefix(normalizeRepoPath(host, "/"+repoPath), "/")
	if repoPath == "." || repoPath == "" {
		return host, true
	}
	return "https://" + host + "/" + repoPath, true
}

func normalizeRepoPath(host string, repoPath string) string {
	if host == "github.com" {
		return strings.ToLower(repoPath)
	}
	return repoPath
}
