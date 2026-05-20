package project

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Project struct {
	Root       string
	Identifier string
	ID         string
}

func Detect(rootOverride string) (Project, error) {
	root, gitRoot, err := detectRoot(rootOverride)
	if err != nil {
		return Project{}, err
	}
	identifier, err := detectIdentifier(root, gitRoot)
	if err != nil {
		return Project{}, err
	}
	identifier = CanonicalizeIdentifier(identifier)
	return Project{
		Root:       root,
		Identifier: identifier,
		ID:         HashIdentifier(identifier),
	}, nil
}

func detectRoot(rootOverride string) (root string, gitRoot string, err error) {
	if strings.TrimSpace(rootOverride) != "" {
		root, err = filepath.Abs(rootOverride)
		if err != nil {
			return "", "", fmt.Errorf("resolve --root: %w", err)
		}
		root = filepath.Clean(root)
		info, err := os.Stat(root)
		if err != nil {
			return "", "", fmt.Errorf("resolve --root: %w", err)
		}
		if !info.IsDir() {
			return "", "", fmt.Errorf("resolve --root: %s is not a directory", root)
		}
		return root, gitTopLevel(root), nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("resolve current directory: %w", err)
	}
	cwd = filepath.Clean(cwd)
	gitRoot = gitTopLevel(cwd)
	if gitRoot != "" {
		return gitRoot, gitRoot, nil
	}
	return cwd, "", nil
}

func detectIdentifier(root, gitRoot string) (string, error) {
	if gitRoot != "" {
		if remote := gitOriginURL(gitRoot); remote != "" {
			return remote, nil
		}
		return gitRoot, nil
	}
	return root, nil
}

func gitTopLevel(dir string) string {
	out, err := runGit(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return ""
	}
	root, err := filepath.Abs(strings.TrimSpace(out))
	if err != nil {
		return ""
	}
	return filepath.Clean(root)
}

func gitOriginURL(dir string) string {
	out, err := runGit(dir, "config", "--get", "remote.origin.url")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return stdout.String(), nil
}
