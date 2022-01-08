package qrcodeTerminal

import (
	"bytes"
	"github.com/gocq/qrcode"
	"io/ioutil"
)

func DisplayImg(path string) {
	fii, _ := ioutil.ReadFile(path)
	fi, err := qrcode.Decode(bytes.NewReader(fii))
	if err != nil {
		panic(err)
	}
	New2(ConsoleColors.BrightBlack, ConsoleColors.BrightWhite, QRCodeRecoveryLevels.Low).Get(fi.Content).Print()

}
