package f

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
)

// CopyIfExists copies the file from any of the roots + relPath to destRoot + relPath.
// Returns true if a file was copied.
func CopyIfExists(roots []string, destRoot, relPath string, perm os.FileMode) (bool, error) {
	for _, root := range roots {
		cand := filepath.Join(root, relPath)
		info, err := os.Stat(cand)
		if err == nil && info.Mode().IsRegular() {
			dest := filepath.Join(destRoot, relPath)
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return false, err
			}

			b, err := os.ReadFile(cand)
			if err != nil {
				return false, err
			}
			if err := os.WriteFile(dest, b, perm); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func IsEmptyDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	return len(entries) == 0
}

// PathResolve expands ~ to home directory and resolves relative paths to absolute
// Returns nil on error
func PathResolve(path *string) *string {
	if path == nil {
		return nil
	}

	p := *path
	if len(p) > 0 && p[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil
		}
		p = filepath.Join(homeDir, p[1:])
	}

	absPath, err := filepath.Abs(p)
	if err != nil {
		return nil
	}
	return &absPath
}

// PathToSha computes the SHA hash of a file
// bits can be 256 or 512
// Returns pointer to hex-encoded hash string, or nil on error
func PathToSha(path *string, bits int) *string {
	if path == nil || *path == "" {
		return nil
	}

	f, err := os.Open(*path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var h hash.Hash
	switch bits {
	case 256:
		h = sha256.New()
	case 512:
		h = sha512.New()
	default:
		return nil
	}

	if _, err := io.Copy(h, f); err != nil {
		return nil
	}

	result := hex.EncodeToString(h.Sum(nil))
	return &result
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func DirsIn(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	return dirs, nil
}

func FilesIn(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

func FilesFind(root string, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			// Ignore unreadable paths
			return nil
		}
		if info == nil || info.IsDir() {
			return nil
		}
		// Match against relative path first
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			rel = p
		}
		if ok, mErr := filepath.Match(pattern, rel); mErr == nil && ok {
			matches = append(matches, p)
			return nil
		} else if mErr != nil {
			return mErr
		}
		// Also allow matching just the basename
		base := filepath.Base(p)
		if ok, mErr := filepath.Match(pattern, base); mErr == nil && ok {
			matches = append(matches, p)
		} else if mErr != nil {
			return mErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func ClearDir(targetPath string) error {
	// Clean out target directory contents (but keep the directory itself)
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return fmt.Errorf("error listing target directory contents: %s", err)
	}
	for _, e := range entries {
		p := filepath.Join(targetPath, e.Name())
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("error removing existing entry %s: %s", p, err)
		}
	}
	return nil
}

func Cwd() string {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return path
}
