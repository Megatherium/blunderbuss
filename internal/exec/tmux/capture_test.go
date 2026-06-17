// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package tmux

import (
	"context"
	"strings"
	"testing"
)

func TestOutputCapture_StartStop(t *testing.T) {
	fake := NewFakeRunner()
	capture := NewOutputCapture(fake, "@123")

	ctx := context.Background()
	path, err := capture.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if path != "" {
		t.Errorf("Start() returned path %s, expected empty string", path)
	}

	err = capture.Stop(ctx)
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestOutputCapture_ReadOutput(t *testing.T) {
	fake := NewFakeRunner()
	capture := NewOutputCapture(fake, "@123")

	// The fake runner returns this when ANY command is run.
	testOutput := []byte("Hello from capture-pane")
	fake.AlwaysReturn = testOutput

	content, err := capture.ReadOutput(context.Background())
	if err != nil {
		t.Fatalf("ReadOutput() error = %v", err)
	}

	if string(content) != string(testOutput) {
		t.Errorf("ReadOutput() = %q, want %q", string(content), string(testOutput))
	}

	// Verify the right command was executed
	if len(fake.Commands) == 0 {
		t.Fatal("No commands captured")
	}

	foundCapturePane := false
	for _, cmd := range fake.Commands {
		if strings.Contains(cmd, "capture-pane") && strings.Contains(cmd, "-t @123") && strings.Contains(cmd, "-p") {
			foundCapturePane = true
			break
		}
	}
	if !foundCapturePane {
		t.Error("capture-pane command not found in executed commands")
	}
}

func TestOutputCapture_ReadOutput_EmptyWindow(t *testing.T) {
	fake := NewFakeRunner()
	capture := NewOutputCapture(fake, "")

	if _, err := capture.ReadOutput(context.Background()); err == nil {
		t.Error("Expected error for empty window id, got nil")
	}
}

func TestOutputCapture_Scrollback(t *testing.T) {
	fake := NewFakeRunner()
	fake.AlwaysReturn = []byte("history")
	capture := NewOutputCaptureWithScrollback(fake, "@9", 250)

	if _, err := capture.ReadOutput(context.Background()); err != nil {
		t.Fatalf("ReadOutput() error = %v", err)
	}

	if len(fake.Commands) == 0 {
		t.Fatal("no command recorded")
	}
	if !strings.Contains(fake.Commands[0], "-S -250") {
		t.Errorf("expected scrollback flag -S -250 in command, got %q", fake.Commands[0])
	}
}

func TestOutputCapture_ZeroScrollbackOmitsFlag(t *testing.T) {
	fake := NewFakeRunner()
	fake.AlwaysReturn = []byte("visible")
	capture := NewOutputCaptureWithScrollback(fake, "@9", 0)

	if _, err := capture.ReadOutput(context.Background()); err != nil {
		t.Fatalf("ReadOutput() error = %v", err)
	}
	if strings.Contains(fake.Commands[0], "-S") {
		t.Errorf("zero scrollback must not emit -S, got %q", fake.Commands[0])
	}
}

func TestCaptureFactory_NewCapture(t *testing.T) {
	fake := NewFakeRunner()
	fake.AlwaysReturn = []byte("captured")
	factory := NewCaptureFactory(fake)

	capture := factory.NewCapture("@7")
	if capture == nil {
		t.Fatal("NewCapture returned nil for non-empty id")
	}

	if _, err := capture.ReadOutput(context.Background()); err != nil {
		t.Fatalf("ReadOutput() error = %v", err)
	}

	if len(fake.Commands) == 0 {
		t.Fatal("no command recorded")
	}
	if !strings.Contains(fake.Commands[0], "-t @7") {
		t.Errorf("capture not bound to @7, got %q", fake.Commands[0])
	}
}

func TestCaptureFactory_EmptyIDReturnsNil(t *testing.T) {
	fake := NewFakeRunner()
	factory := NewCaptureFactory(fake)

	if c := factory.NewCapture(""); c != nil {
		t.Errorf("expected nil capture for empty id, got %v", c)
	}
}

func TestCaptureFactory_NilSafe(t *testing.T) {
	var factory *CaptureFactory
	if c := factory.NewCapture("@1"); c != nil {
		t.Errorf("expected nil capture from nil factory, got %v", c)
	}
}
