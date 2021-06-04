package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLanguageString(t *testing.T) {
	t.Run("UnknownReturnsUnknown", func(t *testing.T) {
		assert.Equal(t, "Unknown", LanguageUnknown.String())
	})

	t.Run("GolangReturnsGolang", func(t *testing.T) {
		assert.Equal(t, "Golang", LanguageGolang.String())
	})

	t.Run("NodeReturnsNode", func(t *testing.T) {
		assert.Equal(t, "Node", LanguageNode.String())
	})

	t.Run("MavenReturnsMaven", func(t *testing.T) {
		assert.Equal(t, "Maven", LanguageMaven.String())
	})

	t.Run("DotnetReturnsDotnet", func(t *testing.T) {
		assert.Equal(t, "Dotnet", LanguageDotnet.String())
	})

	t.Run("PythonReturnsPython", func(t *testing.T) {
		assert.Equal(t, "Python", LanguagePython.String())
	})
}
