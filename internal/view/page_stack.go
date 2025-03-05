package view

import (
	"github.com/isan-rivkin/surf/internal/model"
	"github.com/isan-rivkin/surf/internal/ui"
)

// PageStack represents a stack of pages.
type PageStack struct {
	*ui.Pages

	app *AppController
}

// NewPageStack returns a new page stack.
func NewPageStack() *PageStack {
	return &PageStack{
		Pages: ui.NewPages(),
	}
}

// Init initializes the view.
func (p *PageStack) Init(app *AppController) (err error) {
	p.app = app
	p.Stack.AddListener(p)

	return nil
}

// StackPushed notifies a new page was added.
func (p *PageStack) StackPushed(c model.Component) {
	c.Start()
	p.app.SetFocus(c)
}

// StackPopped notifies a page was removed.
func (p *PageStack) StackPopped(o, top model.Component) {
	o.Stop()
	p.StackTop(top)
}

// StackTop notifies for the top component.
func (p *PageStack) StackTop(top model.Component) {
	if top == nil {
		return
	}
	top.Start()
	p.app.SetFocus(top)
}
