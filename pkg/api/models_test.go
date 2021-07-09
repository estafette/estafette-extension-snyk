package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLanguageString(t *testing.T) {
	t.Run("UnknownReturnsUnknown", func(t *testing.T) {
		assert.Equal(t, "Unknown", PackageManagerUnknown.String())
	})

	t.Run("GolangReturnsGolang", func(t *testing.T) {
		assert.Equal(t, "GoModules", PackageManagerGoModules.String())
	})

	t.Run("NodeReturnsNode", func(t *testing.T) {
		assert.Equal(t, "Npm", PackageManagerNpm.String())
	})

	t.Run("MavenReturnsMaven", func(t *testing.T) {
		assert.Equal(t, "Maven", PackageManagerMaven.String())
	})

	t.Run("DotnetReturnsDotnet", func(t *testing.T) {
		assert.Equal(t, "Nuget", PackageManagerNuget.String())
	})

	t.Run("PythonReturnsPython", func(t *testing.T) {
		assert.Equal(t, "Pip", PackageManagerPip.String())
	})

	t.Run("DockerReturnsDocker", func(t *testing.T) {
		assert.Equal(t, "Docker", PackageManagerDocker.String())
	})
}
