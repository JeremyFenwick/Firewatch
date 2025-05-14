package voraciouscodestorage_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/JeremyFenwick/firewatch/internal/voraciouscodestorage"
	"github.com/stretchr/testify/assert"
)

func TestTrackExistingFolder(t *testing.T) {
	path, _ := filepath.Abs("./data/subfolder")
	folder, err := voraciouscodestorage.TrackExistingFolder(path)
	assert.NoError(t, err, "should not return an error when tracking an existing folder")
	assert.True(t, folder.HasChildFile("new_text.txt"))
	assert.True(t, folder.HasChildFile("text.txt"))
	assert.True(t, folder.HasChildFolder("innerfolder"))
}

func TestCreateNewFolder(t *testing.T) {
	path, _ := filepath.Abs("./data/subfolder/testfolder")
	folder, err := voraciouscodestorage.CreateNewFolder(path)
	assert.NoError(t, err, "should not return an error when creating a new folder")
	assert.Equal(t, "testfolder", folder.Name)
	assert.Equal(t, path, folder.FullPath)
	assert.Empty(t, folder.Files)
	assert.Empty(t, folder.Children)

	// Clean up
	err = folder.Remove()
	assert.NoError(t, err, "should not return an error when removing the folder")
}

func TestReadFile(t *testing.T) {
	path, _ := filepath.Abs("./data/subfolder")
	folder, err := voraciouscodestorage.TrackExistingFolder(path)
	assert.NoError(t, err, "should not return an error when tracking an existing folder")
	assert.True(t, folder.HasChildFile("text.txt"))
	var buffer bytes.Buffer
	folder.ReadLatestFile("text.txt", &buffer)
	assert.Equal(t, "singing ranger", buffer.String(), "should read the latest file correctly")
	buffer.Reset()
	folder.ReadFile("text.txt", 1, &buffer)
	assert.Equal(t, "prancing horse", buffer.String(), "should read the older file version correctly")
}
