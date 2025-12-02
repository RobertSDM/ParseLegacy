package main

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"parseLegacy/utils"
	winkb "parseLegacy/windowsKeyboard"

	"github.com/atotto/clipboard"
	"github.com/sqweek/dialog"
	"github.com/xuri/excelize/v2"
)

var columnsToDrop = []string{"Usuario", "Data"}

const columnLineNumber = 5

var tableColumnsAndTypes = [][]string{
	{"Loja", "str"},
	{"Fabr", "str"},
	{"Prod", "int"},
	{"Qtd. Pedida", "int"},
	{"Qtd. Receb.", "int"},
	{"Qtd. Corte", "int"},
	{"Data", "str"},
	{"Hora", "str"},
	{"Usuario", "str"},
}

var (
	ErrInitApp   error = errors.New("erro ao iniciar o app")
	ErrDirectory       = errors.New("você precisa selecionar uma pasta")
	ErrSave            = errors.New("erro ao salvar")
)

func main() {
	isRunning := true
	var table map[string][]string
	strDate := time.Now().Format("02-01-2006")

	if err := winkb.ListenKeys([]string{"VK_ESCAPE"}, func(k string) {
		isRunning = false
	}); err != nil {
		dialog.Message("%s", ErrInitApp).Title("Erro :(").Error()
		panic(err)
	}

	outDir, err := dialog.Directory().Title("Local para salvar o relatório").Browse()
	if err != nil {
		dialog.Message("%s", ErrDirectory).Title("Erro :(").Error()
		panic(err)
	}

	time.Sleep(2 * time.Second)

	for isRunning {
		page := getPage()
		lines := strings.Split(page, "\n")

		// Making sure the logic will work by cleaning and padding the lines
		for j := range lines {
			lines[j] = " " + strings.ReplaceAll(lines[j], "\r", "") + " "
		}

		tableRows := getTable(lines)

		columnTitleLine := lines[columnLineNumber]
		columnTitlePositions := columnsPosition(columnTitleLine)

		tb := parseTable(tableRows, columnTitlePositions)

		if table == nil {
			table = tb
		} else {
			table = appendTables(table, tb)
		}

		if isLastPage(lines) {
			isRunning = false
			continue
		}

		// GO to the next page
		winkb.KeyPress(winkb.VK_F8)
		time.Sleep(200 * time.Millisecond)
	}

	for _, col := range columnsToDrop {
		delete(table, col)
	}

	outFile := filepath.Join(outDir, fmt.Sprintf("relatorio_%s.xlsx", strDate))

	if err := saveExcel(table, outFile); err != nil {
		return
	}

	exec.Command("explorer", outDir).Run()

	// if err != nil && errors.Is(err, &exec.ExitError{}) {
	// 	dialog.Message("O processo foi finalizado, o relatório está neste caminho: %s", outDir).Title("Processo finalizado").Info()
	// } else if err != nil {
	// 	dialog.Message("%s o relatório", ErrSave).Title("Erro :(").Error()
	// 	panic(err)
	// }
}

// Verify if a page is the last page
func isLastPage(pageLines []string) bool {
	return strings.Contains(pageLines[len(pageLines)-3], "ULTIMA PAGINA") || strings.Contains(pageLines[len(pageLines)-3], "ULTIMA PÁGINA")
}

// Append the second table to the first and return the first table
func appendTables(table1 map[string][]string, table2 map[string][]string) map[string][]string {
	headers := []string{}

	for _, tct := range tableColumnsAndTypes {
		headers = append(headers, tct[0])
	}

	rowi := 0
	for rowi < len(utils.MapValues(table2)[0]) {
		for _, h := range headers {
			table1[h] = append(table1[h], table2[h][rowi])
		}
		rowi++
	}

	return table1
}

// Return the legacy screen as text
func getPage() string {
	winkb.KeyHold(winkb.VK_CONTROL, func() {
		winkb.KeyPress(winkb.VK_A)
	})
	winkb.KeyHold(winkb.VK_CONTROL, func() {
		winkb.KeyPress(winkb.VK_C)
	})

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

	for _, l := range lines[tableStart:]{
		fmt.Println(l)
	}
	
	for i, line := range lines[tableStart:] {
		if strings.Contains(line, "TOTAL=>") || strings.Count(line, " ") == len(line) {
			tableEnd = tableStart + i
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
func parseTable(table []string, columnPositions map[string]int) map[string][]string {
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
		if utils.Contains(columnsToDrop, tct[0]) {
			continue
		}

		headers = append(headers, tct[0])
	}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	rowi := 0
	for rowi < len(utils.MapValues(table)[0]) {
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
