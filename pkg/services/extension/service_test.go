package extension

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFileMatches(t *testing.T) {
	t.Run("ReturnsMatchIfSearchingForModExtension", func(t *testing.T) {

		service := service{}

		// act
		matches, err := service.findFileMatches("../../..", []string{"*.mod"}, []string{})

		assert.Nil(t, err)
		assert.Equal(t, 1, len(matches))
		assert.Equal(t, "go.mod", filepath.Base(matches[0]))
	})
}
