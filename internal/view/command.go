package view

import (
	"sync"

	"github.com/isan-rivkin/surf/internal/model"
)

// Command represents a user command.
type Command struct {
	app *AppController
	// used to manage aliases for same command
	// alias *dao.Alias
	mx sync.Mutex
}

// NewCommand returns a new command.
func NewCommand(app *AppController) *Command {
	return &Command{
		app: app,
	}
}
func (c *Command) Init(initCtx model.Provider) error {
	//1. TODO update ctx, for example for AWS load use the profile and all the details
	return nil
}

func (c *Command) defaultCmd() error {
	//TODO run default command, i.e list default resources for provider (in aws list ec2 instances)
	return nil
}
