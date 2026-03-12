// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package data

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GitClient abstracts git operations for worktree discovery.
// This interface enables testing with fake implementations and supports
// alternative backends (e.g., worktree manager tools, non-git systems).
type GitClient interface {
	ListWorktrees(ctx context.Context, repoRoot string) ([]WorktreeEntry, error)
	DetectMainBranch(ctx context.Context, repoRoot string) (string, error)
	CheckDirty(ctx context.Context, path string) bool
}

// WorktreeEntry represents a single worktree from git worktree list output.
type WorktreeEntry struct {
	Path   string
	Commit string
	Branch string
}

// gitClient implements GitClient by executing actual git commands.
type gitClient struct{}

// NewGitClient creates a new gitClient for real git operations.
func NewGitClient() GitClient {
	return &gitClient{}
}

func (g *gitClient) ListWorktrees(ctx context.Context, repoRoot string) ([]WorktreeEntry, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		if isNotGitRepo(err) {
			return nil, fmt.Errorf("not a git repository: %s", repoRoot)
		}
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	return parseWorktreePorcelain(output), nil
}

func (g *gitClient) DetectMainBranch(ctx context.Context, repoRoot string) (string, error) {
	for _, candidate := range []string{"main", "master", "develop"} {
		cmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "rev-parse", "--verify", "refs/heads/"+candidate)
		if err := cmd.Run(); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no main branch detected")
}

func (g *gitClient) CheckDirty(ctx context.Context, path string) bool {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(bytes.TrimSpace(output)) > 0
}

// parseWorktreePorcelain parses the output of `git worktree list --porcelain`.
// Each worktree is separated by an empty line, with fields in key-value format.
func parseWorktreePorcelain(output []byte) []WorktreeEntry {
	var worktrees []WorktreeEntry
	var current *WorktreeEntry

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "worktree ") {
			if current != nil {
				worktrees = append(worktrees, *current)
			}
			current = &WorktreeEntry{
				Path: strings.TrimPrefix(line, "worktree "),
			}
		} else if current != nil {
			if strings.HasPrefix(line, "HEAD ") {
				current.Commit = strings.TrimPrefix(line, "HEAD ")
			} else if strings.HasPrefix(line, "branch ") {
				branchRef := strings.TrimPrefix(line, "branch ")
				current.Branch = extractBranchName(branchRef)
			}
		}
	}

	if current != nil {
		worktrees = append(worktrees, *current)
	}

	return worktrees
}

// extractBranchName extracts the branch name from a git reference.
// For example, "refs/heads/main" becomes "main".
func extractBranchName(ref string) string {
	if strings.HasPrefix(ref, "refs/heads/") {
		return strings.TrimPrefix(ref, "refs/heads/")
	}
	return ref
}

// isNotGitRepo checks if an error indicates that a path is not a git repository.
func isNotGitRepo(err error) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode() == 128
	}
	return false
}
