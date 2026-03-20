package ui

import (
	"context"
	"errors"
	"testing"

	"github.com/megatherium/blunderbust/internal/domain"
)

func TestHandleErrMsg_SetsErrorState(t *testing.T) {
	model := NewTestModel()
	model.app = newTestApp()

	testErr := errors.New("test error")
	errMsg := errMsg{err: testErr}
	newModel, _ := model.handleErrMsg(errMsg)

	uiModel := newModel.(UIModel)
	if uiModel.state != ViewStateError {
		t.Errorf("Expected state ViewStateError, got %v", uiModel.state)
	}
	if uiModel.err == nil {
		t.Error("Expected error to be set")
	}
}

func TestHandleErrMsg_SetsRetryStore(t *testing.T) {
	model := NewTestModel()
	app := newTestApp()

	project := domain.Project{Dir: "/test/project", Name: "Test Project"}
	app.AddProject(project)

	ctx := context.Background()
	store, _ := app.CreateStore(ctx, "/test/project/.beads")
	app.AddStore("/test/project", store)

	if err := app.SetActiveProject(ctx, "/test/project"); err != nil {
		t.Fatalf("Failed to set active project: %v", err)
	}

	model.app = app

	testErr := errors.New("test error")
	errMsg := errMsg{err: testErr, showRetryOptions: true}
	newModel, _ := model.handleErrMsg(errMsg)

	uiModel := newModel.(UIModel)
	if uiModel.retryStore == nil {
		t.Error("Expected retryStore to be set")
	}
}

func TestHandleWarningMsg_AppendsWarning(t *testing.T) {
	model := NewTestModel()
	initialWarningCount := len(model.warnings)

	testErr := errors.New("test warning")
	msg := warningMsg{err: testErr}
	newModel, _ := model.handleWarningMsg(msg)

	uiModel := newModel.(UIModel)
	if len(uiModel.warnings) != initialWarningCount+1 {
		t.Errorf("Expected warnings count to increase by 1, got %d", len(uiModel.warnings))
	}
}

func TestHandleLaunchResult_Success_RegistersAgent(t *testing.T) {
	model := NewTestModel()
	model.selectedWorktree = "/path/to/worktree"
	model.selection.Ticket = domain.Ticket{ID: "ticket-1", Title: "Test Ticket"}
	model.selection.Harness = domain.Harness{Name: "test-harness"}
	model.selection.Model = "test-model"
	model.selection.Agent = "test-agent"
	model.agents = make(map[string]*RunningAgent)
	model.app = newTestApp()

	launchResult := &domain.LaunchResult{
		LauncherID:   "launcher-123",
		LauncherType: domain.LauncherTypeTmux,
	}
	launchSpec := &domain.LaunchSpec{
		Selection: model.selection,
	}

	msg := launchResultMsg{res: launchResult, spec: launchSpec}
	newModel, _ := model.handleLaunchResult(msg)

	uiModel := newModel.(UIModel)
	if uiModel.state != ViewStateMatrix {
		t.Errorf("Expected state ViewStateMatrix, got %v", uiModel.state)
	}
	if uiModel.agents == nil || len(uiModel.agents) == 0 {
		t.Error("Expected agent to be registered")
	}
}

func TestHandleLaunchResult_Error_SetsErrorState(t *testing.T) {
	model := NewTestModel()
	model.app = newTestApp()

	testErr := errors.New("test error")
	msg := launchResultMsg{err: testErr}
	newModel, _ := model.handleLaunchResult(msg)

	uiModel := newModel.(UIModel)
	if uiModel.state != ViewStateError {
		t.Errorf("Expected state ViewStateError, got %v", uiModel.state)
	}
}

func TestHandleLaunchResult_NilResult_ReturnsToMatrix(t *testing.T) {
	model := NewTestModel()
	model.app = newTestApp()
	model.state = ViewStateConfirm

	msg := launchResultMsg{res: nil, err: nil}
	newModel, _ := model.handleLaunchResult(msg)

	uiModel := newModel.(UIModel)
	if uiModel.state != ViewStateMatrix {
		t.Errorf("Expected state ViewStateMatrix, got %v", uiModel.state)
	}
}

func TestHandleWorktreeSelected_UpdatesFocus(t *testing.T) {
	model := NewTestModel()

	msg := WorktreeSelectedMsg{Path: "/test/path"}
	newModel, _ := model.handleWorktreeSelected(msg)

	uiModel := newModel.(UIModel)
	if uiModel.selectedWorktree != "/test/path" {
		t.Errorf("Expected selectedWorktree /test/path, got %s", uiModel.selectedWorktree)
	}
	if uiModel.focus != FocusTickets {
		t.Errorf("Expected focus FocusTickets, got %v", uiModel.focus)
	}
	if uiModel.sidebar.Focused() {
		t.Error("Expected sidebar to be unfocused")
	}
}

func TestHandleWorktreeSelected_DirtiesTicket(t *testing.T) {
	model := NewTestModel()
	model.dirtyTicket = false

	msg := WorktreeSelectedMsg{Path: "/test/path"}
	newModel, _ := model.handleWorktreeSelected(msg)

	uiModel := newModel.(UIModel)
	if !uiModel.dirtyTicket {
		t.Error("Expected dirtyTicket to be set")
	}
}
