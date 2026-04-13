package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SafeFilePath validates and returns a safe file path within destDir
// Returns error if the path would escape destDir (path traversal attack)
func SafeFilePath(name, destDir string) (string, error) {
	if strings.Contains(name, "..") {
		return "", fmt.Errorf("illegal file name: %s", name)
	}

	cleanName := filepath.Clean(name)
	if filepath.IsAbs(cleanName) || strings.HasPrefix(cleanName, "..") {
		return "", fmt.Errorf("illegal file name: %s", name)
	}

	fpath := filepath.Join(destDir, cleanName)
	if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return "", fmt.Errorf("illegal file path: %s", fpath)
	}

	return fpath, nil
}

// SafeFilePathOrSkip returns empty string instead of error for invalid paths
func SafeFilePathOrSkip(name, destDir string) string {
	fpath, err := SafeFilePath(name, destDir)
	if err != nil {
		return ""
	}
	return fpath
}

// ExtractZipFile extracts a single file from a zip archive to destDir
func ExtractZipFile(f *zip.File, destDir string) error {
	fpath, err := SafeFilePath(f.Name, destDir)
	if err != nil {
		return nil // Skip invalid paths silently
	}

	if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
		return err
	}

	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	_, err = io.Copy(outFile, rc)
	return err
}

// ExtractZipToDir extracts all files from a zip reader to destDir
func ExtractZipToDir(zr *zip.Reader, destDir string) error {
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if err := ExtractZipFile(f, destDir); err != nil {
			return err
		}
	}
	return nil
}
