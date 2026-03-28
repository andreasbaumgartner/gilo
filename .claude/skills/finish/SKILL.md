---
name: finish
description: Finish this feature branch by creating a PR, closing the issue, cleaning up the worktree, and closing the tmux window
disable-model-invocation: true
---

Finish this feature branch by creating a PR, closing the issue, cleaning up the worktree, and closing the tmux window.

Follow these steps exactly. If any step fails, stop and report the error — do not continue to subsequent steps.

## Step 1: Extract context

Run `git branch --show-current` to get the branch name. The branch follows the pattern `issue-{number}-{slug}` (e.g., `issue-23-auto-finish-skill-or-command`). Extract the issue number from the branch name.

Run `git rev-parse --show-toplevel` to get the current worktree path. Store this as `WORKTREE_PATH`.

Determine the main repository path: the worktree lives at `{repo_parent}/.worktrees/{branch}/`, so the main repo is a sibling directory of `.worktrees/`. To find it reliably, run `git -C $(git rev-parse --show-toplevel) rev-parse --path-format=absolute --git-common-dir` which returns the main repo's `.git` directory — go up one level to get the main repo root. Store this as `MAIN_REPO`.

## Step 2: Handle uncommitted changes

Run `git status --porcelain`. If there are any changes (staged, unstaged, or untracked):
- Stage all relevant changes with `git add -A`
- Create a commit with an appropriate message describing the changes
- If unsure what the changes are, run `git diff --staged --stat` to summarize them

If the working tree is clean, skip this step.

## Step 3: Push the branch

Run `git push -u origin HEAD` to push the branch to the remote. If the branch is already up to date, that is fine.

## Step 4: Create a Pull Request

Run `gh pr create --fill --head {branch_name}` to create a PR that targets the default branch. The `--fill` flag auto-populates the title and body from commit messages.

If a PR already exists for this branch, `gh` will report it — note the PR URL and continue. If PR creation fails for another reason, stop and report the error.

Display the PR URL to the user.

## Step 5: Close the GitHub issue

Run `gh issue close {issue_number}`. If the issue is already closed, that is fine — continue.

## Step 6: Remove the worktree

This step requires leaving the worktree directory first since you cannot remove a worktree while inside it.

Run these commands:
```
cd {MAIN_REPO}
git worktree remove {WORKTREE_PATH} --force
```

## Step 7: Close the tmux window

Check if running inside tmux by checking the `TMUX` environment variable.

If inside tmux, run `tmux kill-window` to close the current window. This returns the user to the gilo TUI window. Note: this will terminate the current Claude session — that is expected and intentional.

If not inside tmux, inform the user that the finish process is complete and they can close the terminal manually.
