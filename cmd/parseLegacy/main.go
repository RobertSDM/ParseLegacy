package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"parseLegacy"

	winkb "parseLegacy/windowsKeyboard"

	"github.com/sqweek/dialog"
)

func main() {
	table := parseLegacy.NewTable(parseLegacy.Headers)

	isRunning := true
	strNowDate := time.Now().Format("02-01-2006")

	err := winkb.ListenKeys([]string{"VK_ESCAPE"}, func(k string) { isRunning = false })
	if err != nil {
		dialog.Message("%s", parseLegacy.ErrInitApp).Title("Erro :(").Error()
		panic(err)
	}

	outDir, err := dialog.Directory().Title("Local para salvar o relatório").Browse()
	if err != nil {
		dialog.Message("%s", parseLegacy.ErrDirectory).Title("Erro :(").Error()
		panic(err)
	}

	// Wait time to select the screen
	time.Sleep(2 * time.Second)

	for isRunning {
		page := parseLegacy.GetPage()
		lines := strings.Split(page, "\n")

		if parseLegacy.IsLastPage(lines) {
			isRunning = false
			continue
		}

		// Making sure the logic will work by cleaning and padding the lines
		for j := range lines {
			lines[j] = " " + strings.ReplaceAll(lines[j], "\r", "") + " "
		}

		strTableRows := parseLegacy.GetTableRange(lines)

		columnsNamesLine := lines[parseLegacy.HeaderLineIndex]
		columsNamesPositions := parseLegacy.HeadersPositions(columnsNamesLine)

		tb := parseLegacy.ParseTable(strTableRows, columsNamesPositions)

		table.ConcatTable(tb)

		// GO to the next page
		winkb.KeyPress(winkb.VK_F8)
		time.Sleep(200 * time.Millisecond)
	}

	table.Drop(parseLegacy.ColumnsToDrop)

	outFile := filepath.Join(outDir, fmt.Sprintf("relatorio_%s.xlsx", strNowDate))

	table.ToExcel(outFile, "Relatório Principal")

	exec.Command("explorer", outDir).Run()
}
