package voraciouscodestorage

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileSystem struct {
	Mutex sync.RWMutex
	Root  *Folder
}

func NewFileSystem(rootPath string) (*FileSystem, error) {
	// Check if the root path exists
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("root path %s does not exist", rootPath)
	}
	absolutePath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path of %s: %v", rootPath, err)
	}
	rootFolder := &Folder{
		Name:     rootPath,
		FullPath: absolutePath,
		Files:    make(map[string]*File),
		Children: make(map[string]*Folder),
	}
	// Begin tracking the root folder
	files, folders, _ := rootFolder.GenerateContents()
	rootFolder.TrackFolderContents(files, folders)
	return &FileSystem{
		Root: rootFolder,
	}, nil
}

func (fs *FileSystem) GetFolder(dir string) (*Folder, error) {
	fs.Mutex.RLock()
	defer fs.Mutex.RUnlock()

	folderNames := folderNames(dir)
	// Navigate to the target folder
	currentFolder := fs.Root
	for _, folderName := range folderNames {
		// Check if the current folder has the child folder
		exists := currentFolder.HasChildFolder(folderName)
		if exists {
			currentFolder = currentFolder.Children[folderName]
		} else {
			return nil, fmt.Errorf("folder %s does not exist", folderName)
		}
	}
	// Get all files in the queue
	return currentFolder, nil
}

func (fs *FileSystem) AddFile(r io.Reader, fullPath string) (*File, error) {
	fs.Mutex.Lock()
	defer fs.Mutex.Unlock()

	dir, fileName := splitDirAndFile(fullPath)
	// Check if the file name is empty
	if fileName == "" {
		return nil, fmt.Errorf("file name is empty")
	}
	// Create the folders if required
	targetFolder, err := fs.createFolders(dir)
	if err != nil {
		return nil, fmt.Errorf("error creating folders for file %s: %v", fullPath, err)
	}
	// Add the file
	// Check if the file already exists
	exists := targetFolder.HasChildFile(fileName)
	if exists {
		existingFile := targetFolder.Files[fileName]
		existingFile.AddNewVersion(r)
		return existingFile, nil
	} else {
		filePath := filepath.Join(targetFolder.FullPath, fileName)
		file, err := AddNewFile(r, filePath)
		if err != nil {
			return nil, fmt.Errorf("error adding file %s: %v", fullPath, err)
		}
		// Add the file to the target folder
		targetFolder.Files[fileName] = file
		return file, nil
	}
}

func (fs *FileSystem) ReadLatestFile(fullpath string, w io.Writer) error {
	fs.Mutex.RLock()
	defer fs.Mutex.RUnlock()

	dir, fileName := splitDirAndFile(fullpath)
	// Check if the file name is empty
	if fileName == "" {
		return fmt.Errorf("file name is empty")
	}
	folderNames := folderNames(dir)
	currentFolder := fs.Root
	for _, folderName := range folderNames {
		// Check if the current folder has the child folder
		exists := currentFolder.HasChildFolder(folderName)
		if exists {
			currentFolder = currentFolder.Children[folderName]
		} else {
			return fmt.Errorf("folder %s does not exist", folderName)
		}
	}
	err := currentFolder.ReadLatestFile(fileName, w)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", fullpath, err)
	}
	return nil
}

func (fs *FileSystem) ReadFile(fullPath string, version int, w io.Writer) error {
	fs.Mutex.RLock()
	defer fs.Mutex.RUnlock()

	dir, fileName := splitDirAndFile(fullPath)
	// Check if the file name is empty
	if fileName == "" {
		return fmt.Errorf("file name is empty")
	}
	folderNames := folderNames(dir)
	currentFolder := fs.Root
	for _, folderName := range folderNames {
		// Check if the current folder has the child folder
		exists := currentFolder.HasChildFolder(folderName)
		if exists {
			currentFolder = currentFolder.Children[folderName]
		} else {
			return fmt.Errorf("folder %s does not exist", folderName)
		}
	}
	err := currentFolder.ReadFile(fileName, version, w)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", fullPath, err)
	}
	return nil
}

func (fs *FileSystem) Clear() {
	fs.Mutex.Lock()
	defer fs.Mutex.Unlock()

	entries, err := os.ReadDir(fs.Root.FullPath)
	if err != nil {
		log.Printf("Error reading directory %s: %v", fs.Root.FullPath, err)
		return
	}

	for _, entry := range entries {
		entryPath := filepath.Join(fs.Root.FullPath, entry.Name())
		err := os.RemoveAll(entryPath)
		if err != nil {
			log.Printf("Error removing entry %s: %v", entryPath, err)
			return
		}
	}
	fs.Root.Files = make(map[string]*File)
	fs.Root.Children = make(map[string]*Folder)
}

func (fs *FileSystem) createFolders(fullPath string) (*Folder, error) {
	// Check if the folder already exists
	folderNames := folderNames(fullPath)
	currentFolder := fs.Root
	for _, folderName := range folderNames {
		// Check if the current folder has the child folder
		exists := currentFolder.HasChildFolder(folderName)
		if exists {
			currentFolder = currentFolder.Children[folderName]
		} else {
			// Create the new folder
			newFolder, err := CreateNewFolder(currentFolder.FullPath + string(os.PathSeparator) + folderName)
			if err != nil {
				return nil, fmt.Errorf("error creating folder %s: %v", folderName, err)
			}
			currentFolder.Children[folderName] = newFolder
			currentFolder = newFolder
		}
	}
	return currentFolder, nil
}

func folderNames(directory string) []string {
	// Split the full path into its components
	splitParts := strings.Split(directory, string(os.PathSeparator))
	result := make([]string, 0)
	for _, part := range splitParts {
		// Ignore empty parts (e.g., leading or trailing slashes)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func splitDirAndFile(path string) (dir string, file string) {
	cleaned := filepath.Clean(path)
	base := filepath.Base(cleaned)

	// If base has a file extension, assume it's a file
	if ext := filepath.Ext(base); ext != "" {
		return filepath.Dir(cleaned), base
	}

	// It's a directory path
	return cleaned, ""
}
