package parseLegacy

import (
	"errors"
	"strings"
	"time"

	winkb "parseLegacy/windowsKeyboard"

	"github.com/atotto/clipboard"
	"github.com/moutend/go-hook/pkg/types"
)

// Shown to the user errors

var (
	ErrInitApp        = errors.New("erro ao iniciar o app")
	ErrFolderNotFound = errors.New("a pasta não foi foi encontrado(a)")
	ErrDirectory      = errors.New("você precisa selecionar uma pasta")
	ErrSave           = errors.New("erro ao salvar")
)

// Not shown to the user errors

var (
	ErrEmptyRow        = errors.New("cannot add empty rows")
	ErrTableShape      = errors.New("the tables need to have the same height")
	ErrTableHeaders    = errors.New("the tables need to have the same headers")
	ErrEmptyTable      = errors.New("cannot concat a empty table")
	ErrSameTableConcat = errors.New("cannot concat the same table")
)

const HeaderLineIndex = 5

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

type STATE int

const (
	RUNNING STATE = iota
	PAUSED
	TERMINATED
)

// Verify if a page is the last page
func IsLastPage(pageLines []string) bool {
	for i := len(pageLines) - 1; i >= 0; i-- {
		if strings.Contains(pageLines[i], "ULTIMA TELA") {
			return true
		}
	}
	return false
}

// Sequence to select the text from the terminal
func copyTeminal() {
	winkb.KeyHold(types.VK_CONTROL, func() {
		winkb.KeyPress(types.VK_HOME)
	})
	winkb.KeyHold(types.VK_SHIFT, func() {
		winkb.KeyPress(types.VK_UP)
		winkb.KeyPress(types.VK_LEFT)
	})
}

// Return the legacy screen as text
func GetPage() string {
	copyTeminal()

	winkb.KeyHold(types.VK_CONTROL, func() {
		winkb.KeyPress(types.VK_C)
	})

	time.Sleep(20 * time.Millisecond)
	text, _ := clipboard.ReadAll()

	return text
}

// Get the table contained in a slice of lines
func GetTableRange(lines []string) []string {
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
func HeadersPositions(colLine string) map[string]int {
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
