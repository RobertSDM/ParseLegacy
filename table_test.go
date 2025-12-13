package parseLegacy

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

/// UNIT Tests

// Test if [parseLegacy.Table] insert rows correctly
func TestTableInsertRows(t *testing.T) {
	table := NewTable([]string{"product", "quantity"})

	row := NewRow()
	row.SetValue("product", "Soap")
	row.SetValue("quantity", "3")
	table.AddRow(row)

	row = NewRow()
	row.SetValue("product", "shampoo")
	row.SetValue("quantity", "1")
	table.AddRow(row)

	if table.Height != 2 {
		t.Fatalf("Error inserting rows. Expected %d and received %d", 2, table.Height)
	}
}

// Test if [parseLegacy.Table] concatenate another table
func TestTableConcatAnotherTable(t *testing.T) {
	table1 := NewTable([]string{"product", "quantity"})
	table2 := NewTable([]string{"product", "quantity"})

	row := NewRow()
	row.SetValue("product", "Milk")
	row.SetValue("quantity", "2")

	table1.AddRow(row)
	ctrlHeightTable1 := table1.Height

	row = NewRow()
	row.SetValue("product", "Chocolate Powder")
	row.SetValue("quantity", "3")

	table2.AddRow(row)

	row = NewRow()
	row.SetValue("product", "Egg")
	row.SetValue("quantity", "4")

	table2.AddRow(row)

	table1.ConcatTable(table2)

	if table1.Height != (ctrlHeightTable1 + table2.Height) {
		t.Fatal("Tables are not being concatenated corretly")
	}
}

// Test if [parseLegacy.Table] not concatenate the same table
func TestTableConcatSameTable(t *testing.T) {
	execComplete := make(chan int)

	table1 := NewTable([]string{"product", "quantity"})

	row := NewRow()
	row.SetValue("product", "Milk")
	row.SetValue("quantity", "2")

	table1.AddRow(row)

	go func() {
		table1.ConcatTable(table1)
		execComplete <- 1
	}()

	select {
	case <-execComplete:
	case <-time.After(500 * time.Millisecond):
		t.Fatal(ErrSameTableConcat)
	}
}

// Test if [parseLegacy.Table] not insert empty row
func TestTableNotInsertEmptyRow(t *testing.T) {
	table := NewTable([]string{})
	row := NewRow()

	if err := table.AddRow(row); err == nil {
		t.Fatal(ErrEmptyRow)
	}
}

// Test if [parseLegacy.Table] not append empty tables
func TestTableNotConcatEmptyTable(t *testing.T) {
	table := NewTable([]string{})
	tableToInsert := NewTable([]string{})

	if err := table.ConcatTable(tableToInsert); err == nil {
		t.Fatal(ErrEmptyTable)
	}
}

// Test if [parseLegacy.Table] drop columns
func TestTableDropColumns(t *testing.T) {
	table := NewTable([]string{"product", "quantity"})

	row := NewRow()
	row.SetValue("product", "Milk")
	row.SetValue("quantity", "2")

	table.AddRow(row)

	row = NewRow()
	row.SetValue("product", "Chocolate Powder")
	row.SetValue("quantity", "3")

	table.AddRow(row)

	row = NewRow()
	row.SetValue("product", "Egg")
	row.SetValue("quantity", "4")

	table.AddRow(row)

	ctrlWidth := table.Width

	table.Drop([]string{"product"})

	if table.Width >= ctrlWidth {
		t.Fatal("Error dropping columns from table")
	}

	if len(table.Headers) >= ctrlWidth {
		t.Fatal("Error dropping columns from headers")
	}

	ctrlWidth = table.Width

	for _, row := range table.Rows {
		if len(row) != ctrlWidth {
			t.Fatal("Error dropping columns from rows")
		}
	}
}

// Test if [parseLegacy.Table] save to a .xlsx file
func TestTableExcelSave(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "toExcelTestSave.xlsx")

	table := NewTable([]string{})
	err := table.ToExcel(savePath, "testSheet")
	if err != nil {
		t.Fatal("Failed saving table to .xlsx")
	}

	if _, err := os.Stat(savePath); errors.Is(err, os.ErrNotExist) {
		t.Fatal("Error getting the savePath stat")
	}
}
