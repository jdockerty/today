package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Concept:
// cmd to start = 'today dir(s)'
// collate messages from `git log` (or better plumbing command) from past 12 hours (default)
// output messages for work done today

// Default to searching 12 hours of commits for each repository given.
var defaultSince = 12 * time.Hour

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

func main() {

	flag.Parse()

	// Directories must be tracked by git so that we can read commit messages and use this
	// as a guide on work done throughout a time period.
	err := validatePaths(flag.Args())
	if err != nil {
		fmt.Println(err)
		return
	}

	dirs := flag.Args()

	var repos []*git.Repository
	for _, dir := range dirs {
		repo, err := openGitDir(dir)
		if err != nil {
			fmt.Printf("Unable to open local directory '%s': %s\n", dir, err)
			return
		}

		repos = append(repos, repo)
	}

	for _, repo := range repos {

		// Using current HEAD ref means that we get information about the branch that is currently
		// being pointed to by git, this might not always be main/master.
		ref, err := repo.Head()
		if err != nil {
			fmt.Println(err)
		}
		cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			fmt.Println(err)
		}

		// Iterate from HEAD ref
		// TODO: Very large repositories/monorepos would cause lag here as we iterate all past commits.
		// 'Lower depth clone' style would help here, although we're working with the local directory, meaning this might be down to the use, similar to `git clone --depth=1 <repo>` to not get the entire history.
		err = cIter.ForEach(func(c *object.Commit) error {
			now := time.Now().UTC()
			commitTime := c.Author.When.UTC()

			// The UTC time of now - the provided 'since' value.
			// We use time.Add with a negative number to subtract here, rather than time.Sub, so that we produce a time.Time value to compare, not a time.Duration.
			timeSince := now.Add(-defaultSince)

			// If time of commit is 12 hours (or given value) after current time, then it can be displayed.
			if commitTime.After(timeSince) {
				fmt.Println(c.Message)
				return nil
			}

			return nil
		})

		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}
func main() {

	flag.Parse()

	// Directories must be tracked by git so that we can read commit messages and use this
	// as a guide on work done throughout a time period.
	err := validatePaths(flag.Args())
	if err != nil {
		fmt.Println(err)
		return
	}

	dirs := flag.Args()

	var repos []*git.Repository
	for _, dir := range dirs {
		repo, err := openGitDir(dir)
		if err != nil {
			fmt.Printf("Unable to open local directory '%s': %s\n", dir, err)
			return
		}

		repos = append(repos, repo)
	}

	msgs := getCommitMessages(repos)
	fmt.Println(msgs)
}
