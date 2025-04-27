package insecuresocketlayer_test

import (
	"testing"

	"github.com/JeremyFenwick/firewatch/internal/insecuresocketslayer"
	"github.com/stretchr/testify/assert"
)

func TestMostCommonToy(t *testing.T) {
	toys := "5x Toy1, 3x Toy2, 2x Toy3\n"
	expectedToy := "5x Toy1"
	toy, err := insecuresocketslayer.MostCommonToy([]byte(toys))
	assert.NoError(t, err)
	assert.Equal(t, expectedToy, string(toy))
}
