package main

import (
	"syscall"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TODO: Cleanup the larger tests which use setupRepo func into `suite` tests.

var (
	currentDir, _ = syscall.Getwd()
	testSignature = &git.CommitOptions{
		Author: &object.Signature{
			Name:  "testUser",
			Email: "testEmail",
			When:  time.Now().UTC(),
		},
	}
)

const (
	thisRepo       string        = "https://github.com/jdockerty/today"
	zeroTime       time.Duration = 0 * time.Minute
	oneMinuteSince time.Duration = 1 * time.Minute
	twoDaysSince   time.Duration = 48 * time.Hour
)

// Involves all tests which require reading/writing of a commit, this is requires the setup and teardown
// of a directory which is tracked by git.
type CommitSuite struct {
	suite.Suite
	Repo     *git.Repository // Repo to use within the test suite.
	Worktree *git.Worktree
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

// Setup before every test, whilst this might be a little inefficient for getting the repository before
// every test, we do so to ensure that our commit history is completely fresh from previous test commits.
func (suite *CommitSuite) SetupTest() {
	r, err := setupRepo(thisRepo, suite.T())
	if err != nil {
		suite.FailNow("Unable to setup CommitSuite repo: %s\n", err)
	}

	suite.Repo = r

	w, err := r.Worktree()
	if err != nil {
		suite.FailNow("Unable to setup CommitSuite worktree: %s\n", err)
	}
	suite.Worktree = w
}

func (suite *CommitSuite) TestFullContainsAuthor() {

	assert := assert.New(suite.T())

	w, err := suite.Repo.Worktree()
	assert.Nil(err)

	_, err = w.Commit("TEST", testSignature)
	assert.Nil(err)

	m := make(map[string]*git.Repository, 1)
	m["today"] = suite.Repo
	msgs, err := getCommitMessages(m, testSignature.Author.Name, true, oneMinuteSince)
	assert.Nil(err)

	assert.Contains(msgs, "today")
	assert.GreaterOrEqual(2, len(msgs)) // Our 2 commits here and any others which are within 48 hours.

}

func (suite *CommitSuite) TestFullContainsAuthorHasNoCommits() {

	assert := assert.New(suite.T())

	_, err := suite.Worktree.Commit("TEST", testSignature)
	assert.Nil(err)

	m := make(map[string]*git.Repository, 1)
	m["today"] = suite.Repo
	msgs, err := getCommitMessages(m, "INVALID_COMMIT_AUTHOR", true, oneMinuteSince)
	assert.Nil(err)

	assert.Contains(msgs, "today")
	assert.Equal(0, len(msgs["today"]))
}

func (suite *CommitSuite) TestNoResultsForZeroSinceValue() {

	assert := assert.New(suite.T())

	_, err := suite.Worktree.Commit("TEST", testSignature)
	assert.Nil(err)

	m := make(map[string]*git.Repository, 1)
	m["today"] = suite.Repo
	msgs, err := getCommitMessages(m, "", true, zeroTime)
	assert.Nil(err)

	assert.Equal(0, len(msgs["today"]))

}

func (suite *CommitSuite) TestResultsForMinimalSinceValue() {

	assert := assert.New(suite.T())

	_, err := suite.Worktree.Commit("TEST", testSignature)
	assert.Nil(err)

	m := make(map[string]*git.Repository, 1)
	m["today"] = suite.Repo
	msgs, err := getCommitMessages(m, "", true, oneMinuteSince)
	assert.Nil(err)

	assert.Contains(msgs, "today")
	assert.Equal("TEST", msgs["today"][0]) // We know there is a single message here, as we made it for the test
	assert.Equal(1, len(msgs))

}

func (suite *CommitSuite) TestResultsForLargerSinceValue() {

	assert := assert.New(suite.T())

	_, err := suite.Worktree.Commit("TEST", testSignature)
	assert.Nil(err)

	_, err = suite.Worktree.Commit("TEST2", testSignature)
	assert.Nil(err)

	m := make(map[string]*git.Repository, 1)
	m["today"] = suite.Repo
	msgs, err := getCommitMessages(m, "", true, twoDaysSince)
	assert.Nil(err)

	assert.Contains(msgs, "today")
	assert.GreaterOrEqual(2, len(msgs)) // Our 2 commits here and any others which are within 48 hours.

}

func (suite *CommitSuite) TestShortCommitMessage() {

	assert := assert.New(suite.T())

	_, err := suite.Worktree.Commit("TEST\nNOT SEEN", testSignature)
	assert.Nil(err)

	m := make(map[string]*git.Repository, 1)
	m["today"] = suite.Repo
	msgs, err := getCommitMessages(m, "", true, oneMinuteSince)
	assert.Nil(err)

	assert.Equal(4, len(msgs["today"][0])) // Length of 'TEST' = 4
}

func (suite *CommitSuite) TestLongCommitMessage() {

	assert := assert.New(suite.T())

	_, err := suite.Worktree.Commit("TEST\nSEEN", testSignature)
	assert.Nil(err)

	m := make(map[string]*git.Repository, 1)
	m["today"] = suite.Repo
	msgs, err := getCommitMessages(m, "", false, oneMinuteSince)
	assert.Nil(err)

	assert.Equal("TEST\nSEEN", msgs["today"][0])
}

func TestCommitSuite(t *testing.T) {
	suite.Run(t, new(CommitSuite))
}

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
