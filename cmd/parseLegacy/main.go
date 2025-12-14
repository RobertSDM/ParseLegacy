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

var appState parseLegacy.STATE

func main() {
	table := parseLegacy.NewTable(parseLegacy.Headers)
	appState = parseLegacy.RUNNING

	strNowDate := time.Now().Format("02-01-2006")

	err := winkb.ListenKeys([]string{"VK_ESCAPE", "VK_F12"}, func(k string) {
		switch k {
		case "VK_ESCAPE":
			appState = parseLegacy.ENDED
		case "VK_F12":
			switch appState {
			case parseLegacy.RUNNING:
				appState = parseLegacy.PAUSED
			case parseLegacy.PAUSED:
				appState = parseLegacy.RUNNING
			}

		}
	})
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

	for appState != parseLegacy.ENDED {
		if appState == parseLegacy.PAUSED {
			continue
		}

		page := parseLegacy.GetPage()
		lines := strings.Split(page, "\n")

		if parseLegacy.IsLastPage(lines) {
			appState = parseLegacy.ENDED
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
