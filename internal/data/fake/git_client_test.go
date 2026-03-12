// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fake

import (
	"context"
	"errors"
	"testing"

	"github.com/megatherium/blunderbust/internal/data"
)

var (
	errFakeGit = errors.New("fake git error")
)

func TestFakeGitClient_ListWorktrees_Default(t *testing.T) {
	client := NewFakeGitClient()
	entries, err := client.ListWorktrees(context.Background(), "/repo")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entries != nil {
		t.Errorf("expected nil entries (default), got %v", entries)
	}
}

func TestFakeGitClient_ListWorktrees_Configured(t *testing.T) {
	client := NewFakeGitClient()
	entries := []data.WorktreeEntry{
		{Path: "/repo", Commit: "abc123", Branch: "main"},
		{Path: "/repo/feature", Commit: "def456", Branch: "feature"},
	}
	client.SetWorktrees("/repo", entries)

	result, err := client.ListWorktrees(context.Background(), "/repo")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}

	if result[0].Path != "/repo" {
		t.Errorf("expected path /repo, got %s", result[0].Path)
	}

	if result[1].Branch != "feature" {
		t.Errorf("expected branch feature, got %s", result[1].Branch)
	}
}

func TestFakeGitClient_ListWorktrees_Error(t *testing.T) {
	client := NewFakeGitClient()
	client.SetError("listworktrees", "/repo", errFakeGit)

	_, err := client.ListWorktrees(context.Background(), "/repo")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != errFakeGit {
		t.Errorf("expected errFakeGit, got %v", err)
	}
}

func TestFakeGitClient_DetectMainBranch_Default(t *testing.T) {
	client := NewFakeGitClient()
	branch, err := client.DetectMainBranch(context.Background(), "/repo")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if branch != "" {
		t.Errorf("expected empty branch (default), got %s", branch)
	}
}

func TestFakeGitClient_DetectMainBranch_Configured(t *testing.T) {
	client := NewFakeGitClient()
	client.SetMainBranch("/repo", "develop")

	branch, err := client.DetectMainBranch(context.Background(), "/repo")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if branch != "develop" {
		t.Errorf("expected branch develop, got %s", branch)
	}
}

func TestFakeGitClient_DetectMainBranch_Error(t *testing.T) {
	client := NewFakeGitClient()
	client.SetError("detectmainbranch", "/repo", errFakeGit)

	_, err := client.DetectMainBranch(context.Background(), "/repo")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != errFakeGit {
		t.Errorf("expected errFakeGit, got %v", err)
	}
}

func TestFakeGitClient_CheckDirty_Default(t *testing.T) {
	client := NewFakeGitClient()
	dirty := client.CheckDirty(context.Background(), "/repo")

	if dirty {
		t.Error("expected dirty to be false (default)")
	}
}

func TestFakeGitClient_CheckDirty_Configured(t *testing.T) {
	client := NewFakeGitClient()
	client.SetDirty("/repo", true)
	client.SetDirty("/repo/feature", false)

	if !client.CheckDirty(context.Background(), "/repo") {
		t.Error("expected /repo to be dirty")
	}

	if client.CheckDirty(context.Background(), "/repo/feature") {
		t.Error("expected /repo/feature to not be dirty")
	}
}

func TestFakeGitClient_MultipleRepos(t *testing.T) {
	client := NewFakeGitClient()

	client.SetWorktrees("/repo1", []data.WorktreeEntry{{Path: "/repo1", Branch: "main"}})
	client.SetWorktrees("/repo2", []data.WorktreeEntry{{Path: "/repo2", Branch: "develop"}})

	client.SetMainBranch("/repo1", "main")
	client.SetMainBranch("/repo2", "develop")

	client.SetDirty("/repo1", false)
	client.SetDirty("/repo2", true)

	entries1, _ := client.ListWorktrees(context.Background(), "/repo1")
	entries2, _ := client.ListWorktrees(context.Background(), "/repo2")

	branch1, _ := client.DetectMainBranch(context.Background(), "/repo1")
	branch2, _ := client.DetectMainBranch(context.Background(), "/repo2")

	dirty1 := client.CheckDirty(context.Background(), "/repo1")
	dirty2 := client.CheckDirty(context.Background(), "/repo2")

	if len(entries1) != 1 || entries1[0].Branch != "main" {
		t.Error("repo1 worktree mismatch")
	}

	if len(entries2) != 1 || entries2[0].Branch != "develop" {
		t.Error("repo2 worktree mismatch")
	}

	if branch1 != "main" {
		t.Error("repo1 main branch mismatch")
	}

	if branch2 != "develop" {
		t.Error("repo2 main branch mismatch")
	}

	if dirty1 {
		t.Error("repo1 should not be dirty")
	}

	if !dirty2 {
		t.Error("repo2 should be dirty")
	}
}

func TestFakeGitClient_ErrorWildcard(t *testing.T) {
	client := NewFakeGitClient()
	client.SetError("listworktrees", "*", errFakeGit)

	_, err := client.ListWorktrees(context.Background(), "/repo1")
	if err != errFakeGit {
		t.Errorf("expected wildcard error, got %v", err)
	}

	_, err = client.ListWorktrees(context.Background(), "/repo2")
	if err != errFakeGit {
		t.Errorf("expected wildcard error, got %v", err)
	}
}
