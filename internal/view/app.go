package view

import (
	"github.com/common-nighthawk/go-figure"
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
	app     *tview.Application
	name    string
	version string
}

func NewApp(name, version string) *App {
	return &App{
		app:     tview.NewApplication(),
		name:    name,
		version: version,
	}
}
func mockSecurityGroupResourcesTable() *tview.Table {
	cols := []string{"ID", "Name", "Version", "Region", "Profile"}
	t := tview.NewTable()
	t.SetBorders(false)

	return t
}
func (a *App) Init() error {
	// taking inspiration from https://github.com/rivo/tview/wiki/Postgres
	// Flexbox for layout and pages with stack

	// create context box
	contextTable := tview.NewTable().SetBorders(false)
	// mock default profile
	contextTable.SetCell(0, 0, tview.NewTableCell("Profile:").SetTextColor(tcell.ColorOrange).SetAlign(tview.AlignLeft))
	contextTable.SetCell(0, 1, tview.NewTableCell("default").SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
	// mock sts get-caller-identity
	contextTable.SetCell(1, 0, tview.NewTableCell("Principal:").SetTextColor(tcell.ColorOrange).SetAlign(tview.AlignLeft))
	contextTable.SetCell(1, 1, tview.NewTableCell("arn:aws:iam::123:role/Dev/MyRole").SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
	// mock region us-west-2
	contextTable.SetCell(2, 0, tview.NewTableCell("Region:").SetTextColor(tcell.ColorOrange).SetAlign(tview.AlignLeft))
	contextTable.SetCell(2, 1, tview.NewTableCell("us-west-2").SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
	// mock version
	contextTable.SetCell(2, 0, tview.NewTableCell("Surf Rev:").SetTextColor(tcell.ColorOrange).SetAlign(tview.AlignLeft))
	contextTable.SetCell(2, 1, tview.NewTableCell(a.version).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))

	// add logo
	fig := figure.NewFigure(a.name, "starwars", true)
	title := fig.String()
	title = tview.TranslateANSI(title)

	logo := tview.NewTextView().SetText(title).SetTextColor(tcell.ColorOrange)
	flexHeader := tview.NewFlex().
		SetDirection(tview.FlexRowCSS).
		AddItem(contextTable, 0, 1, false).
		AddItem(logo, 50, 1, false)

	// add prompt
	prompt := tview.NewTextView()
	prompt.SetWordWrap(true)
	prompt.SetWrap(true)
	prompt.SetDynamicColors(true)
	prompt.SetBorder(true)
	prompt.SetBorderPadding(0, 0, 1, 1)
	prompt.SetTextColor(tcell.ColorPaleTurquoise)
	prompt.SetText("> AWS::EC2::SecurityGroup ")

	// add resource description (box placeholder)
	resources := tview.NewBox().SetBorder(true).SetTitle(" Security Groups ")
	// resources pane should be in pages and stack
	resourcesPages := tview.NewPages()
	resourcesPages.AddPage("resources", resources, true, true)

	// set layout using flexbox
	mainPage := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(flexHeader, 0, 1, false).
		AddItem(prompt, 3, 1, false).
		AddItem(resourcesPages, 0, 5, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("BreadCrumbs (3 rows)"), 3, 1, false)

	// create page
	pages := tview.NewPages()
	pages.AddPage("main", mainPage, true, true)
	a.app.SetRoot(pages, true)
	return nil
}
func (a *App) Run() error {
	return a.app.Run()
}
