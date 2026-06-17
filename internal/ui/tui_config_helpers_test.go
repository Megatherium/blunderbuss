package ui

import (
	"testing"

	"github.com/megatherium/blunderbust/internal/app"
	"github.com/megatherium/blunderbust/internal/config"
	"github.com/megatherium/blunderbust/internal/domain"
)

// plainLoader implements only config.Loader (not config.TUILoader), used to
// exercise the type-assertion failure branch in loadTUIConfig/saveTUIConfig.
type plainLoader struct{}

func (p *plainLoader) Load(path string) (*domain.Config, error)   { return &domain.Config{}, nil }
func (p *plainLoader) Save(path string, cfg *domain.Config) error { return nil }

func TestLoadTUIConfig_NilApp(t *testing.T) {
	cfg, err := loadTUIConfig(nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil config, got %+v", cfg)
	}
}

func TestLoadTUIConfig_EmptyPath(t *testing.T) {
	a := &app.App{}
	cfg, err := loadTUIConfig(a)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil config for empty path, got %+v", cfg)
	}
}

func TestLoadTUIConfig_NonTUILoader(t *testing.T) {
	a := &app.App{
		Loader: &plainLoader{},
		Opts:   domain.AppOptions{TUIConfigPath: "/tmp/fake.yaml"},
	}
	cfg, err := loadTUIConfig(a)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil config for non-TUILoader, got %+v", cfg)
	}
}

func TestSaveTUIConfig_NilApp(t *testing.T) {
	if err := saveTUIConfig(nil, &config.TUIConfig{}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSaveTUIConfig_EmptyPath(t *testing.T) {
	a := &app.App{}
	if err := saveTUIConfig(a, &config.TUIConfig{}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSaveTUIConfig_NonTUILoader(t *testing.T) {
	a := &app.App{
		Loader: &plainLoader{},
		Opts:   domain.AppOptions{TUIConfigPath: "/tmp/fake.yaml"},
	}
	if err := saveTUIConfig(a, &config.TUIConfig{}); err != nil {
		t.Fatalf("expected nil error for non-TUILoader, got %v", err)
	}
}

// Compile-time: plainLoader satisfies config.Loader but NOT config.TUILoader.
var _ config.Loader = (*plainLoader)(nil)
