package view

import (
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/isan-rivkin/surf/internal/model"
	"github.com/isan-rivkin/surf/internal/ui"
	"github.com/rivo/tview"
)

const (
	splashDelay = 1 * time.Second
)

type App struct {
	*tview.Application
	// Root view that holds everything
	Main    *ui.Pages
	actions *ui.KeyActions
	mx      sync.RWMutex
	running bool
	// executes all prompt commands, goToResource, etc.
	command *Command
}

func NewApp() *App {
	return &App{
		Application: tview.NewApplication(),
		Main:        ui.NewPages(),
		actions:     ui.NewKeyActions(),
	}
}
func (a *App) Init() {
	a.bindKeys()
	a.SetRoot(a.Main, true)
}

func (a *App) bindKeys() {
	a.actions = ui.NewKeyActionsFromMap(ui.KeyMap{
		ui.KeyColon:    ui.NewKeyAction("Cmd", a.activateCmd, false),
		tcell.KeyCtrlR: ui.NewKeyAction("Redraw", a.redrawCmd, false),
		// tcell.KeyCtrlP: ui.NewKeyAction("Persist", a.saveCmd, false),
		// tcell.KeyCtrlC: ui.NewKeyAction("Quit", a.quitCmd, false),
		// tcell.KeyCtrlU: ui.NewSharedKeyAction("Clear Filter", a.clearCmd, false),
		// tcell.KeyCtrlQ: ui.NewSharedKeyAction("Clear Filter", a.clearCmd, false),
	})
}

// RedrawCmd forces a redraw.
func (a *App) redrawCmd(evt *tcell.EventKey) *tcell.EventKey {
	a.QueueUpdateDraw(func() {})
	return evt
}

// TODO ':' event for prompt  ActivateCmd toggle command mode.
func (a *App) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	// if a.InCmdMode() {
	// 	return evt
	// }
	// a.ResetPrompt(a.cmdBuff)
	// a.cmdBuff.ClearText(true)

	return nil
}

// HasAction checks if key matches a registered binding.
func (a *App) HasAction(key tcell.Key) (ui.KeyAction, bool) {
	return a.actions.Get(key)
}

// GetActions returns a collection of actions.
func (a *App) GetActions() *ui.KeyActions {
	return a.actions
}

// AddActions returns the application actions.
func (a *App) AddActions(aa *ui.KeyActions) {
	a.actions.Merge(aa)
}

// SetRunning sets the app run state.
func (a *App) SetRunning(f bool) {
	a.mx.Lock()
	defer a.mx.Unlock()
	a.running = f
}

// IsRunning checks if app is actually running.
func (a *App) IsRunning() bool {
	a.mx.RLock()
	defer a.mx.RUnlock()
	return a.running
}

// QueueUpdate queues up a ui action.
func (a *App) QueueUpdate(f func()) {
	if a.Application == nil {
		return
	}
	go func() {
		a.Application.QueueUpdate(f)
	}()
}

// QueueUpdateDraw queues up a ui action and redraw the ui.
func (a *App) QueueUpdateDraw(f func()) {
	if a.Application == nil {
		return
	}
	go func() {
		a.Application.QueueUpdateDraw(f)
	}()
}

type AppController struct {
	*App
	version string
	// the resources table main content view (added to Main with all other components such as prompt)
	Content *PageStack
	// should be used for suggestions and (more ?)
	cmdHistory *model.History
	// get's updated by Browser object (ResourceViewer) in resource itself, accessible since same package...
	filterHistory *model.History
}

func NewAppController(version string) *AppController {
	return &AppController{
		App:           NewApp(),
		version:       version,
		cmdHistory:    model.NewHistory(model.MaxHistory),
		filterHistory: model.NewHistory(model.MaxHistory),
		Content:       NewPageStack(),
	}
}

func (ac *AppController) Init() error {
	if err := ac.Content.Init(ac); err != nil {
		return err
	}
	// TODO add listeners here for stack events i.e crumbs should change when stack changes
	// ac.Content.Stack.AddListener(a.Crumbs())
	// ac.Content.Stack.AddListener(a.Menu())
	ac.App.Init()
	// ac.keyboard will be called and handle any key events and forward None, same or differnt key
	// to main Application routine to continue
	ac.SetInputCapture(ac.keyboard)
	// bind control high level keys
	ac.bindKeys()
	ac.command = NewCommand(ac)
	// TODO init command replace mock provider with actual aws for example that shoud set initial context
	if err := ac.command.Init(&model.MockProvider{}); err != nil {
		return err
	}

	ac.layout()
	return nil
}

func (ac *AppController) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if k, ok := ac.HasAction(ui.AsKey(evt)); ok && !ac.Content.IsTopDialog() {
		return k.Action(evt)
	}

	return evt
}

func (ac *AppController) bindKeys() {
	ac.AddActions(ui.NewKeyActionsFromMap(ui.KeyMap{
		ui.KeyShift9: ui.NewSharedKeyAction("DumpGOR", ac.dumpGOR, false),
		// tcell.KeyCtrlE: ui.NewSharedKeyAction("ToggleHeader", a.toggleHeaderCmd, false),
		// tcell.KeyCtrlG: ui.NewSharedKeyAction("toggleCrumbs", a.toggleCrumbsCmd, false),
		// ui.KeyHelp:     ui.NewSharedKeyAction("Help", a.helpCmd, false),
		// tcell.KeyCtrlA: ui.NewSharedKeyAction("Aliases", a.aliasCmd, false),
		// tcell.KeyEnter: ui.NewKeyAction("Goto", a.gotoCmd, false),
	}))
}

func (a *App) dumpGOR(evt *tcell.EventKey) *tcell.EventKey {
	log.Printf("GOR %d\n", runtime.NumGoroutine())
	// bb := make([]byte, 5_000_000)
	// runtime.Stack(bb, true)
	// log.Debug().Msgf("GOR\n%s", string(bb))
	return evt
}

func (ac *AppController) layout() {
	main := tview.NewFlex().SetDirection(tview.FlexRow)
	// 	main.AddItem(ac.Content, 0, 10, true)
	ac.Main.AddPage("main", main, true, false)
}

// Run starts the application loop.
func (ac *AppController) Run() error {
	// ac.Resume()

	go func() {
		<-time.After(splashDelay)
		ac.QueueUpdateDraw(func() {
			ac.Main.SwitchToPage("main")
			// if command bar is already active, focus it
			// if ac.CmdBuff().IsActive() {
			// 	ac.SetFocus(ac.Prompt())
			// }
		})
	}()

	if err := ac.command.defaultCmd(); err != nil {
		return err
	}
	ac.SetRunning(true)
	if err := ac.Application.Run(); err != nil {
		return err
	}

	return nil
}
