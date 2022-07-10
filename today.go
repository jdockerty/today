package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Since is a flag used to control the amount of time to look back in a repository for commits.
// The provided time units must be parseable via time.ParseDuration and it defaults to 12 hours.
var since time.Duration

// Short is a flag for discarding larger commit messages, this will only display the first line of a commit message.
// This is ideal for repositories where commits may contain longer explanations or reasoning behind the change, but you are familiar with it already.
var short bool

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

func getCommitMessages(dirToRepo map[string]*git.Repository, since time.Duration) (map[string][]string, error) {

	msgs := make(map[string][]string)

	for dir, repo := range dirToRepo {

		// Initialise map before populating messages.
		// This largely comes in handy when a directory is passed where there are no messages in the given 'since' range
		// so it can be displayed as no messages, as opposed to no output whatsoever.
		msgs[dir] = []string{}

		ref, err := repo.Head()
		if err != nil {
			return nil, err
		}
		cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			return nil, err
		}

		// Iterate from HEAD ref
		// TODO: Very large repositories/monorepos would cause lag here as we iterate all past commits.
		// 'Lower depth clone' style would help here, although we're working with the local directory, meaning this might be down to the use, similar to `git clone --depth=1 <repo>` to not get the entire history.
		err = cIter.ForEach(func(c *object.Commit) error {
			now := time.Now().UTC()
			commitTime := c.Author.When.UTC()

			// The UTC time of now - the provided 'since' value.
			// We use time.Add with a negative number to subtract here, rather than time.Sub, so that we produce a time.Time value to compare, not a time.Duration.
			timeSince := now.Add(-since)

			// If time of commit is 12 hours (or given value) after current time, add it to the map.
			if commitTime.After(timeSince) {
				if short {
					firstLine, _, _ := strings.Cut(c.Message, "\n")
					msgs[dir] = append(msgs[dir], firstLine)
				} else {
					msgs[dir] = append(msgs[dir], c.Message)
				}

				return nil
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return msgs, nil
}

func main() {

	flag.BoolVar(&short, "short", false, "display the first line of commit messages only")
	flag.DurationVar(&since, "since", 12*time.Hour, "how far back to check for commits from now")
	flag.Parse()

	// Directories must be tracked by git so that we can read commit messages and use this
	// as a guide on work done throughout a time period.
	err := validatePaths(flag.Args())
	if err != nil {
		fmt.Println(err)
		return
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

	msgs, err := getCommitMessages(dirToRepo, since)
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
	}
}
