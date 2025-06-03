package voraciouscodestorage

import (
	"errors"
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

var (
	ErrNonTextData = errors.New("non-text data provided")
)

func NewTextFile(r io.Reader, absPath, ext string, numBytes int) (*File, error) {
	// Data validation function
	validText := func(data []byte) bool {
		for _, b := range data {
			if (b < 32 || b > 126) && b != '\n' && b != '\r' && b != '\t' {
				return false
			}
		}
		return true
	}
	// Function to check if text is valid
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
	buff := make([]byte, 4096)
	for {
		n, err := r.Read(buff)
		if n > 0 {
			// Check if the bytes read are valid UTF-8 and do not contain null bytes
			if !validText(buff[:n]) {
				// Delete the file
				file.Close()       // Close the file before deleting
				os.Remove(absPath) // Remove the file if it contains non-text data
				return nil, ErrNonTextData
			}
			// Write the valid bytes to the file
			if _, err := file.Write(buff[:n]); err != nil {
				return nil, fmt.Errorf("error writing to file: %v", err)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading from reader: %v", err)
		}
	}
	// Return the file object
	return &File{
		Name:         fileName,
		Ext:          ext,
		Bytes:        numBytes,
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
