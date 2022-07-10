# today

Easily view the work that you have done for the day, at a glance, using your `git commit` history. This is ideal for a daily standup where you want to see what you have done in the past.

This leverages your `git log`, so works best when paired with clear and concise commit messages. A great example of this is [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/#summary).

## Install

The easiest way to install is through the `go` command.

```bash
go install github.com/jdockerty/today@latest # or tag
```

Or by downloading a pre-compiled binary on the [releases](https://github.com/jdockerty/today/releases) page.


## Usage

Simply pass a local directory that you wish to view the work for, this also works over multiple directories, as work is not always confined to a single project.

```bash
today work/very-important-business-app

today work/api work/frontend
```
