package voraciouscodestorage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type File struct {
	Name         string
	Ext          string
	Bytes        int
	AbsolutePath string
}

func NewFile(r io.Reader, absPath, ext string, bytes int) (*File, error) {
	path := filepath.Dir(absPath)
	fileName := filepath.Base(absPath)
	// Return an error if the path does not exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %v", err)
	}
	// Create the file
	file, err := os.Create(absPath)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()
	// Write the file contents
	_, err = io.Copy(file, r)
	if err != nil {
		return nil, fmt.Errorf("error writing file contents: %v", err)
	}
	// Return the file object
	return &File{
		Name:         fileName,
		Ext:          ext,
		Bytes:        bytes,
		AbsolutePath: absPath,
	}, nil
}

func (f *File) ReadFile(w io.Writer) error {
	// Read the file
	file, err := os.Open(f.AbsolutePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Copy the file contents to the writer
	_, err = io.Copy(w, file)
	if err != nil {
		return fmt.Errorf("error copying file contents: %v", err)
	}
	return nil
}

func (f *File) Delete() error {
	// Delete the file
	err := os.Remove(f.AbsolutePath)
	if err != nil {
		return fmt.Errorf("error deleting file: %v", err)
	}
	return nil
}
