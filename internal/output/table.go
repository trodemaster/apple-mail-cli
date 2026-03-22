package output

import (
	"fmt"
	"strings"
)

type Table struct {
	headers []string
	rows    [][]string
}

func NewTable(headers ...string) *Table {
	return &Table{headers: headers}
}

func (t *Table) AddRow(cols ...string) {
	t.rows = append(t.rows, cols)
}

func (t *Table) Print() {
	widths := make([]int, len(t.headers))
	for i, h := range t.headers {
		widths[i] = len(h)
	}
	for _, row := range t.rows {
		for i, col := range row {
			if i < len(widths) && len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	printRow := func(cols []string) {
		parts := make([]string, len(widths))
		for i, w := range widths {
			val := ""
			if i < len(cols) {
				val = cols[i]
			}
			parts[i] = fmt.Sprintf("%-*s", w, val)
		}
		fmt.Println(strings.Join(parts, "  "))
	}

	printRow(t.headers)
	sep := make([]string, len(widths))
	for i, w := range widths {
		sep[i] = strings.Repeat("-", w)
	}
	printRow(sep)
	for _, row := range t.rows {
		printRow(row)
	}
}
