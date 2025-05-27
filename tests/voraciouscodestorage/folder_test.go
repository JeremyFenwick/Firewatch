package voraciouscodestorage_test

import (
	"os"
	"strings"
	"testing"

	"github.com/JeremyFenwick/firewatch/internal/voraciouscodestorage"
	"github.com/stretchr/testify/assert"
)

func TestCreateFolder(t *testing.T) {
	// This test will create a folder and check if it exists.
	// It will also clean up by deleting the folder after the test.
	dir, err := os.Getwd()
	assert.NoError(t, err)

	folderPath := dir + "/test_folder"
	newFolder, err := voraciouscodestorage.NewFolder(folderPath)
	assert.NoError(t, err)
	// Check if the folder was created
	_, err = os.Stat(folderPath)
	assert.NoError(t, err)
	// Clean up by deleting the folder
	err = newFolder.Delete()
	assert.NoError(t, err)
}

func TestCreateSubfolder(t *testing.T) {
	// This test will create a subfolder within an existing folder.
	dir, err := os.Getwd()
	assert.NoError(t, err)

	folderPath := dir + "/test_folder"
	newFolder, err := voraciouscodestorage.NewFolder(folderPath)
	assert.NoError(t, err)

	subFolderName := "sub_folder"
	_, err = newFolder.AddNewSubFolder(subFolderName)
	assert.NoError(t, err)

	// Check if the subfolder was created
	_, err = os.Stat(folderPath + "/" + subFolderName)
	assert.NoError(t, err)

	// Clean up by deleting the subfolder and the main folder
	err = newFolder.Delete()
	assert.NoError(t, err)
}

func TestCreateFileInFolder(t *testing.T) {
	// This test will create a file in an existing folder.
	dir, err := os.Getwd()
	assert.NoError(t, err)

	folderPath := dir + "/test_folder"
	newFolder, err := voraciouscodestorage.NewFolder(folderPath)
	assert.NoError(t, err)

	fileName := "test_file.txt"
	content := "Hello, World!"
	reader := strings.NewReader(content)

	_, err = newFolder.AddNewFile(fileName, reader, len(content))
	assert.NoError(t, err)

	// Check if the first version of the file was created
	_, err = os.Stat(folderPath + "/" + fileName + ".1")
	assert.NoError(t, err)

	// Clean up by deleting the folder
	err = newFolder.Delete()
	assert.NoError(t, err)
}
