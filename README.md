# today

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/jdockerty/today?color=blue)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/jdockerty/today?color=black)

View your commit history across multiple directories, ideal for daily standup.

Easily view the work that you have done for the day, or longer, leveraging the power of `git` tracking one or more repositories. This tool is simple to use and requires no extra setup, it simply utilises an pre-existing workflow that you are already familiar with.

This works best when paired with clear and concise commit messages. A great example of this is [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/#summary).

## Install

The easiest way to install is through the `go` command.

```bash
go install github.com/jdockerty/today@latest # or tag, e.g. @v0.1.2
```

Or by downloading a pre-compiled binary on the [releases](https://github.com/jdockerty/today/releases) page.

## Usage

Simply pass one or more directories that you wish to view the commits for.

```bash
today ./ # View the current directory

today work/api work/frontend work/new-important-serivce # You've been very busy

today --since 48h work/api # You missed standup yesterday

today --short work/fun-poc # Only display first line of the commit message

today --author "Jack" projects/backend-api # View commits with author name containing 'Jack'
```

You can always call `today --help` or `today -h` to view the default help at any time.

### Flag Options

* `--author` can be used to change which commits are displayed, based on a particular author.
    * **The default is to display all authors.**
    * This filter is done using [`strings.Contains`](https://pkg.go.dev/strings#Contains). As such, multiple authors may be displayed depending on the value provided.
    * This allows you to filter for your own or someone else's commits.
* `--colour` can be used to show a preset colourised output to the terminal. Directories which have no commits are shown in red, whilst others which do have commits green.
    * **The default is no colour.**
* `--short` can be used to display only the first line of a commit.
    * **The default is to display the entire commit message.**
    * Useful when commit messages are incredibly descriptive, spanning below the fold to explain the intention of a change.
    * This also has a side effect of reducing verbosity.
* `--since` can be used to modify the time range.
    * **The default is 12 hours, given in the format of 12h00m00s**.
    * Valid time units must conform to [`time.ParseDuration`](https://pkg.go.dev/time#ParseDuration).

