package voraciouscodestorage

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Folder struct {
	Name     string
	FullPath string
	Files    map[string]*File
	Children map[string]*Folder
}

func CreateNewFolder(fullPath string) (*Folder, error) {
	folderName := folderName(fullPath)
	// Check if the folder already exists
	_, err := os.Stat(fullPath)
	if err == nil {
		return nil, fmt.Errorf("folder %s already exists", fullPath)
	}
	// Create the folder
	err = os.MkdirAll(fullPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating folder %s: %v", fullPath, err)
	}
	// Create the folder object
	folder := &Folder{
		Name:     folderName,
		FullPath: fullPath,
		Files:    make(map[string]*File),
		Children: make(map[string]*Folder),
	}
	// Generate the contents of the folder
	files, folders, _ := folder.GenerateContents()
	folder.TrackFolderContents(files, folders)
	return folder, nil
}

func TrackExistingFolder(fullPath string) (*Folder, error) {
	// Check if the folder exists
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("folder %s does not exist", fullPath)
	}
	// Get the folder name and create the folder object
	folderName := folderName(fullPath)
	folder := &Folder{
		Name:     folderName,
		FullPath: fullPath,
		Files:    make(map[string]*File),
		Children: make(map[string]*Folder),
	}
	// Generate the contents of the folder
	files, folders, err := folder.GenerateContents()
	if err != nil {
		return nil, fmt.Errorf("error generating contents for folder %s: %v", fullPath, err)
	}
	// Generate all children
	folder.TrackFolderContents(files, folders)
	// Return the result
	return folder, nil
}

func (f *Folder) GetChildAllFiles() []*File {
	files := make([]*File, 0)
	for _, file := range f.Files {
		files = append(files, file)
	}
	return files
}

func (f *Folder) GetChildAllFolders() []*Folder {
	folders := make([]*Folder, 0)
	for _, folder := range f.Children {
		folders = append(folders, folder)
	}
	return folders
}

func (f *Folder) Remove() error {
	// Check if the folder exists
	_, err := os.Stat(f.FullPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("folder %s does not exist", f.FullPath)
	}
	// Remove the folder
	err = os.RemoveAll(f.FullPath)
	if err != nil {
		return fmt.Errorf("error removing folder %s: %v", f.FullPath, err)
	}
	return nil
}

func (f *Folder) HasChildFolder(folderName string) bool {
	_, exists := f.Children[folderName]
	return exists
}

func (f *Folder) HasChildFile(fileName string) bool {
	_, exists := f.Files[fileName]
	return exists
}

func (f *Folder) TrackFolderContents(files, folders []string) {
	// Generate all child files
	for _, file := range files {
		filePath := filepath.Join(f.FullPath, file)
		newFile, err := TrackExistingFile(filePath)
		if err != nil {
			log.Printf("Error tracking file %s: %v", file, err)
			continue
		}
		f.Files[newFile.Name] = newFile
	}
	// Generate all child folders
	for _, childFolder := range folders {
		childFolderPath := filepath.Join(f.FullPath, childFolder)
		newFodler, err := TrackExistingFolder(childFolderPath)
		if err != nil {
			log.Printf("Error tracking folder %s: %v", childFolder, err)
			continue
		}
		f.Children[newFodler.Name] = newFodler
	}
}

func (f *Folder) GenerateContents() ([]string, []string, error) {
	files := make([]string, 0)
	folders := make([]string, 0)
	entries, err := os.ReadDir(f.FullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading directory %s: %v", f.FullPath, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		} else {
			folders = append(folders, entry.Name())
		}
	}
	cleanedFiles := cleanFileList(files)
	return cleanedFiles, folders, nil
}

func (f *Folder) ReadLatestFile(fileName string, w io.Writer) error {
	return f.ReadFile(fileName, f.Files[fileName].LatestVersion, w)
}

func (f *Folder) ReadFile(fileName string, version int, w io.Writer) error {
	if !f.HasChildFile(fileName) {
		return fmt.Errorf("file %s does not exist in folder %s", fileName, f.FullPath)
	}
	err := f.Files[fileName].ReadFile(version, w)
	if err != nil {
		return fmt.Errorf("error reading file %s version %d: %v", fileName, version, err)
	}
	return nil
}

func cleanFileList(files []string) []string {
	// We use a set to avoid duplicates
	set := make(map[string]struct{})
	for _, file := range files {
		lastIndex := strings.LastIndex(file, ".")
		version := file[lastIndex+1:]
		if version == "" {
			continue
		}
		if _, err := strconv.Atoi(version); err != nil {
			continue
		}
		cleanedFile := file[:lastIndex]
		if _, exists := set[cleanedFile]; !exists {
			set[cleanedFile] = struct{}{}
		}
	}
	// Convert the set back to a slice
	cleanedFiles := make([]string, 0, len(set))
	for file := range set {
		cleanedFiles = append(cleanedFiles, file)
	}
	return cleanedFiles
}

func folderName(path string) string {
	cleaned := filepath.Clean(path)
	return filepath.Base(cleaned)
}
