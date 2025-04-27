package insecuresocketlayer_test

import (
	"testing"

	"github.com/JeremyFenwick/firewatch/internal/insecuresocketslayer"
	"github.com/stretchr/testify/assert"
)

func TestToyCount(t *testing.T) {
	toyDesc := "5x Toy1"
	expected := 5
	actual, err := insecuresocketslayer.GetToyCount(toyDesc)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestMostCommonToy(t *testing.T) {
	toys := "5x Toy1, 3x Toy2, 2x Toy3\n"
	expectedToy := "5x Toy1"
	toy, err := insecuresocketslayer.MostCommonToy(toys)
	assert.NoError(t, err)
	assert.Equal(t, expectedToy, toy)
}
