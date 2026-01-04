package windowskeyboard

import (
	"os"
	"os/signal"
	"time"

	"parseLegacy/utils"

	"github.com/moutend/go-hook/pkg/keyboard"
	"github.com/moutend/go-hook/pkg/types"
)

func ListenKeys(keys []types.VKCode, cb func(k string)) (err error) {
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
				if k.Message.String() == "WM_KEYDOWN" && utils.SliceContains(keys, k.VKCode) {
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
