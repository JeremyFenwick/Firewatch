package voraciouscodestorage

import (
	"io"
	"os"
	"path/filepath"
)

type Folder struct {
	AbsolutePath string
	Name         string
	Files        map[string]*VersionedFile
	SubFolders   map[string]*Folder
}

func NewFolder(absPath string) (*Folder, error) {
	// Create the folder if it doesn't exist
	err := os.MkdirAll(absPath, 0755)
	if err != nil {
		return nil, err
	}
	folder := &Folder{
		AbsolutePath: absPath,
		Name:         filepath.Base(absPath),
		Files:        make(map[string]*VersionedFile),
		SubFolders:   make(map[string]*Folder),
	}
	return folder, nil
}

func (f *Folder) AddNewFile(fileName string, r io.Reader, bytes int) (*VersionedFile, error) {
	// Check if the file already exists
	if _, exists := f.Files[fileName]; exists {
		return nil, os.ErrExist
	}
	// Create the file
	file, err := NewVersionedFile(f.AbsolutePath+"/"+fileName, r, bytes)
	if err != nil {
		return nil, err
	}
	f.Files[fileName] = file
	return file, nil
}

func (f *Folder) AddNewSubFolder(folderName string) (*Folder, error) {
	fullPath := f.AbsolutePath + "/" + folderName
	err := os.MkdirAll(fullPath, 0755)
	if err != nil {
		return nil, err
	}
	subFolder := &Folder{
		AbsolutePath: fullPath,
		Files:        make(map[string]*VersionedFile),
		SubFolders:   make(map[string]*Folder),
	}
	f.SubFolders[folderName] = subFolder
	return subFolder, nil
}

func (f *Folder) GetFiles() []*VersionedFile {
	files := make([]*VersionedFile, 0, len(f.Files))
	for _, file := range f.Files {
		files = append(files, file)
	}
	return files
}

func (f *Folder) GetSubFolders() []*Folder {
	subFolders := make([]*Folder, 0, len(f.SubFolders))
	for _, folder := range f.SubFolders {
		subFolders = append(subFolders, folder)
	}
	return subFolders
}

func (f *Folder) IsEmpty() bool {
	return len(f.Files) == 0 && len(f.SubFolders) == 0
}

func (f *Folder) Delete() error {
	err := os.RemoveAll(f.AbsolutePath)
	if err != nil {
		return err
	}
	return nil
}
