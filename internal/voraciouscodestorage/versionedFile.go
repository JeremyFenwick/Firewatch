package voraciouscodestorage

import (
	"fmt"
	"io"
	"path/filepath"
)

type VersionedFile struct {
	LatestVersion int
	AbsolutePath  string
	FileName      string
	Files         map[int]*File
}

func NewVersionedFile(absPath string, r io.Reader, bytes int) (*VersionedFile, error) {
	versionedAbsPath := absPath + ".1"
	ext := filepath.Ext(absPath)
	file, err := NewFile(r, versionedAbsPath, ext, bytes)
	if err != nil {
		return nil, fmt.Errorf("error creating versioned file: %v", err)
	}
	versionedFile := &VersionedFile{
		LatestVersion: 1,
		AbsolutePath:  absPath,
		FileName:      filepath.Base(absPath),
		Files:         make(map[int]*File),
	}
	versionedFile.Files[0] = file
	return versionedFile, nil
}

func (vf *VersionedFile) AddVersion(r io.Reader, bytes int) (*File, error) {
	ext := vf.Files[vf.LatestVersion-1].Ext
	vf.LatestVersion++
	versionedAbsPath := fmt.Sprintf("%s.%d", vf.AbsolutePath, vf.LatestVersion)
	file, err := NewFile(r, versionedAbsPath, ext, bytes)
	if err != nil {
		return nil, fmt.Errorf("error creating new version: %v", err)
	}
	vf.Files[vf.LatestVersion-1] = file
	return file, nil
}

func (vf *VersionedFile) GetLatest() (*File, error) {
	if vf.LatestVersion < 1 {
		return nil, fmt.Errorf("no versions available")
	}
	return vf.Files[vf.LatestVersion-1], nil
}

func (vf *VersionedFile) GetVersion(version int) (*File, error) {
	if version < 1 || version > vf.LatestVersion {
		return nil, fmt.Errorf("invalid version number: %d", version)
	}
	return vf.Files[version-1], nil
}

func (vf *VersionedFile) Delete() error {
	for i := 0; i < vf.LatestVersion; i++ {
		err := vf.Files[i].Delete()
		if err != nil {
			return fmt.Errorf("error deleting file: %v", err)
		}
	}
	return nil
}
