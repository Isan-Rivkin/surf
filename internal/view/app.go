package view

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

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
func generateRandomDigits(n int) string {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator
	digits := "0123456789"
	result := make([]byte, n)

	for i := range result {
		result[i] = digits[rand.Intn(len(digits))] // Random digit
	}

	return string(result)
}

// Generate a random word with a random length
func generateRandomWord(minLen, maxLen int) string {
	letters := "abcdefghijklmnopqrstuvwxyz"
	wordLength := rand.Intn(maxLen-minLen+1) + minLen
	word := make([]byte, wordLength)

	for i := range word {
		word[i] = letters[rand.Intn(len(letters))]
	}

	return string(word)
}

// Generate a random sentence of length between minLen and maxLen characters
func generateRandomSentence(minLen, maxLen int) string {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator
	var sentence strings.Builder

	// Choose a random sentence length between minLen and maxLen
	sentenceLength := rand.Intn(maxLen-minLen+1) + minLen

	// Keep generating words until we reach the desired length
	for sentence.Len() < sentenceLength {
		word := generateRandomWord(3, 8) // Random word between 3 and 8 characters

		// Ensure we don't exceed the maximum length, accounting for spaces
		if sentence.Len()+len(word)+1 > sentenceLength {
			break
		}

		if sentence.Len() > 0 {
			sentence.WriteString(" ") // Add a space between words
		}
		sentence.WriteString(word)
	}

	return sentence.String()
}
func mockSecurityGroupResourcesTable() *tview.Table {
	t := tview.NewTable()
	t.SetBorders(false)
	rowsSize := 200
	cols := []string{"#", "ID", "Name", "Description", "Ingress Rules", "Egress Rules", "VpcId"}
	rows := [][]string{
		{"1", "sg-12345678987654322", "my-security-group", "Clusters Test Security group created in Python", "4", "1", "vpc-12345678987654321"},
	}
	for i := 0; i < rowsSize; i++ {
		idx := fmt.Sprintf("%d", i+2)
		id := fmt.Sprintf("sg-%s", generateRandomDigits(17))
		vpc := fmt.Sprintf("vpc-%s", generateRandomDigits(17))
		name := generateRandomWord(5, 20)
		description := generateRandomSentence(0, 100)
		egressRules := fmt.Sprintf("%d", rand.Intn(10))
		ingressRules := fmt.Sprintf("%d", rand.Intn(10))
		rows = append(rows, []string{idx, id, name, description, ingressRules, egressRules, vpc})
	}
	// update columns
	for i := 0; i < len(cols); i++ {
		t.SetCell(0, i, tview.NewTableCell(cols[i]).SetTextColor(tcell.ColorOrange).SetAlign(tview.AlignLeft))
	}
	// update rows
	for i := 0; i < len(rows); i++ {
		for j := 0; j < len(rows[i]); j++ {
			t.SetCell(i+1, j, tview.NewTableCell(rows[i][j]).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
		}
	}
	t.SetFixed(1, 1)
	t.SetSelectable(true, false)
	t.Select(1, 0)
	t.SetSelectedFunc(func(row int, column int) {
		t.GetCell(row, column).SetTextColor(tcell.ColorRed)
	})
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
	sgTable := mockSecurityGroupResourcesTable()
	resources := sgTable
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
	a.app.SetFocus(resources)
	return nil
}
func (a *App) Run() error {
	return a.app.Run()
}
