package voraciouscodestorage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type VersionedFile struct {
	Mutex         sync.RWMutex
	LatestVersion int
	AbsolutePath  string
	FileName      string
	Files         map[int]*File
}

func NewVersionedFile(absPath string, r io.Reader, bytes int) (*VersionedFile, error) {
	versionedAbsPath := absPath + ".1"
	ext := filepath.Ext(absPath)
	file, err := NewTextFile(r, versionedAbsPath, ext, bytes)
	if err != nil {
		return nil, err
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
	vf.Mutex.Lock()
	defer vf.Mutex.Unlock()

	oldFile := vf.Files[vf.LatestVersion-1]
	versionedAbsPath := fmt.Sprintf("%s.%d", vf.AbsolutePath, vf.LatestVersion+1)
	newFile, err := NewTextFile(r, versionedAbsPath, oldFile.Ext, bytes)
	if err != nil {
		return nil, err
	}
	match, err := compareFiles(oldFile.AbsolutePath, newFile.AbsolutePath)
	if err != nil {
		return nil, fmt.Errorf("error comparing files: %v", err)
	}
	if match {
		// If the new file is identical to the old one, do not increment the version
		os.Remove(newFile.AbsolutePath) // Remove the new file since it's a duplicate
		return oldFile, nil
	}
	vf.LatestVersion++
	vf.Files[vf.LatestVersion-1] = newFile
	return newFile, nil
}

func (vf *VersionedFile) GetLatest() (*File, error) {
	vf.Mutex.RLock()
	defer vf.Mutex.RUnlock()

	if vf.LatestVersion < 1 {
		return nil, fmt.Errorf("no versions available")
	}
	return vf.Files[vf.LatestVersion-1], nil
}

func (vf *VersionedFile) GetVersion(version int) (*File, error) {
	vf.Mutex.RLock()
	defer vf.Mutex.RUnlock()

	if version < 1 || version > vf.LatestVersion {
		return nil, fmt.Errorf("invalid version number: %d", version)
	}
	return vf.Files[version-1], nil
}

func (vf *VersionedFile) Delete() error {
	vf.Mutex.Lock()
	defer vf.Mutex.Unlock()

	for i := 0; i < vf.LatestVersion; i++ {
		err := vf.Files[i].Delete()
		if err != nil {
			return fmt.Errorf("error deleting file: %v", err)
		}
	}
	return nil
}

func compareFiles(file1AbsPath string, file2AbsPath string) (bool, error) {
	stat1, err := os.Stat(file1AbsPath)
	if err != nil {
		return false, err
	}
	stat2, err := os.Stat(file2AbsPath)
	if err != nil {
		return false, err
	}
	if stat1.Size() != stat2.Size() {
		return false, nil
	}
	if os.SameFile(stat1, stat2) {
		return true, nil
	}
	file1, err := os.Open(file1AbsPath)
	if err != nil {
		return false, fmt.Errorf("error opening file %s: %v", file1AbsPath, err)
	}
	defer file1.Close()
	file2, err := os.Open(file2AbsPath)
	if err != nil {
		return false, fmt.Errorf("error opening file %s: %v", file2AbsPath, err)
	}
	defer file2.Close()
	buf1 := make([]byte, 4096)
	buf2 := make([]byte, 4096)
	for {
		// Read chunks of data from both files
		n1, err1 := file1.Read(buf1)
		n2, err2 := file2.Read(buf2)
		// Check for errors and compare the number of bytes read
		if err1 != nil && err1 != io.EOF {
			return false, fmt.Errorf("error reading file %s: %v", file1AbsPath, err1)
		}
		if err2 != nil && err2 != io.EOF {
			return false, fmt.Errorf("error reading file %s: %v", file2AbsPath, err2)
		}
		if n1 != n2 {
			return false, nil
		}
		// Compare the bytes read
		for i := 0; i < n1; i++ {
			if buf1[i] != buf2[i] {
				return false, nil
			}
		}
		// If we reached the end of both files, they are equal
		if err1 == io.EOF && err2 == io.EOF {
			break
		}
		// If we reached the end of one file but not the other, they are not equal
		if err1 == io.EOF || err2 == io.EOF {
			return false, nil
		}
	}
	return true, nil
}
