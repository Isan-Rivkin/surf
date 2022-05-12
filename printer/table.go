package printer

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type Table interface {
	PrintInfoBox(infoBox map[string]string)
}

type TableWidget struct {
}

func NewTablePrinter() Table {
	return &TableWidget{}
}

func (tp *TableWidget) rowsFromInfoBox(infoBox map[string]string) []table.Row {
	rows := []table.Row{}

	for col, val := range infoBox {

		rows = append(rows, table.Row{col, val})

	}
	return rows
}

func (tp *TableWidget) PrintInfoBox(infoBox map[string]string) {
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}

	t := table.NewWriter()

	rows := tp.rowsFromInfoBox(infoBox)

	for _, r := range rows {
		t.AppendRow(r, rowConfigAutoMerge)
	}
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Colors: text.Colors{text.FgGreen}},
		{Number: 2},
	})

	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.SetOutputMirror(os.Stdout)
	t.Render()
}
