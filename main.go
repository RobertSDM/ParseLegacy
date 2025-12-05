package main

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
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

var legacyTableAndAlignment = [][]string{
	{"Loja", "left"},
	{"Fabricante", "left"},
	{"Produto", "right"},
	{"Qtd. Pedida", "right"},
	{"Qtd. Recebimento", "right"},
	{"Qtd. Corte", "right"},
	{"Data", "left"},
	{"Hora", "left"},
	{"Usuario", "left"},
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

		if isLastPage(lines) {
			isRunning = false
			continue
		}

		// Making sure the logic will work by cleaning and padding the lines
		for j := range lines {
			lines[j] = " " + strings.ReplaceAll(lines[j], "\r", "") + " "
		}

		tableRows := getTable(lines)

		columnsNamesLine := lines[columnLineNumber]
		columsNamesPositions := columnsPosition(columnsNamesLine)

		tb := parseTable(tableRows, columsNamesPositions)

		if table == nil {
			table = tb
		} else {
			table = appendTables(table, tb)
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
	return strings.Contains(pageLines[len(pageLines)-3], "ULTIMA TELA")
}

// Append the second table to the first and return the first table
func appendTables(table1 map[string][]string, table2 map[string][]string) map[string][]string {
	headers := []string{}

	for _, tct := range legacyTableAndAlignment {
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
				if legacyTableAndAlignment[coli][1] == "right" {
					positions[legacyTableAndAlignment[coli][0]] = i
				}
				coli++
			}
			col = false
			continue
		}

		if !col {
			if legacyTableAndAlignment[coli][1] == "left" {
				positions[legacyTableAndAlignment[coli][0]] = i
			}
		}
		col = true
	}

	return positions
}

// Parse the columns values
func parseTable(table []string, columnPositions map[string]int) map[string][]string {
	cols := map[string][]string{}
	for _, c := range legacyTableAndAlignment {
		cols[c[0]] = []string{}
	}

	tmp := ""
	for _, row := range table {
		gi := 0
		for i := 0; i < len(row); i++ {
			cl := row[i]
			if cl == ' ' {
				if tmp != "" {
					name := legacyTableAndAlignment[gi][0]
					cols[name] = append(cols[name], tmp)
					tmp = ""
					gi++
				} else {
					if gi < len(legacyTableAndAlignment) {
						typ := legacyTableAndAlignment[gi][1]
						name := legacyTableAndAlignment[gi][0]
						if typ == "right" {
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

// Group all rows with the same value in the specified columnsToHash param.
// The group behaviour is to sum the values.
//
// The returned table will contain only the columns in the columnsToHash and columnsToGroup. The others will be ignored.
func groupby(table map[string][]string, columnsToHash []string, columnsToGroup []string) map[string][]string {
	groupedTable := make(map[string]map[string]string)

	tableSize := len(utils.MapValues(table)[0])

	for row := 0; row < tableSize; row++ {
		rowLine := make(map[string]string)

		for k, v := range table {

			if !utils.SliceContains(columnsToGroup, k) && !utils.SliceContains(columnsToHash, k) {
				continue
			}

			rowLine[k] = v[row]
		}

		hashValues := []string{}

		for _, c := range columnsToHash {
			hashValues = append(hashValues, rowLine[c])
		}

		hash := strings.Join(hashValues, "\\")

		rowLineGrouped, ok := groupedTable[hash]

		if !ok {
			groupedTable[hash] = rowLine
			continue
		}

		for _, c := range columnsToGroup {

			// Just skip when both are "-" to avoid changing the column value to 0
			if rowLine[c] == "-" && rowLineGrouped[c] == "-" {
				continue
			}

			rowLineV, _ := strconv.ParseInt(rowLine[c], 10, 64)
			gTV, _ := strconv.ParseInt(rowLineGrouped[c], 10, 64)

			rowLineGrouped[c] = fmt.Sprint(rowLineV + gTV)
		}
	}

	finalTable := make(map[string][]string, len(groupedTable))

	for _, t := range groupedTable {
		for k, v := range t {
			finalTable[k] = append(finalTable[k], v)
		}
	}

	return finalTable
}

// Save a table as a .xlsx
func saveExcel(table map[string][]string, savePath string) error {
	f := excelize.NewFile()
	mainSheet := "Relatório Principal"
	f.SetSheetName("Sheet1", mainSheet)

	headers := []string{}

	for k := range table {
		if utils.SliceContains(columnsToDrop, k) {
			continue
		}

		headers = append(headers, k)
	}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(mainSheet, cell, h)
	}

	rowi := 0
	for rowi < len(utils.MapValues(table)[0]) {
		for j, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowi+2)
			f.SetCellValue(mainSheet, cell, table[h][rowi])
		}
		rowi++
	}

	/// TODO: Rewrite saveExcel

	/// Grouping by Loja and Produto

	groupedTable := groupby(table, []string{"Loja", "Produto"}, []string{
		"Qtd. Pedida",
		"Qtd. Recebimento",
		"Qtd. Corte",
	})

	groupbySheet := "Por Loja e Produto"
	f.NewSheet(groupbySheet)

	headers = []string{}

	for k := range groupedTable {
		if utils.SliceContains(columnsToDrop, k) {
			continue
		}

		headers = append(headers, k)
	}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(groupbySheet, cell, h)
	}

	rowi = 0
	for rowi < len(utils.MapValues(groupedTable)[0]) {
		for j, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowi+2)
			f.SetCellValue(groupbySheet, cell, groupedTable[h][rowi])
		}
		rowi++
	}

	//

	if err := f.SaveAs(savePath); err != nil {
		return err
	}
	return nil
}
