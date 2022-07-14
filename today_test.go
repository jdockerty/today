package main

import (
	"syscall"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

var currentDir, _ = syscall.Getwd()
var testSignature = &git.CommitOptions{
	Author: &object.Signature{
		Name:  "testUser",
		Email: "testEmail",
		When:  time.Now().UTC(),
	},
}

const thisRepo string = "https://github.com/jdockerty/today"

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

func TestResultsForLargerSinceValue(t *testing.T) {

	assert := assert.New(t)

	twoDaysSince := 48 * time.Hour

	r, err := setupRepo(thisRepo, t)
	assert.Nil(err)

	w, err := r.Worktree()
	assert.Nil(err)

	_, err = w.Commit("TEST", testSignature)
	assert.Nil(err)
	_, err = w.Commit("TEST2", testSignature)
	assert.Nil(err)

	m := make(map[string]*git.Repository, 1)
	m["today"] = r
	msgs, err := getCommitMessages(m, true, twoDaysSince)
	assert.Nil(err)

	assert.Contains(msgs, "today")
	assert.GreaterOrEqual(2, len(msgs)) // Our 2 commits here and any others which are within 48 hours.

}

func TestNoResultsForMinimalSinceValue(t *testing.T) {

	assert := assert.New(t)

	oneMinuteSince := 1 * time.Minute

	r, err := setupRepo(thisRepo, t)
	assert.Nil(err)

	w, err := r.Worktree()
	assert.Nil(err)

	_, err = w.Commit("TEST", testSignature)
	assert.Nil(err)

	m := make(map[string]*git.Repository, 1)
	m["today"] = r
	msgs, err := getCommitMessages(m, true, oneMinuteSince)
	assert.Nil(err)

	assert.Contains(msgs, "today")
	assert.Equal("TEST", msgs["today"][0]) // We know there is a single message here, as we made it for the test
	assert.Equal(1, len(msgs))

}

func setupRepo(url string, t *testing.T) (*git.Repository, error) {

	r, err := git.PlainClone(t.TempDir(), false, &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return nil, err
	}

	return r, nil
}
