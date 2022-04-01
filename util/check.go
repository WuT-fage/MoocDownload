package util

import (
	"os/exec"

	"github.com/gookit/color"
	"github.com/spf13/viper"
	"github.com/wangluozhe/requests"
)

func CheckVersion() {
	r, _ := requests.Get("http://1.15.249.230:6000/checkVersion", nil)
	vision := r.Text
	if vision != viper.GetString("info.version") {
		color.Red.Println("\n版本已更新！请前往公众号获取最新版链接。")
	}

}

func CheckFfmpeg() bool {
	_, lookErr := exec.LookPath("ffmpeg")
	if lookErr != nil {
		return false
	}
	return true
}
