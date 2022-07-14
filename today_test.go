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
var oneMinuteSince time.Duration = 1 * time.Minute
var twoDaysSince time.Duration = 48 * time.Hour
var zeroTime time.Duration = 0 * time.Minute

const thisRepo string = "https://github.com/jdockerty/today"

// TODO: Cleanup the larger tests which use setupRepo func.

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

func TestFullContainsAuthor(t *testing.T) {

	assert := assert.New(t)

	r, err := setupRepo(thisRepo, t)
	assert.Nil(err)

	w, err := r.Worktree()
	assert.Nil(err)

	_, err = w.Commit("TEST", testSignature)
	assert.Nil(err)

	m := make(map[string]*git.Repository, 1)
	m["today"] = r
	msgs, err := getCommitMessages(m, testSignature.Author.Name, true, oneMinuteSince)
	assert.Nil(err)

	assert.Contains(msgs, "today")
	assert.GreaterOrEqual(2, len(msgs)) // Our 2 commits here and any others which are within 48 hours.

}
func TestFullContainsAuthorHasNoCommits(t *testing.T) {

	assert := assert.New(t)

	r, err := setupRepo(thisRepo, t)
	assert.Nil(err)

	w, err := r.Worktree()
	assert.Nil(err)

	_, err = w.Commit("TEST", testSignature)
	assert.Nil(err)

	m := make(map[string]*git.Repository, 1)
	m["today"] = r
	msgs, err := getCommitMessages(m, "INVALID_COMMIT_AUTHOR", true, oneMinuteSince)
	assert.Nil(err)

	assert.Contains(msgs, "today")
	assert.Equal(0, len(msgs["today"]))

}

func TestNoResultsForZeroSinceValue(t *testing.T) {

	assert := assert.New(t)

	r, err := setupRepo(thisRepo, t)
	assert.Nil(err)

	w, err := r.Worktree()
	assert.Nil(err)

	_, err = w.Commit("TEST", testSignature)
	assert.Nil(err)

	m := make(map[string]*git.Repository, 1)
	m["today"] = r
	msgs, err := getCommitMessages(m, "", true, zeroTime)
	assert.Nil(err)

	assert.Equal(0, len(msgs["today"]))

}
func TestResultsForMinimalSinceValue(t *testing.T) {

	assert := assert.New(t)

	r, err := setupRepo(thisRepo, t)
	assert.Nil(err)

	w, err := r.Worktree()
	assert.Nil(err)

	_, err = w.Commit("TEST", testSignature)
	assert.Nil(err)

	m := make(map[string]*git.Repository, 1)
	m["today"] = r
	msgs, err := getCommitMessages(m, "", true, oneMinuteSince)
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
