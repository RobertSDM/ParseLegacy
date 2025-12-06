package parseLegacy

import (
	"strings"
	"time"

	winkb "parseLegacy/windowsKeyboard"

	"github.com/atotto/clipboard"
)

var (
	ErrInitApp      string = "erro ao iniciar o app"
	ErrTableShape          = "as tabela precisam ter a mesma largura"
	ErrTableHeaders        = "as tabela precisam ter as mesma headers"
	ErrNotFound            = "%s não foi encontrado(a)"
	ErrDirectory           = "você precisa selecionar uma pasta"
	ErrSave                = "erro ao salvar"
)

const ColumnLineNumber = 5

var ColumnsToDrop = []string{"Usuario", "Data"}

var Headers = []string{
	"Loja",
	"Fabricada",
	"Produto",
	"Qtd. Pedida",
	"Qtd. Recebida",
	"Qtd. Corte",
	"Data",
	"Hora",
	"Usuario",
}

var HeadersAlignment = []string{
	"left",
	"right",
	"right",
	"right",
	"right",
	"right",
	"left",
	"left",
	"left",
}

// Verify if a page is the last page
func IsLastPage(pageLines []string) bool {
	return strings.Contains(pageLines[len(pageLines)-3], "ULTIMA TELA")
}

// Return the legacy screen as text
func GetPage() string {
	winkb.KeyHold(winkb.VK_CONTROL, func() {
		winkb.KeyPress(winkb.VK_A)
	})
	winkb.KeyHold(winkb.VK_CONTROL, func() {
		winkb.KeyPress(winkb.VK_C)
	})

	time.Sleep(20 * time.Millisecond)
	text, _ := clipboard.ReadAll()

	return text
}

// Get the table contained in a slice of lines
func GetTable(lines []string) []string {
	tableStart := -1
	tableEnd := -1

	for i, line := range lines {
		if strings.Contains(line, "------") {
			tableStart = i + 4
			break
		}
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
func ColumnsPosition(colLine string) map[string]int {
	positions := map[string]int{}
	coli := 0

	for hi, h := range Headers {

		for coli < len(colLine) && colLine[coli] == ' ' {
			coli++
		}

		if HeadersAlignment[hi] == "left" {
			positions[h] = coli
		}

		for coli < len(colLine) && colLine[coli] != ' ' {
			coli++
		}

		if HeadersAlignment[hi] == "right" {
			positions[h] = coli - 1
		}
	}

	return positions
}

// Parse the columns values
func ParseTable(strTable []string, columnPositions map[string]int) *Table {
	table := NewTable(Headers)

	for _, line := range strTable {
		row := NewRow()
		coli := 0

		for hi, header := range table.Headers {
			var tmp strings.Builder

			for coli < len(line) && line[coli] == ' ' {
				if HeadersAlignment[hi] == "right" && coli >= columnPositions[header] {
					row.SetValue(header, "-")
					break
				}

				coli++
			}

			for coli < len(line) && line[coli] != ' ' {
				tmp.WriteByte(line[coli])
				coli++
			}

			if tmp.Len() > 0 {
				row.SetValue(header, tmp.String())
				continue
			}

		}

		table.AddRow(row)
	}

	return table
}
