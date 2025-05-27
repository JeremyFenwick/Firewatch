package voraciouscodestorage

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type FileManager struct {
	Root *Folder
}

type RawFile struct {
	AbsolutePath string
	Name         string
	Version      int
}

func NewFileManager(rootPath string) (*FileManager, error) {
	// Return an error if the root path does not exist
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return nil, err
	}
	// Create a new FileManager with the root folder
	rootFolder, err := NewFolder(rootPath)
	if err != nil {
		return nil, err
	}
	fm := &FileManager{
		Root: rootFolder,
	}
	fm.populate()
	return fm, nil
}

func (fm *FileManager) GetFolder(path string) (*Folder, error) {
	currentFolder := fm.Root
	// Split the path into parts
	parts := strings.Split(filepath.Clean(path), string(os.PathSeparator))
	for _, part := range parts {
		if part == "" {
			continue // Skip empty parts (e.g., leading or trailing slashes)
		}
		// Check if the subfolder exists
		subFolder, exists := currentFolder.SubFolders[part]
		if !exists {
			return nil, os.ErrNotExist // Return an error if the folder does not exist
		}
		currentFolder = subFolder // Move to the next subfolder
	}
	return currentFolder, nil // Return the found folder
}

func (fm *FileManager) AddFile(fullPath string, r io.Reader, bytes int) (*VersionedFile, error) {
	pathSplit := strings.Split(fullPath, string(os.PathSeparator))
	fileName := pathSplit[len(pathSplit)-1]
	folders := pathSplit[:len(pathSplit)-1]
	currentFolder := fm.Root
	for _, folderName := range folders {
		subFolder, exists := currentFolder.SubFolders[folderName]
		if !exists {
			// Create the folder if it does not exist
			subFolder, err := NewFolder(filepath.Join(currentFolder.AbsolutePath, folderName))
			if err != nil {
				return nil, err
			}
			currentFolder.SubFolders[folderName] = subFolder
			currentFolder = subFolder
			continue
		}
		currentFolder = subFolder
	}
	// Check if the file already exists in the current folder
	if existingFile, exists := currentFolder.Files[fileName]; exists {
		// If it exists, add a new version to the existing file
		_, err := existingFile.AddVersion(r, bytes)
		if err != nil {
			return nil, err
		}
		return currentFolder.Files[fileName], nil
	} else {
		// Create the file in the current folder
		currentFolder.AddNewFile(fileName, r, bytes)
		return currentFolder.Files[fileName], nil
	}
}

func (fm *FileManager) Clear() {
	for _, subFolder := range fm.Root.SubFolders {
		subFolder.Delete()
	}
	for _, file := range fm.Root.Files {
		file.Delete() // Delete all files in the root folder
	}
	fm.Root.SubFolders = make(map[string]*Folder)   // Clear subfolders
	fm.Root.Files = make(map[string]*VersionedFile) // Clear files in the root folder
}

func (fm *FileManager) populate() error {
	if fm.Root == nil {
		return nil
	}
	err := populateFolder(fm.Root)
	if err != nil {
		return err
	}
	return nil
}

func populateFolder(folder *Folder) error {
	// Get the files and subfolders in the folder
	entries, err := os.ReadDir(folder.AbsolutePath)
	if err != nil {
		return err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			// If it's a directory, create a new Folder object and populate it
			subFolder, err := NewFolder(folder.AbsolutePath + "/" + entry.Name())
			if err != nil {
				return err
			}
			folder.SubFolders[entry.Name()] = subFolder
			err = populateFolder(subFolder)
			if err != nil {
				return err
			}
			continue
		}
		// If it's a file we need to reconstruct the versioned file
		files = append(files, folder.AbsolutePath+"/"+entry.Name())
	}
	// Group, validate and convert the files into versioned files
	groupedFiles := groupVersionedFiles(files)
	validGroups := parseValidGroups(groupedFiles)
	versionedFiles := groupsToVersionedFiles(validGroups)
	for _, vf := range versionedFiles {
		folder.Files[vf.FileName] = vf
	}
	return nil
}

func groupVersionedFiles(entries []string) map[string][]RawFile {
	grouped := make(map[string][]RawFile)

	for _, absPath := range entries {
		filename := filepath.Base(absPath)
		lastDot := strings.LastIndex(filename, ".")
		if lastDot == -1 {
			continue // Skip files without a version suffix
		}
		name := filename[:lastDot]
		versionStr := filename[lastDot+1:]
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			continue // Skip files with invalid version numbers
		}
		rawFile := RawFile{
			AbsolutePath: absPath,
			Name:         name,
			Version:      version,
		}
		grouped[name] = append(grouped[name], rawFile)
	}
	return grouped
}

func parseValidGroups(grouped map[string][]RawFile) map[string][]RawFile {
	validGroups := make(map[string][]RawFile)

	for name, entries := range grouped {
		// Sort the entries by version number
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Version < entries[j].Version
		})
		// Now check for gaps. If we don't see one, add the valid group
		hasGap := false
		for i := 0; i < len(entries); i++ {
			if entries[i].Version != i+1 {
				hasGap = true
				break
			}
		}
		if !hasGap {
			validGroups[name] = entries
		}
	}

	return validGroups
}

func groupsToVersionedFiles(groups map[string][]RawFile) []*VersionedFile {
	versionedFiles := make([]*VersionedFile, 0, len(groups))

	for name, entries := range groups {
		versionedFile := &VersionedFile{
			LatestVersion: entries[len(entries)-1].Version,
			FileName:      name,
			AbsolutePath:  entries[0].AbsolutePath, // Use the first entry's path as the base
			Files:         make(map[int]*File),
		}
		versionedFiles = append(versionedFiles, versionedFile)
		// Create a file for each version
		for i, entry := range entries {
			info, err := os.Stat(entry.AbsolutePath)
			if err != nil {
				continue // Skip if we can't stat the file
			}
			size := info.Size()
			file := &File{
				AbsolutePath: entry.AbsolutePath,
				Name:         entry.Name,
				Ext:          filepath.Ext(entry.AbsolutePath),
				Bytes:        int(size),
			}
			versionedFile.Files[i] = file
		}
	}

	return versionedFiles
}
