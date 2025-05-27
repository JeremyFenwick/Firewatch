package voraciouscodestorage_test

import (
	"os"
	"strings"
	"testing"

	"github.com/JeremyFenwick/firewatch/internal/voraciouscodestorage"
	"github.com/stretchr/testify/assert"
)

func TestFileManagerPopulate(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	folderPath := dir + "/data"
	fm, err := voraciouscodestorage.NewFileManager(folderPath)
	assert.NoError(t, err)
	folder, err := fm.GetFolder("/subfolder/subsubfolder")
	assert.NoError(t, err)
	assert.NotNil(t, folder)
	rootFolder, err := fm.GetFolder("/")
	assert.NoError(t, err)
	rootFolderFiles := rootFolder.GetFiles()
	assert.Len(t, rootFolderFiles, 1)
	assert.Equal(t, "test.txt", rootFolderFiles[0].FileName)
}

func TestFileManagerAddFile(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	folderPath := dir + "/data"
	fm, err := voraciouscodestorage.NewFileManager(folderPath)
	assert.NoError(t, err)

	// Create a new file
	fileContent := "This is a test file."
	r := strings.NewReader(fileContent)
	versionedFile, err := fm.AddFile("newfile.txt", r, len(fileContent))
	assert.NoError(t, err)
	assert.NotNil(t, versionedFile)

	// Check if the file was added correctly
	folder, err := fm.GetFolder("/")
	assert.NoError(t, err)
	files := folder.GetFiles()
	assert.Len(t, files, 1)
	assert.Equal(t, "newfile.txt", files[0].FileName)

	// Clean up the created file
	files[0].Delete()
}

func TestFileManagerAddFileSubfolder(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	folderPath := dir + "/data"
	fm, err := voraciouscodestorage.NewFileManager(folderPath)
	assert.NoError(t, err)

	// Create a new file in a subfolder
	fileContent := "This is a test file in a subfolder."
	r := strings.NewReader(fileContent)
	versionedFile, err := fm.AddFile("newsubfolder/newfile.txt", r, len(fileContent))
	assert.NoError(t, err)
	assert.NotNil(t, versionedFile)

	// Check if the file was added correctly
	folder, err := fm.GetFolder("/newsubfolder")
	assert.NoError(t, err)
	files := folder.GetFiles()
	assert.Len(t, files, 1)
	assert.Equal(t, "newfile.txt", files[0].FileName)

	// Clean up the created file
	files[0].Delete()
}
