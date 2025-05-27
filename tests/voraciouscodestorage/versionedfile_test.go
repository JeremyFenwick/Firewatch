package voraciouscodestorage_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/JeremyFenwick/firewatch/internal/voraciouscodestorage"
	"github.com/stretchr/testify/assert"
)

func TestNewVersionedFile(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	reader := strings.NewReader("swimming swan")
	newFile, err := voraciouscodestorage.NewVersionedFile(dir+"/test.txt", reader, 13)
	assert.NoError(t, err)
	latestFile, err := newFile.GetLatest()
	assert.NoError(t, err)
	var buff bytes.Buffer
	latestFile.ReadFile(&buff)
	assert.Equal(t, "swimming swan", buff.String())
	buff.Reset()
	firstFile, err := newFile.GetVersion(1)
	assert.NoError(t, err)
	firstFile.ReadFile(&buff)
	assert.Equal(t, "swimming swan", buff.String())
	newFile.Delete()
	assert.NoError(t, err)
}

func TestAddNewVersion(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	reader := strings.NewReader("swimming swan")
	newFile, err := voraciouscodestorage.NewVersionedFile(dir+"/test.txt", reader, 13)
	assert.NoError(t, err)
	nextReader := strings.NewReader("dancing duck")
	newFile.AddVersion(nextReader, 13)
	latestFile, err := newFile.GetLatest()
	assert.NoError(t, err)
	var buff bytes.Buffer
	latestFile.ReadFile(&buff)
	assert.Equal(t, "dancing duck", buff.String())
	buff.Reset()
	firstFile, err := newFile.GetVersion(1)
	assert.NoError(t, err)
	firstFile.ReadFile(&buff)
	assert.Equal(t, "swimming swan", buff.String())
	newFile.Delete()
}
