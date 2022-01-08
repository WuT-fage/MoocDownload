package login

import (
	"github.com/go-resty/resty/v2"
	"github.com/gookit/color"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"moocVip/qrcodeTerminal"
	"os/exec"
	"time"
)

func WebLogin() string {
	cmd := exec.Command("cmd.exe", "/c", "loginWeb.exe")
	cmd.Start()
	client := resty.New()
	var codeStatus int

	client.SetHeaders(map[string]string{
		"referer":    "https://www.icourse163.org/",
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36",
	})

	// 获取二维码图片以及pollkey
	rep, _ := client.R().Get("http://127.0.0.1:3001/qrcode")
	codeUrl := rep.String()

	//下载图片
	rep, _ = client.R().Get(codeUrl)
	ioutil.WriteFile("QR.png", rep.Body(), 0666)
	qrcodeTerminal.DisplayImg("QR.png")

	token := ""
	//登录状态检查
	for codeStatus != 4 {
		rep, _ = client.R().
			Get("http://127.0.0.1:3001/pollKey")
		codeStatus = int(gjson.Get(rep.String(), "result.codeStatus").Int())
		switch codeStatus {
		case 0:
			color.Cyan.Println("请及时用mooc app扫码!")
		case 1:
			color.Green.Println("请点击登录!")
		case 2:
			color.Red.Println("登录成功!")
			codeStatus = 4
			token = gjson.Get(rep.String(), "result.token").String()
		case 3:
			color.Red.Println("二维码已失效，请重新运行程序")
			time.Sleep(4 * time.Second)
		}
		time.Sleep(2 * time.Second)
	}
	rep, _ = client.R().SetQueryParam("token", token).Get("http://127.0.0.1:3001/mocMobChangeCookie")
	cookieStr := rep.String()
	err := cmd.Process.Kill()
	cmd.Process.Wait()
	if err != nil {
		panic(err)
	}
	return cookieStr
}
