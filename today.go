package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Since is a flag used to control the amount of time to look back in a repository for commits.
// The provided time units must be parseable via time.ParseDuration and it defaults to 12 hours.
var since time.Duration

// Short is a flag for condensing larger messages, this will only display the first line of a commit message.
// This is ideal for repositories where commits may contain longer explanations or reasoning behind the change, but you are familiar with it already and only need a high-level overview.
var short bool

// Author is a 'contains' match on the author of a commit. For example, searching for 'John' will display all commits by the author name '*John*'.
var author string

// validatePaths is used to ensure that only directories that are tracked by git are passed into the application,
// as these directories are used to track the work which was been done, via commit messages.
func validatePaths(paths []string) error {
	for _, p := range paths {

		_, err := os.Stat(p)
		if err != nil {
			return fmt.Errorf("expected directory, but got %s\n", p)
		}

		// Use git to read commit logs for general purpose guide on work done for the day.
		gitDir := fmt.Sprintf("%s/.git", p)
		_, err = os.Stat(gitDir)
		if err != nil {
			return fmt.Errorf("%s is not tracked by git", p)
		}

	}
	return nil
}

// openGitDir is used to open a validated directory which is tracked by git, this returns information
// about the repository that is being tracked.
func openGitDir(dir string) (*git.Repository, error) {

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, err

	}

	return repo, nil
}

// getRepositories will return the git repository definition given a list of directory paths.
func getRepositories(dirs []string) ([]*git.Repository, error) {

	var repos []*git.Repository
	for _, dir := range dirs {
		repo, err := openGitDir(dir)
		if err != nil {
			fmt.Printf("Unable to open local directory '%s': %s\n", dir, err)
			return nil, err
		}

		repos = append(repos, repo)
	}

	return repos, nil
}

// containsAuthor will return whether the commit contains the provided author.
func containsAuthor(c *object.Commit, author string) bool {
	return strings.Contains(c.Author.Name, author)
}

// getBaseDirectoryName is a simple wrapper for getting the base of the provided directory
// with added benefit of using the correct current directory when provided.
func getBaseDirectoryName(p string) (string, error) {
	if p == "./" || p == "." {
		currentDir, err := syscall.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Base(currentDir), nil
	}

	return filepath.Base(p), nil
}

// getCommitMessages is used to map together the repository to a list of valid messages, dependent on the flags that were passed.
func getCommitMessages(dirToRepo map[string]*git.Repository, author string, short bool, since time.Duration) (map[string][]string, error) {

	msgs := make(map[string][]string)

	for dir, repo := range dirToRepo {

		sanitisedDir, err := getBaseDirectoryName(dir)
		if err != nil {
			return nil, err
		}
		// Initialise map before populating messages.
		// This largely comes in handy when a directory is passed where there are no messages in the given 'since' range
		// so it can be displayed as no messages, as opposed to no output whatsoever.
		msgs[sanitisedDir] = []string{}

		ref, err := repo.Head()
		if err != nil {
			return nil, err
		}

		cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			return nil, err
		}

		now := time.Now().UTC()
		currentCommit, err := cIter.Next()
		if err != nil {
			return nil, err
		}

		commitTime := currentCommit.Author.When.UTC()

		// The UTC time of now - the provided 'since' value.
		// We use time.Add with a negative number to subtract here, rather than time.Sub, so that we produce a time.Time value to compare, not a time.Duration.
		timeSince := now.Add(-since)

		// Only iterate whilst we meet the criteria of the current commit being before our `since` value.
		// Once we have reached the commit where this is not the case, we can stop as commits are in chronological order.
		// Note: We are not accounting for any `--date` manipulation, this will simply use the timestamp it currently has,
		// meaning that it can stop prematurely if it no longer matches the loop clause.
		for commitTime.After(timeSince) {

			// Get the next commit ready here so avoid needing to duplicate logic branches
			// when needing to skip commits.
			// TODO: Can we tidy this up in an elegant way?
			nextCommit, err := cIter.Next()
			if err != nil {
				return nil, err
			}

			// Skip commits which do not contain the author name provided
			if author != "" && !containsAuthor(currentCommit, author) {
				currentCommit = nextCommit
				commitTime = currentCommit.Author.When.UTC()
				continue
			}

			if short {
				// Multi-line commit messages span over newlines, taking the text before this is the main message and the rest can be discarded.
				firstLine, _, _ := strings.Cut(currentCommit.Message, "\n")
				msgs[dir] = append(msgs[dir], firstLine)
			} else {
				msgs[dir] = append(msgs[dir], currentCommit.Message)
			}

			currentCommit = nextCommit
			commitTime = currentCommit.Author.When.UTC()
		}
	}

	return msgs, nil
}

func printUsage() {
	var executableName string
	fullPath, err := os.Executable()
	if err != nil {
		executableName = "today"
	} else {
		executableName = filepath.Base(fullPath)
	}

	fmt.Fprintf(os.Stderr, "Usage: %s [options] git_directory...\n", executableName)
	flag.PrintDefaults()
}

func main() {

	flag.Usage = printUsage

	flag.BoolVar(&short, "short", false, "display the first line of commit messages only")
	flag.DurationVar(&since, "since", 12*time.Hour, "how far back to check for commits from now")
	flag.StringVar(&author, "author", "", "display commits from a particular author")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Missing mandatory argument: git_directory")
		printUsage()
		os.Exit(1)
	}

	// Directories must be tracked by git so that we can read commit messages and use this
	// as a guide on work done throughout a time period.
	err := validatePaths(flag.Args())
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	dirs := flag.Args()

	repos, err := getRepositories(dirs)
	if err != nil {
		fmt.Println(err)
		return
	}

	dirToRepo := make(map[string]*git.Repository)
	for i := 0; i < len(dirs); i++ {
		dirToRepo[dirs[i]] = repos[i]
	}

	msgs, err := getCommitMessages(dirToRepo, author, short, since)
	if err != nil {
		fmt.Println(err)
		return
	}

	for dir, commitMsgs := range msgs {
		fmt.Printf("%s\n", dir)

		if len(commitMsgs) == 0 {
			fmt.Printf("\tThere are no messages for this directory.\n")
		}
		for _, msg := range commitMsgs {
			fmt.Printf("\t%s\n", msg)
		}

		// Simple newline before the next entry.
		fmt.Println()
	}
}
