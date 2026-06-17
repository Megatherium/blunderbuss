package ui

import (
	"context"
	"time"

	"github.com/megatherium/blunderbust/internal/domain"
)

// context returns the app's cancellable context for monitoring commands, or
// context.Background() when the app is nil (e.g. in unit tests that use
// NewTestModel). This avoids nil-pointer panics in test setups while keeping
// real cancellation in production.
func (m UIModel) context() context.Context {
	if m.app == nil {
		return context.Background()
	}
	return m.app.Context()
}

// NewTestModel creates a minimal UIModel for testing purposes
func NewTestModel() *UIModel {
	m := UIModel{
		app:       nil,
		state:     ViewStateMatrix,
		focus:     FocusSidebar,
		selection: domain.Selection{},
		sidebar:   NewSidebarModel(),
		animState: AnimationState{
			StartTime:       time.Now(),
			ColorCycleStart: time.Now(),
			CurrentThemeIdx: 0,
		},
	}
	return &m
}
