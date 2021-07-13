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

	t.Run("ReturnsNoMatchIfSearchingForFileInSkippedDirectory", func(t *testing.T) {

		service := service{}

		// act
		matches, err := service.findFileMatches("../../..", []string{"package.json"}, []string{"dist", ".git", "node_modules"})

		assert.Nil(t, err)
		assert.Equal(t, 0, len(matches))
	})

	t.Run("ReturnsMatchIfSearchingForFileOutsideSkippedDirectory", func(t *testing.T) {

		service := service{}

		// act
		matches, err := service.findFileMatches("../../..", []string{"package.json"}, []string{".git", "node_modules"})

		assert.Nil(t, err)
		assert.Equal(t, 1, len(matches))
		assert.Equal(t, "package.json", filepath.Base(matches[0]))
	})
}
