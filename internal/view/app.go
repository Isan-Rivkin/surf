package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

/*
------------------------
| context |		| logo |
------------------------
| 		search bar 	   |
------------------------
|       table          |
|                      |
------------------------
| bread crumbs |       |
------------------------
*/

type App struct {
	app *tview.Application
}

func NewApp() *App {
	return &App{
		app: tview.NewApplication(),
	}
}

func (a *App) Init() error {
	// create context box
	contextTable := tview.NewTable().SetBorders(false)
	contextTable.SetCell(0, 0, tview.NewTableCell("Profile").SetTextColor(tcell.ColorOrange).SetAlign(tview.AlignCenter))
	return nil
}
