package ui

import (
	"fmt"
	"os"

	"github.com/megatherium/blunderbust/internal/app"
	"github.com/megatherium/blunderbust/internal/config"
)

// loadTUIConfig loads TUI configuration via the App's Loader, using the
// config.TUILoader interface instead of calling package-level free functions.
// Returns nil config (no error) when the app is nil, has no TUI config path,
// or the loader does not implement TUILoader (the last case logs a warning to
// stderr so test/dev setups with a plain Loader are not silently no-op'd).
func loadTUIConfig(a *app.App) (*config.TUIConfig, error) {
	if a == nil || a.Opts.TUIConfigPath == "" {
		return nil, nil
	}
	tuiLoader, ok := a.Loader.(config.TUILoader)
	if !ok {
		fmt.Fprintf(os.Stderr, "Warning: loader %T does not implement config.TUILoader; TUI config disabled\n", a.Loader)
		return nil, nil
	}
	return tuiLoader.LoadTUI(a.Opts.TUIConfigPath)
}

// saveTUIConfig saves TUI configuration via the App's Loader, using the
// config.TUILoader interface. Returns nil error when the app is nil, has no
// TUI config path, or the loader does not implement TUILoader (the last case
// logs a warning to stderr so test/dev setups with a plain Loader are not
// silently no-op'd).
func saveTUIConfig(a *app.App, cfg *config.TUIConfig) error {
	if a == nil || a.Opts.TUIConfigPath == "" {
		return nil
	}
	tuiLoader, ok := a.Loader.(config.TUILoader)
	if !ok {
		fmt.Fprintf(os.Stderr, "Warning: loader %T does not implement config.TUILoader; TUI config disabled\n", a.Loader)
		return nil
	}
	return tuiLoader.SaveTUI(a.Opts.TUIConfigPath, cfg)
}
