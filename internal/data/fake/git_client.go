// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fake

import (
	"context"

	"github.com/megatherium/blunderbust/internal/data"
)

// fakeGitClient is an in-memory fake implementing data.GitClient.
type fakeGitClient struct {
	worktrees  map[string][]data.WorktreeEntry
	mainBranch map[string]string
	dirty      map[string]bool
	errors     map[string]error
}

// Verify interface compliance at compile time.
var _ data.GitClient = (*fakeGitClient)(nil)

// NewFakeGitClient creates a new fakeGitClient with empty default values.
func NewFakeGitClient() *fakeGitClient {
	return &fakeGitClient{
		worktrees:  make(map[string][]data.WorktreeEntry),
		mainBranch: make(map[string]string),
		dirty:      make(map[string]bool),
		errors:     make(map[string]error),
	}
}

// ListWorktrees returns the configured worktrees for the given repo root.
// For unconfigured repos, returns nil (not an error), matching behavior
// when a real git repo has no worktrees beyond the main one.
func (f *fakeGitClient) ListWorktrees(ctx context.Context, repoRoot string) ([]data.WorktreeEntry, error) {
	if err := f.getError("listworktrees", repoRoot); err != nil {
		return nil, err
	}
	return f.worktrees[repoRoot], nil
}

// DetectMainBranch returns the configured main branch for the given repo root.
func (f *fakeGitClient) DetectMainBranch(ctx context.Context, repoRoot string) (string, error) {
	if err := f.getError("detectmainbranch", repoRoot); err != nil {
		return "", err
	}
	return f.mainBranch[repoRoot], nil
}

// CheckDirty returns the configured dirty state for the given path.
func (f *fakeGitClient) CheckDirty(ctx context.Context, path string) bool {
	return f.dirty[path]
}

// SetWorktrees configures the worktrees for a specific repo root.
func (f *fakeGitClient) SetWorktrees(repoRoot string, entries []data.WorktreeEntry) {
	f.worktrees[repoRoot] = entries
}

// SetMainBranch configures the main branch for a specific repo root.
func (f *fakeGitClient) SetMainBranch(repoRoot, branch string) {
	f.mainBranch[repoRoot] = branch
}

// SetDirty configures the dirty state for a specific path.
func (f *fakeGitClient) SetDirty(path string, isDirty bool) {
	f.dirty[path] = isDirty
}

// SetError configures an error to be returned for a specific operation and path.
func (f *fakeGitClient) SetError(operation, path string, err error) {
	f.errors[operation+":"+path] = err
}

func (f *fakeGitClient) getError(operation, path string) error {
	if err := f.errors[operation+":"+path]; err != nil {
		return err
	}
	if err := f.errors[operation+":*"]; err != nil {
		return err
	}
	return nil
}
