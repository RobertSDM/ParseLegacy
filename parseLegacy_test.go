package parseLegacy

import (
	"path/filepath"
	"strings"
	"testing"

	"parseLegacy/testdata"

	winkb "parseLegacy/windowsKeyboard"

	"github.com/sqweek/dialog"
)

// UNIT Tests

func TestIsLastPage(t *testing.T) {
	lastPage := testdata.Pages[len(testdata.Pages)-1]
	lines := strings.Split(lastPage, "\n")

	if !IsLastPage(lines) {
		t.Fatal("Not detecting the last page")
	}

	somePage := testdata.Pages[0]
	lines = strings.Split(somePage, "\n")

	if IsLastPage(lines) {
		t.Fatal("False positive detection")
	}
}

func TestGetTableRange(t *testing.T) {
	notFullTable := testdata.Pages[len(testdata.Pages)-1]
	lines := strings.Split(notFullTable, "\n")

	tableRange := GetTableRange(lines)

	if len(tableRange) >= len(lines) {
		t.Fatal("Table range was not found")
	}

	for _, line := range tableRange {
		if strings.Count(line, " ") == len(line) {
			t.Fatal("Adding empty lines to the table range")
		}
	}
}

func TestColumnsPositions(t *testing.T) {
	// These column titles are 9 because this is the size of the constant Headers array
	smallColumnGap := " abc dfg hij klm nop qrs tuv wxy z"
	bigColumnGap := " abc            dfg hij     klm nop            qrs tuv     wxy       z"

	smallColumnGapLookupPositions := map[string]int{
		"Loja":          1,
		"Fabricada":     7,
		"Produto":       11,
		"Qtd. Pedida":   15,
		"Qtd. Recebida": 19,
		"Qtd. Corte":    23,
		"Data":          25,
		"Hora":          29,
		"Usuario":       33,
	}

	bigColumnGapLookupPositions := map[string]int{
		"Loja":          1,
		"Fabricada":     18,
		"Produto":       22,
		"Qtd. Pedida":   30,
		"Qtd. Recebida": 34,
		"Qtd. Corte":    49,
		"Data":          51,
		"Hora":          59,
		"Usuario":       69,
	}

	for k, v := range HeadersPositions(smallColumnGap) {
		if smallColumnGapLookupPositions[k] != v {
			t.Fatalf("Expected the small gap \"%s\" position to be %d but received %d", k, smallColumnGapLookupPositions[k], v)
		}
	}

	for k, v := range HeadersPositions(bigColumnGap) {
		if bigColumnGapLookupPositions[k] != v {
			t.Fatalf("Expected the big gap \"%s\" position to be %d but received %d", k, bigColumnGapLookupPositions[k], v)
		}
	}
}

func TestParseTable(t *testing.T) {
	page := testdata.Pages[len(testdata.Pages)-1]
	lines := strings.Split(page, "\n")
	tableRange := GetTableRange(lines)

	headerLine := lines[HeaderLineIndex]

	columnsPositions := HeadersPositions(headerLine)

	table := ParseTable(tableRange, columnsPositions)
	if table.Height != len(tableRange) {
		t.Fatal("Did not parse every table line")
	}

	if table.Width != len(Headers) {
		t.Fatal("Did not parse all columns")
	}
}

// E2E Test

func TestE2E(t *testing.T) {
	var tmpTable *Table

	table := NewTable(Headers)
	lineCount := 0

	appState := RUNNING
	stopped := false
	pausedRounds := 0

	err := winkb.ListenKeys([]winkb.VK_CODE{winkb.VK_ESCAPE, winkb.VK_F12}, func(k string) {
		switch k {
		case "VK_ESCAPE":
			stopped = true
		case "VK_F12":
			switch appState {
			case RUNNING:
				appState = PAUSED
			case PAUSED:
				appState = RUNNING
			}
		}
	})
	if err != nil {
		dialog.Message("%s", ErrInitApp).Title("Erro :(").Error()
		panic(err)
	}

	i := 0

	for appState != TERMINATED {
		if appState == PAUSED {
			pausedRounds++
			if pausedRounds == 3 {
				winkb.KeyPress(winkb.VK_F12)
			}
			continue
		}

		page := testdata.Pages[i]
		lines := strings.Split(page, "\n")

		if IsLastPage(lines) {
			appState = TERMINATED
			continue
		}

		tableRange := GetTableRange(lines)
		lineCount += len(tableRange)

		if i%3 == 0 {
			winkb.KeyPress(winkb.VK_ESCAPE)
		}

		if pausedRounds == 0 {
			winkb.KeyPress(winkb.VK_F12)
		}

		headerLine := lines[HeaderLineIndex]

		columnsPositions := HeadersPositions(headerLine)

		tmpTable = ParseTable(tableRange, columnsPositions)
		table.ConcatTable(tmpTable)

		i++
	}
	if table.Height != lineCount {
		t.Fatal("Not reading all lines")
	}

	if !stopped {
		t.Fatal("The stop event was not detected")
	}

	if pausedRounds != 3 {
		t.Fatal("Pause not working")
	}

	savePath := filepath.Join(t.TempDir(), "testTableSave.xlsx")
	table.ToExcel(savePath, "testTableSave")
}
