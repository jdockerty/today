package main

import (
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

var currentDir, _ = syscall.Getwd()

func TestValidatePathsProducesErrorWithInvalidDir(t *testing.T) {
	invalidPath := []string{"/does/not/exist"}
	err := validatePaths(invalidPath)
	assert.Error(t, err)
}

func TestValidatePathsProducesErrorWithNoGitDir(t *testing.T) {
	gitDirNotExist := []string{"/tmp"}
	err := validatePaths(gitDirNotExist)
	assert.Error(t, err)
}

func TestValidatePathsWithGitDir(t *testing.T) {
	gitDirExists := []string{currentDir} // Current directory is tracked by git
	err := validatePaths(gitDirExists)
	assert.Nil(t, err)
}
