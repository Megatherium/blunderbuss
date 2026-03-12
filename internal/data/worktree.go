// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package data

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/megatherium/blunderbust/internal/domain"
)

// WorktreeDiscoverer discovers worktrees in a git repository using a GitClient.
type WorktreeDiscoverer struct {
	gitClient GitClient
}

// NewWorktreeDiscoverer creates a new WorktreeDiscoverer with the given GitClient.
// If gitClient is nil, a default gitClient is used.
func NewWorktreeDiscoverer(gitClient GitClient) *WorktreeDiscoverer {
	if gitClient == nil {
		gitClient = NewGitClient()
	}
	return &WorktreeDiscoverer{
		gitClient: gitClient,
	}
}

// Discover discovers all worktrees in the given git repository.
// It returns a slice of WorktreeInfo with metadata for each worktree.
func (d *WorktreeDiscoverer) Discover(ctx context.Context, repoRoot string) ([]domain.WorktreeInfo, error) {
	entries, err := d.gitClient.ListWorktrees(ctx, repoRoot)
	if err != nil {
		return nil, err
	}

	mainBranch, _ := d.gitClient.DetectMainBranch(ctx, repoRoot)

	results := make([]domain.WorktreeInfo, 0, len(entries))
	for _, wt := range entries {
		// Fallback for legacy repos where DetectMainBranch fails to detect the main branch.
		// This ensures backward compatibility with repos using "main" or "master" conventions.
		isMain := wt.Branch == mainBranch || wt.Branch == "master" || wt.Branch == "main"
		info := domain.WorktreeInfo{
			Path:       wt.Path,
			Branch:     wt.Branch,
			CommitHash: wt.Commit,
			IsMain:     isMain,
		}

		info.IsDirty = d.gitClient.CheckDirty(ctx, wt.Path)
		info.Name = d.extractName(wt.Path, isMain)

		results = append(results, info)
	}

	return results, nil
}

func (d *WorktreeDiscoverer) extractName(path string, isMain bool) string {
	if isMain {
		return "main"
	}
	return filepath.Base(path)
}

// FindRepoRoot finds the root of the git repository containing the given path.
func FindRepoRoot(startPath string) (string, error) {
	path, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	for {
		gitDir := filepath.Join(path, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return path, nil
		}

		parent := filepath.Dir(path)
		if parent == path {
			return "", fmt.Errorf("not a git repository: %s", startPath)
		}
		path = parent
	}
}

// GetProjectName returns the base name of the repository root directory.
func GetProjectName(repoRoot string) string {
	return filepath.Base(repoRoot)
}
