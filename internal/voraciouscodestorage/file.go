package voraciouscodestorage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

type File struct {
	LatestVersion int
	Name          string
	FullPath      string
}

func (f *File) AddNewVersion(r io.Reader) error {
	latestVersion := f.LatestVersion + 1
	versionPath := filePathWithVersion(f.FullPath, latestVersion)
	file, err := os.Create(versionPath)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	_, err = io.Copy(writer, r)
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	f.LatestVersion = latestVersion
	return nil
}

func AddNewFile(r io.Reader, fullPath string) (*File, error) {
	fileName, filePath := splitFilePath(fullPath)
	latestVersion := 1
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("path to folder %s does not exist", filePath)
	}
	versionPath := filePathWithVersion(fullPath, latestVersion)
	file, err := os.Create(versionPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = io.Copy(writer, r)
	if err != nil {
		return nil, err
	}
	err = writer.Flush()
	if err != nil {
		return nil, err
	}
	return &File{
		LatestVersion: latestVersion,
		Name:          fileName,
		FullPath:      fullPath,
	}, nil
}

func TrackExistingFile(fullPath string) (*File, error) {
	exePath, _ := os.Executable()
	fmt.Println("Executable Path:", exePath)
	latestVersion := 0
	for {
		nextVersion := latestVersion + 1
		versionPath := filePathWithVersion(fullPath, nextVersion)
		_, err := os.Stat(versionPath)
		if os.IsNotExist(err) && nextVersion == 1 {
			return nil, fmt.Errorf("file %s does not exist", fullPath)
		}
		if _, err := os.Stat(versionPath); os.IsNotExist(err) {
			break
		}
		latestVersion++
	}
	_, fileName := filepath.Split(fullPath)
	return &File{
		LatestVersion: latestVersion,
		Name:          fileName,
		FullPath:      fullPath,
	}, nil
}

func (f *File) ReadLatestFile(w io.Writer) error {
	return f.ReadFile(f.LatestVersion, w)
}

func (f *File) ReadFile(version int, w io.Writer) error {
	if version < 0 || version > f.LatestVersion {
		return fmt.Errorf("invalid version %d for file %s", version, f.FullPath)
	}
	fullPath := filePathWithVersion(f.FullPath, version)
	file, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	_, err = io.Copy(w, reader)

	return err
}

func (f *File) Remove() error {
	// Remove all versions of the file
	for i := 1; i <= f.LatestVersion; i++ {
		versionPath := filePathWithVersion(f.FullPath, i)
		err := os.Remove(versionPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func filePathWithVersion(fullPath string, version int) string {
	versionString := strconv.Itoa(version)
	return fullPath + "." + versionString
}

func splitFilePath(path string) (string, string) {
	// Split the path into directory and fileName name
	dir := filepath.Dir(path)
	fileName := filepath.Base(path)
	return fileName, dir
}
