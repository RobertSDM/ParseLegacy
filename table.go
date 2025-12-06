package parseLegacy

import (
	"errors"
	"fmt"

	"parseLegacy/utils"

	"github.com/xuri/excelize/v2"
)

type Row map[string]string

// Create or update a column's value
func (r Row) SetValue(h string, v string) {
	r[h] = v
}

// Return the column value
func (r Row) GetValue(h string) string {
	return r[h]
}

// Delete the header
func (r Row) drop(h string) {
	delete(r, h)
}

func NewRow() Row {
	return make(Row)
}

// Struct with rows and columns
type Table struct {
	// The columns names.
	// The in this structure will be the order the columns will be printed and saved
	Headers []string
	Rows    []Row
	Width   int
	Height  int
}

// Return the width and height as a string
func (t *Table) Shape() string {
	return fmt.Sprintf("%dx%d", t.Width, t.Height)
}

// Save the table in excel format
func (t *Table) ToExcel(savePath string, sheetName string) error {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", sheetName)

	for i, h := range t.Headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	for i, row := range t.Rows {
		for j, h := range t.Headers {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
			f.SetCellValue(sheetName, cell, row[h])
		}
	}

	if err := f.SaveAs(savePath); err != nil {
		return err
	}

	return nil
}

// Append a row to the table's end
func (t *Table) AddRow(row Row) {
	t.Rows = append(t.Rows, row)
	t.Height++
}

// Append a new column to headers attribute and all rows
func (t *Table) AddColumn(header string) {
	t.Headers = append(t.Headers, header)

	for _, row := range t.Rows {
		row.SetValue(header, "")
	}
}

// Drop the headers from all rows and headers attribute
func (t *Table) Drop(headers []string) error {
	for _, h := range headers {
		if !utils.SliceContains(t.Headers, h) {
			return fmt.Errorf(ErrNotFound, h)
		}
	}

	t.Width -= len(headers)

	for _, h := range headers {
		for _, row := range t.Rows {
			row.drop(h)
		}
	}

	newHeaders := []string{}

	for _, h1 := range t.Headers {
		if utils.SliceContains(headers, h1) {
			continue
		}
		newHeaders = append(newHeaders, h1)
	}

	t.Headers = newHeaders

	return nil
}

// Append all the rows from the table
func (t *Table) ConcatTable(table *Table) error {
	if t.Width != table.Width {
		return errors.New(ErrTableShape)
	}

	for i, h := range t.Headers {
		if table.Headers[i] != h {
			return errors.New(ErrTableHeaders)
		}
	}

	for i := 0; i < table.Height; i++ {
		t.AddRow(table.Rows[i])
	}

	return nil
}

// Returns a new table with the provided headers
func NewTable(headers []string) *Table {
	return &Table{
		Headers: headers,
		Rows:    make([]Row, 0),
		Width:   len(headers),
		Height:  0,
	}
}
