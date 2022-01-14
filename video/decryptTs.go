package video

import (
	"encoding/hex"
	"fmt"
	"github.com/go-resty/resty/v2"
	"moocVip/util"
	"os"
	"sync"
)

//付费视频解密下载
func VipDecryptTs(chapterNamePath string, TsUrl string, key string, index int, wg *sync.WaitGroup) func() {
	return func() {
		//fmt.Println(index, "<---->", TsUrl)
		client := *resty.New()
		client.SetHeaders(map[string]string{
			"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36",
			"origin":     "https://www.icourse163.org",
			"referer":    "https://www.icourse163.org/",
			"authority":  "mooc2vod.stu.126.net",
		})
		res, _ := client.R().Get(TsUrl)
		encrypter := res.Body()
		path := fmt.Sprintf("%s\\tem\\%d.ts", chapterNamePath, index)
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()
		iv := util.Iv(index)
		Key, _ := hex.DecodeString(key)
		Byte := CBCDecrypter(encrypter, Key, iv)
		file.Write(Byte)
		wg.Done()
	}
}

//公开课视频下载
func FreeTs(chapterNamePath string, TsUrl string, index int, wg *sync.WaitGroup) func() {
	return func() {
		//fmt.Println(index, "<---->", TsUrl)
		client := *resty.New()
		client.SetHeaders(map[string]string{
			"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36",
			"origin":     "https://www.icourse163.org",
			"referer":    "https://www.icourse163.org/",
			"authority":  "mooc2vod.stu.126.net",
		})
		res, _ := client.R().Get(TsUrl)
		VideoByte := res.Body()
		path := fmt.Sprintf("%s\\tem\\%d.ts", chapterNamePath, index)
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()
		file.Write(VideoByte)
		wg.Done()
	}
}