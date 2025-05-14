package voraciouscodestorage_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/JeremyFenwick/firewatch/internal/voraciouscodestorage"
	"github.com/stretchr/testify/assert"
)

func TestAddNewFile(t *testing.T) {
	path, _ := filepath.Abs("./data/subfolder/new_text.txt")
	fileBuffer := bytes.NewBufferString("singing ranger")
	file, err := voraciouscodestorage.AddNewFile(fileBuffer, path)
	assert.NoError(t, err, "should not return an error when adding a new file")
	assert.NotNil(t, file, "should return a file object")
	var buffer bytes.Buffer
	file.ReadLatestFile(&buffer)
	assert.Equal(t, "singing ranger", buffer.String(), "should read the latest file correctly")
	buffer.Reset()
	file.ReadFile(1, &buffer)
	assert.Equal(t, "singing ranger", buffer.String(), "should read the latest file correctly")
	buffer.Reset()
	file.AddNewVersion(bytes.NewBufferString("prancing horse"))
	file.ReadLatestFile(&buffer)
	assert.Equal(t, "prancing horse", buffer.String(), "should read the latest file correctly")
	buffer.Reset()
	file.ReadFile(1, &buffer)
	assert.Equal(t, "singing ranger", buffer.String(), "should read the older file version correctly")
	err = file.Remove()
	assert.NoError(t, err, "should not return an error when removing the file")
}

func TestTrackExistingFile(t *testing.T) {
	path, _ := filepath.Abs("./data/subfolder/text.txt")
	file, err := voraciouscodestorage.TrackExistingFile(path)
	assert.NoError(t, err, "should not return an error when tracking an existing file")
	assert.NotNil(t, file, "should return a file object")
	var buffer bytes.Buffer
	file.ReadLatestFile(&buffer)
	assert.Equal(t, "singing ranger", buffer.String(), "should read the latest file correctly")
	buffer.Reset()
	file.ReadFile(1, &buffer)
	assert.Equal(t, "prancing horse", buffer.String(), "should read the older file version correctly")
}
