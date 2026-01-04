package windowskeyboard

import (
	"testing"
	"time"

	"github.com/moutend/go-hook/pkg/types"
)

func TestListener(t *testing.T) {
	presschan := make(chan int8)

	ListenKeys([]types.VKCode{types.VK_ESCAPE}, func(k string) {
		presschan <- 1
	})

	KeyPress(types.VK_ESCAPE)

	select {
	case <-presschan:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Did not detect the key press")
	}
}
