package windowskeyboard

import (
	"testing"
	"time"
)

func TestListener(t *testing.T) {
	presschan := make(chan int8)

	ListenKeys([]VK_CODE{VK_ESCAPE}, func(k string) {
		presschan <- 1
	})

	time.Sleep(10 * time.Millisecond)
	KeyPress(VK_ESCAPE)

	select {
	case <-presschan:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Did not detect the key press")
	}
}
