package voraciouscodestorage_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/JeremyFenwick/firewatch/internal/voraciouscodestorage"
	"github.com/stretchr/testify/assert"
)

func TestGenerateFileSystem(t *testing.T) {
	fs, err := voraciouscodestorage.NewFileSystem("./data")
	assert.NoError(t, err, "should not return an error when generating the file system")
	assert.NotNil(t, fs, "FileSystem should not be nil")
}

func TestFilesInDirectoryFromRoot(t *testing.T) {
	fs, err := voraciouscodestorage.NewFileSystem("./data")
	assert.NoError(t, err, "should not return an error when generating the file system")
	assert.NotNil(t, fs, "FileSystem should not be nil")
	folder, err := fs.GetFolder("/")
	assert.NoError(t, err, "should not return an error when getting files in directory")
	files := folder.GetChildAllFiles()
	assert.NoError(t, err, "should not return an error when getting files in directory")
	assert.Len(t, files, 0, "should find 2 files in the directory")
}

func TestFilesInDirectoryFromSubFolder(t *testing.T) {
	fs, err := voraciouscodestorage.NewFileSystem("./data")
	assert.NoError(t, err, "should not return an error when generating the file system")
	assert.NotNil(t, fs, "FileSystem should not be nil")
	folder, err := fs.GetFolder("/subfolder/innerfolder")
	assert.NoError(t, err, "should not return an error when getting files in directory")
	files := folder.GetChildAllFiles()
	assert.NoError(t, err, "should not return an error when getting files in directory")
	assert.Len(t, files, 1, "should find 1 file in the directory")
	assert.Equal(t, "inner_text.txt", files[0].Name, "first file should be inner_text.txt")
}

func TestFSCreateFile(t *testing.T) {
	fs, err := voraciouscodestorage.NewFileSystem("./data")
	content := strings.NewReader("test content\n")
	assert.NoError(t, err, "should not return an error when generating the file system")
	assert.NotNil(t, fs, "FileSystem should not be nil")
	file, err := fs.AddFile(content, "/subfolder/createfiletest/test.txt")
	assert.NoError(t, err, "should not return an error when creating a file")
	assert.Equal(t, "test.txt", file.Name, "folder name should match the provided name")
	content2 := strings.NewReader("test content 2\n")
	file2, err := fs.AddFile(content2, "/subfolder/createfiletest/test.txt")
	assert.NoError(t, err, "should not return an error when creating a file")
	assert.Equal(t, 2, file2.LatestVersion, "file version should be 2")
	file.Remove()
}

func TestFSReadFile(t *testing.T) {
	fs, err := voraciouscodestorage.NewFileSystem("./data")
	assert.NoError(t, err, "should not return an error when generating the file system")
	var buffer bytes.Buffer
	err = fs.ReadLatestFile("subfolder/text.txt", &buffer)
	assert.NoError(t, err, "should not return an error when reading the latest file")
	assert.Equal(t, "singing ranger", buffer.String(), "file content should match the expected content")
}

func TestFSReadFileVersion(t *testing.T) {
	fs, err := voraciouscodestorage.NewFileSystem("./data")
	assert.NoError(t, err, "should not return an error when generating the file system")
	var buffer bytes.Buffer
	err = fs.ReadFile("subfolder/text.txt", 1, &buffer)
	assert.NoError(t, err, "should not return an error when reading the latest file")
	assert.Equal(t, "prancing horse", buffer.String(), "file content should match the expected content")
}
