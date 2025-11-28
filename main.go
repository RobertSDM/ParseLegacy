package main

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/go-vgo/robotgo"
	"github.com/sqweek/dialog"
	"github.com/xuri/excelize/v2"
)

const columnLinePosition = 5

var tableColumnsAndTypes = [][]string{
	{"Loja", "str"},
	{"Fabr", "str"},
	{"Prod", "int"},
	{"Qtd. Pedida", "int"},
	{"Qtd. Receb.", "int"},
	{"Qtd. Corte", "int"},
	{"Data", "str"},
	{"Hora", "str"},
	{"Usuário", "str"},
}

func main() {
	outDir, err := dialog.Directory().Title("Local para salvar o relatório").Browse()
	if err != nil {
		return
	}

	strDate := time.Now().Format("02-01-2006")

	time.Sleep(2 * time.Second)

	var table map[string][]string

	for i := 0; i < 100; i++ {
		page := getPage()

		lines := strings.Split(page, "\n")
		// Making sure the logic will work by cleaning and padding the lines
		for j := range lines {
			lines[j] = " " + strings.ReplaceAll(lines[j], "\r", "") + " "
		}

		tableStr := getTable(lines)

		columnLine := lines[columnLinePosition]
		positions := columnsPosition(columnLine)

		tb := parseCols(tableStr, positions)

		if table == nil {
			table = tb
		} else {
			table = appendTables(table, tb)
		}
		// Go to the next page
		robotgo.KeyTap("f8")
		time.Sleep(200 * time.Millisecond)
	}

	delete(table, "Usuário")

	outFile := filepath.Join(outDir, fmt.Sprintf("relatorio_%s.xlsx", strDate))
	if err := saveExcel(table, outFile); err != nil {
		return
	}

	// fmt.Printf("Relatório salvo em: %s\n", outFile)
}

// Append the second table to the first and return the first table
func appendTables(table1 map[string][]string, table2 map[string][]string) map[string][]string {
	headers := []string{}

	for _, tct := range tableColumnsAndTypes {
		headers = append(headers, tct[0])
	}

	rowi := 0
	for rowi < len(slices.Collect(maps.Values(table2))[0]) {
		for _, h := range headers {
			table1[h] = append(table1[h], table2[h][rowi])
		}
		rowi++
	}

	return table1
}

// Return the legacy screen as text
func getPage() string {
	robotgo.KeyTap("a", "ctrl")
	robotgo.KeyTap("c", "ctrl")

	time.Sleep(100 * time.Millisecond)
	text, _ := clipboard.ReadAll()
	return text
}

// Get the table contained in a slice of lines
func getTable(lines []string) []string {
	tableStart := -1
	tableEnd := -1

	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "---") {
			tableStart = i + 4
			break
		}
	}

	if tableStart == -1 {
		return nil
	}

	for i := tableStart; i < len(lines); i++ {
		if strings.TrimSpace(lines[i][:20]) == "" {
			tableEnd = i
			break
		}
	}

	return lines[tableStart:tableEnd]
}

// Get the position that best fits the columns title alignment
func columnsPosition(colLine string) map[string]int {
	positions := map[string]int{}
	col := false
	coli := 0

	for i, ch := range colLine {
		if ch == ' ' {
			if col {
				if tableColumnsAndTypes[coli][1] == "int" {
					positions[tableColumnsAndTypes[coli][0]] = i
				}
				coli++
			}
			col = false
			continue
		}

		if !col {
			if tableColumnsAndTypes[coli][1] == "str" {
				positions[tableColumnsAndTypes[coli][0]] = i
			}
		}
		col = true
	}

	return positions
}

// Parse the columns values
func parseCols(table []string, columnPositions map[string]int) map[string][]string {
	cols := map[string][]string{}
	for _, c := range tableColumnsAndTypes {
		cols[c[0]] = []string{}
	}

	tmp := ""
	for _, row := range table {
		gi := 0
		for i := 0; i < len(row); i++ {
			cl := row[i]
			if cl == ' ' {
				if tmp != "" {
					name := tableColumnsAndTypes[gi][0]
					cols[name] = append(cols[name], tmp)
					tmp = ""
					gi++
				} else {
					if gi < len(tableColumnsAndTypes) {
						typ := tableColumnsAndTypes[gi][1]
						name := tableColumnsAndTypes[gi][0]
						if typ == "int" {
							if pos, ok := columnPositions[name]; ok {
								if i >= pos {
									cols[name] = append(cols[name], "-")
									gi++
								}
							}
						}
					}
				}
				continue
			}
			tmp += string(cl)
		}
	}

	return cols
}

// Save a table as a .xlsx
func saveExcel(table map[string][]string, outFile string) error {
	f := excelize.NewFile()
	sheet := "Relatório"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{}

	for _, tct := range tableColumnsAndTypes {
		if tct[0] == "Usuário" {
			continue
		}

		headers = append(headers, tct[0])
	}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	rowi := 0
	for rowi < len(slices.Collect(maps.Values(table))[0]) {
		for j, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowi+2)
			f.SetCellValue(sheet, cell, table[h][rowi])
		}
		rowi++
	}

	if err := f.SaveAs(outFile); err != nil {
		return err
	}
	return nil
}
