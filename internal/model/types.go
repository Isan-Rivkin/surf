package model

import (
	"context"

	"github.com/rivo/tview"
)

// Igniter represents a runnable view.
type Igniter interface {
	// Start starts a component.
	Init(ctx context.Context) error

	// Start starts a component.
	Start()

	// Stop terminates a component.
	Stop()
}

// TODO(isan) what is hiter for? do I need it for menu?
// Hinter represent a menu mnemonic provider.
type Hinter interface {
	// Hints returns a collection of menu hints.
	// Hints() MenuHints

	// ExtraHints returns additional hints.
	// ExtraHints() map[string]string
}

// Primitive represents a UI primitive.
type Primitive interface {
	tview.Primitive

	// Name returns the view name.
	Name() string
}

// Commander tracks prompt status.
type Commander interface {
	// InCmdMode checks if prompt is active.
	InCmdMode() bool
}

// Component represents a ui component.
type Component interface {
	Primitive
	Igniter
	Hinter
	Commander
	Filterer
}

type Filterer interface {
	SetFilter(string)
	SetLabelFilter(map[string]string)
}
