package extension

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFileMatches(t *testing.T) {
	t.Run("ReturnsMatchIfSearchingForModExtension", func(t *testing.T) {

		service := service{}

		// act
		matches, err := service.findFileMatches("../..", "*.mod")

		assert.Nil(t, err)
		assert.Equal(t, 1, len(matches))
		assert.True(t, strings.HasSuffix(matches[0], "go.mod"))
	})
}
