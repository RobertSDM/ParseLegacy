package windowskeyboard

import (
	"os"
	"os/signal"
	"time"

	"parseLegacy/utils"

	"github.com/moutend/go-hook/pkg/keyboard"
	"github.com/moutend/go-hook/pkg/types"
)

func stringToVkCode(key string) VK_CODE {
	switch key {
	case "VK_A":
		return VK_A
	case "VK_C":
		return VK_C
	case "VK_CONTROL":
		return VK_CONTROL
	case "VK_F8":
		return VK_F8
	case "VK_ESCAPE":
		return VK_ESCAPE
	}
	return 0
}

func ListenKeys(keys []VK_CODE, cb func(k string)) (err error) {
	lowLevelKeychan := make(chan types.KeyboardEvent, 100)

	if err := keyboard.Install(nil, lowLevelKeychan); err != nil {
		return err
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		defer keyboard.Uninstall()
		defer close(lowLevelKeychan)

		for {
			select {
			case k := <-lowLevelKeychan:
				if k.Message.String() == "WM_KEYDOWN" && utils.SliceContains(keys, stringToVkCode(k.VKCode.String())) {
					cb(k.VKCode.String())
				}

			case <-signalChan:
				return
			}
		}
	}()
	time.Sleep(10 * time.Millisecond)

	return nil
}
