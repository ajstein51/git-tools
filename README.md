# Peddi Tooling CLI

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/ajstein51/git-tools/actions)
[![Version](https://img.shields.io/badge/version-1.0-blue)](https://github.com/astein-peddi/git-tooling/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Peddi Tooling is a command-line interface (CLI) designed to streamline and automate common Git and GitHub workflows at Peddinghaus. It provides powerful commands for managing GitHub Projects and comparing pull request statuses between branches, eliminating manual checks and saving developer time.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Projects (`projects`)](#projects-projects)
  - [Pull Requests (`prs`)](#pull-requests-prs)
- [Building from Source](#building-from-source)
- [License](#license)

## Features

- **Interactive Branch Comparison:** Interactively find which pull requests have been merged into one branch (e.g., `dev`) but not yet into another (e.g., `main` or `rtm`).
- **Powerful GitHub Projects Integration:** List and filter items in a GitHub Project, with specific commands to find issues with/without PRs, unmerged PRs, and items assigned to you for review.
- **Automatic Repository Detection:** The CLI is context-aware and automatically detects the current repository, requiring no manual configuration of owner or repo names.

## Installation

The tool is distributed via a Windows installer for easy setup.

1.  Navigate to the [**Releases**](https://github.com/astein-peddi/git-tooling/releases) page of the GitHub repository.
2.  Download the latest `peddi-tooling-installer.exe` file.
3.  Run the installer. It's a standard "Next, Next, Finish" installation that requires no special privileges. It will automatically:
    - Install the executable to `%APPDATA%\PeddiTooling`.
    - Safely add this directory to your user `PATH`.
    - Generate and configure the PowerShell tab completion script.

4.  **Important:** After the installation is complete, you **must close and reopen** any existing PowerShell windows for the changes to your `PATH` and profile to take effect.

## Configuration

This tool authenticates with GitHub using the standard GitHub CLI (`gh`). Before you can use commands that interact with the API (`projects`, `prs`), you must be authenticated.

If you haven't already, run the following command and follow the on-screen prompts:
```sh
winget install --id GitHub.cli -e

gh auth login
```

## Usage

The CLI is organized into a series of commands and subcommands.

### Projects (projects)

The projects command helps you interact with GitHub Projects (V2) linked to the current repository. By default, it operates on the most recently updated project, but you can target a specific project with the --id flag.

Base Command: 
    
```sh
peddi-tooling projects list <subcommand>
```

#### Available Subcommands:

#### all: 
List all cards (Issues, PRs, and Drafts) in the project.

```Sh
peddi-tooling projects list all
```

#### no-pr: 
List all items that do not have a linked Pull Request.

```Sh
peddi-tooling projects list no-pr
```

#### with-pr: 
List all items that are either a Pull Request or are linked to one.

```Sh
peddi-tooling projects list with-pr
```

#### pr-not-merged: 
List items that are associated with an open, unmerged Pull Request.

```Sh
peddi-tooling projects list pr-not-merged
```

#### reviewer: 
List items where you are a requested reviewer.

```Sh
peddi-tooling projects list reviewer
```

You can also check for PRs assigned to a specific teammate using the --name (or -n) flag:

```Sh
peddi-tooling projects list reviewer --name other-github-username
```

### Pull Requests (prs)
The prs command is designed to compare the state of two branches to understand what work is pending release.


Finds all pull requests that have been merged into branchA but whose changes are not yet present in branchB. It fetches the most recent PRs first and interactively prompts you to load more pages if necessary.

```sh
peddi-tooling prs <branchA> <branchB>
```

#### Available Subcommands:

#### limit
Limit the amount of commits to scan through. This starts from the most recent commits.

```Sh
peddi-tooling prs <branchA> <branchB> --page-size 1
```

#### local
Use local branches only

```Sh
peddi-tooling prs <branchA> <branchB> --page-size 1
```

#### page-size
Limits the quantity of prs displayed

```Sh
peddi-tooling prs <branchA> <branchB> --page-size 1
```

Example:

To see what has been merged into dev that has not yet been released to main, you would run:

```Sh
peddi-tooling prs dev main
```

Example Output:

```sh
Comparing recent PRs merged into 'dev' that are not yet in 'main'...

#402: Add new feature for production (https://github.com/org/repo/pull/402)

-- Showing 1-1 of 3. Use 'h' to navigate left, 'l' to navigate right, 'q' to quit --
```