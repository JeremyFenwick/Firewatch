package mobinthemiddle_test

import (
	"testing"

	"github.com/JeremyFenwick/firewatch/internal/mobinthemiddle"
	"github.com/stretchr/testify/assert"
)

func TestInjectionAttack(t *testing.T) {
	t.Run("Basic replacement", func(t *testing.T) {
		result := mobinthemiddle.MotmAttack("7adNeSwJkMakpEcln9HEtthSRtxdmEHOT8T")
		assert.Equal(t, result, "7YWHMfk9JZe0LM0g1ZauHuiSxhI")
	})

	t.Run("Sentence replacement", func(t *testing.T) {
		result := mobinthemiddle.MotmAttack(" Hi alice, please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX")
		assert.Equal(t, result, " Hi alice, please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI")
	})

	t.Run("Whitespace test", func(t *testing.T) {
		result := mobinthemiddle.MotmAttack("  7F1u3wSD5RbOHQmupo9nx4TnhQ Hi alice, please send payment to   ")
		assert.Equal(t, result, "  7YWHMfk9JZe0LM0g1ZauHuiSxhI Hi alice, please send payment to   ")
	})

	t.Run("Do nothing", func(t *testing.T) {
		result := mobinthemiddle.MotmAttack("Hi alice, please send payment to")
		assert.Equal(t, result, "Hi alice, please send payment to")
	})

	t.Run("Empty string", func(t *testing.T) {
		result := mobinthemiddle.MotmAttack("  ")
		assert.Equal(t, result, "  ")
	})
}
