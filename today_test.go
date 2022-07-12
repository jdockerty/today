package main

import (
	"syscall"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
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

func TestGetRepositories(t *testing.T) {
	gitDir := []string{currentDir}
	_, err := getRepositories(gitDir)
	assert.Nil(t, err)
}

func TestDoesContainAuthor(t *testing.T) {
	testCommit := &object.Commit{
		Author: object.Signature{
			Name:  "testUser",
			Email: "test@example.com",
			When:  time.Now().UTC(),
		},
	}

	got := containsAuthor(testCommit, "testUser")
	assert.True(t, got)
}

func TestDoesNotContainAuthor(t *testing.T) {
	testCommit := &object.Commit{
		Author: object.Signature{
			Name:  "testUser",
			Email: "test@example.com",
			When:  time.Now().UTC(),
		},
	}

	got := containsAuthor(testCommit, "John")
	assert.False(t, got)
}
