package voraciouscodestorage_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/JeremyFenwick/firewatch/internal/voraciouscodestorage"
	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	reader := strings.NewReader("swimming swan")
	newFile, err := voraciouscodestorage.NewFile(reader, dir+"/test.txt", ".txt", 0)
	assert.NoError(t, err)
	var buff bytes.Buffer
	newFile.ReadFile(&buff)
	assert.Equal(t, "swimming swan", buff.String())
	newFile.Delete()
	assert.NoError(t, err)
}
