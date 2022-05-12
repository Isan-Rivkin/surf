package printer

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type Table interface {
	PrintInfoBox(infoBox map[string]string, labelsOrder []string)
}

type TableWidget struct {
}

func NewTablePrinter() Table {
	return &TableWidget{}
}

func (tp *TableWidget) rowsFromInfoBox(infoBox map[string]string, labelsOrder []string) []table.Row {
	rows := []table.Row{}
	temp := map[string]table.Row{}

	for col, val := range infoBox {
		temp[col] = table.Row{col, val}
	}

	for _, l := range labelsOrder {
		row, found := temp[l]
		if !found {
			continue
		}
		rows = append(rows, row)
	}

	return rows
}

func (tp *TableWidget) PrintInfoBox(infoBox map[string]string, labelsOrder []string) {
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}

	t := table.NewWriter()

	rows := tp.rowsFromInfoBox(infoBox, labelsOrder)

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
