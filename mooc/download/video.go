package download

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"

	"github.com/gookit/color"
	"github.com/panjf2000/ants/v2"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"

	"MoocDownload/crypt"
	"MoocDownload/mooc/js"
	"MoocDownload/mooc/utils"
)

func VipDecryptTs(chapterNamePath string, TsUrl string, key string, index int, wg *sync.WaitGroup) func() {
	return func() {
		defer wg.Done()
		headers := url.NewHeaders()
		headers.Set("origin", "https://www.icourse163.org")
		headers.Set("referer", "https://www.icourse163.org/")
		headers.Set("authority", "mooc2vod.stu.126.net")
		headers.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36")
		req := url.NewRequest()
		req.Headers = headers
		r, _ := requests.Get(TsUrl, req)
		encrypter := r.Content
		path := fmt.Sprintf("%s\\tem\\%d.ts", chapterNamePath, index)
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()
		iv := utils.Iv(index)
		Key, _ := hex.DecodeString(key)
		Byte := crypt.CBCDecrypter(encrypter, Key, iv)
		file.Write(Byte)
	}
}

// 公开课视频下载
func FreeTs(chapterNamePath string, TsUrl string, index int, wg *sync.WaitGroup) func() {
	return func() {
		defer wg.Done()
		headers := url.NewHeaders()
		headers.Set("origin", "https://www.icourse163.org")
		headers.Set("referer", "https://www.icourse163.org/")
		headers.Set("authority", "mooc2vod.stu.126.net")
		headers.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36")
		req := url.NewRequest()
		req.Headers = headers
		r, _ := requests.Get(TsUrl, req)
		VideoByte := r.Content
		path := fmt.Sprintf("%s\\tem\\%d.ts", chapterNamePath, index)
		target, _ := os.Create(path)
		target.Write(VideoByte)
		target.Close()
	}
}

func VipGetTsKey(encryptStr string, videoId int) ([]string, string) {
	videoId_ := strconv.Itoa(videoId)
	m3u8 := js.M3u8(encryptStr, videoId_)
	tsCmp := regexp.MustCompile("http.*?ts")
	// 获取ts列表
	tsList := tsCmp.FindAllString(m3u8, -1)
	// 获取key
	keyCmp := regexp.MustCompile(`URI="(.*?)"`)
	keyUrl := keyCmp.FindStringSubmatch(m3u8)[1]

	headers := url.NewHeaders()
	headers.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36")
	headers.Set("origin", "https://www.icourse163.org")
	headers.Set("referer", "https://www.icourse163.org/")
	headers.Set("authority", "mooc2vod.stu.126.net")
	res, _ := requests.Get(keyUrl, &url.Request{Headers: headers})
	text := res.Text
	key := js.Key(text, videoId_)
	return tsList, key
}

func FreeGetTs(M3u8Str string) []string {
	tsCmp := regexp.MustCompile("\\d+.*?ts")
	// 获取ts列表
	tsList := tsCmp.FindAllString(M3u8Str, -1)
	return tsList
}

// 付费视频
func VipVideo(TsList []string, key string, unitName string, chapterNamePath string) {
	temPath := fmt.Sprintf("%s\\tem", chapterNamePath)
	utils.PathExists(temPath)
	var wg sync.WaitGroup
	pool, _ := ants.NewPool(15)
	defer pool.Release()

	barP := mpb.New(mpb.WithWidth(60), mpb.WithWaitGroup(&wg))

	total := len(TsList)
	name := fmt.Sprintf("%s.mp4 :", unitName)
	// create a single bar, which will inherit container's width
	bar := barP.New(int64(total),
		// BarFillerBuilder with custom style
		mpb.BarStyle().Lbound("╢").Filler("▌").Tip("▌").Padding("░").Rbound("╟"),

		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(name, decor.WC{W: len(name), C: decor.DidentRight}),
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}), "done",
			),
		),
		mpb.AppendDecorators(decor.Percentage()),
	)
	wg.Add(total)
	for index, TsUrl := range TsList {
		_ = pool.Submit(VipDecryptTs(chapterNamePath, TsUrl, key, index, &wg))
		bar.Increment()
	}
	barP.Wait()
	wg.Wait()
	MergeTs(len(TsList), unitName, chapterNamePath)
	fmt.Printf("%s.mp4 done\n", unitName)
	err := os.RemoveAll(temPath)
	if err != nil {
		panic(err)
	}
}

// 公开课视频
func FreeVideo(TsList []string, unitName string, chapterNamePath string) {
	// temPath := fmt.Sprintf("%s\\tem", chapterNamePath)
	// utils.PathExists(temPath)

	var wg sync.WaitGroup
	pool, _ := ants.NewPool(15)
	defer pool.Release()

	barP := mpb.New(mpb.WithWidth(60), mpb.WithWaitGroup(&wg))

	total := len(TsList)
	name := fmt.Sprintf("%s.mp4 :", unitName)
	// create a single bar, which will inherit container's width
	bar := barP.New(int64(total),
		// BarFillerBuilder with custom style
		mpb.BarStyle().Lbound("╢").Filler("▌").Tip("▌").Padding("░").Rbound("╟"),
		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(name, decor.WC{W: len(name), C: decor.DidentRight}),
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}), "done",
			),
		),
		mpb.AppendDecorators(decor.Percentage()),
	)
	wg.Add(total)
	for index, TsUrl := range TsList {
		// wg.Add(1)
		_ = pool.Submit(FreeTs(chapterNamePath, TsUrl, index, &wg))
		bar.Increment()
	}
	barP.Wait()
	wg.Wait()
	MergeTs(len(TsList), unitName, chapterNamePath)

}

func MergeTs(count int, name string, chapterNamePath string) {
	var concatStr string
	for i := 0; i < count; i++ {
		if i != count-1 {
			concatStr += fmt.Sprintf("%s\\tem\\%d.ts|", chapterNamePath, i)
		} else {
			concatStr += fmt.Sprintf("%s\\tem\\%d.ts", chapterNamePath, i)
		}
	}
	ffmpeg, lookErr := exec.LookPath("ffmpeg")
	if lookErr != nil {
		color.Red.Println("\n您还没安装ffmpeg,无法下载，请您安装后下载!\n")
		return
	}
	args := []string{
		"-i",
		fmt.Sprintf("concat:%s", concatStr),
		"-acodec",
		"copy",
		"-vcodec",
		"copy",
		"-absf",
		"aac_adtstoasc",
		fmt.Sprintf("%s\\%s.mp4", chapterNamePath, name),
	}
	cmd := exec.Command(ffmpeg, args...)
	_, err := cmd.Output()
	if err != nil {
		color.Red.Println("！！！！！！！！！！！！警告，合成视频失败！！！！！！！！！！！\n")
		return
		// panic(err)
	}
	// fmt.Println(string(r))
	fmt.Println(name + ".mp4 done\n")
}

func Mp4Flv(videoUrl, unitName, chapterNamePath string, Isflv2mp4 bool) {
	fmt.Println("", unitName+".mp4", " start")
	var path string
	if Isflv2mp4 {
		path = fmt.Sprintf("%s//%s.flv", chapterNamePath, unitName)
	} else {
		path = fmt.Sprintf("%s//%s.mp4", chapterNamePath, unitName)
	}
	r, _ := http.Get(videoUrl)
	// 获取文件大小
	fileSize, _ := strconv.Atoi(r.Header["Content-Length"][0])
	// 创建文件
	target, _ := os.Create(path)
	p := mpb.New(mpb.WithWidth(60))
	bar := p.New(int64(fileSize),
		mpb.BarStyle().Rbound("|"),
		mpb.PrependDecorators(
			decor.CountersKibiByte("% .2f / % .2f"),
			decor.OnComplete(
				decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}), " done",
			),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_GO, 90),
			decor.Name(" ] "),
			decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
		),
	)
	reader := bar.ProxyReader(r.Body)
	defer reader.Close()
	// 将下载的文件流拷贝到临时文件
	if _, err := io.Copy(target, reader); err != nil {
		target.Close()
	}
	target.Close()
	p.Wait()
	if Isflv2mp4 {
		ffmpeg, lookErr := exec.LookPath("ffmpeg")
		if lookErr != nil {
			color.Red.Println("\n您还没安装ffmpeg,无法下载，请您安装后下载!\n")
			return
		}
		args := []string{
			"-i",
			path,
			"-vcodec",
			"copy",
			"-acodec",
			"copy",
			fmt.Sprintf("%s\\%s.mp4", chapterNamePath, unitName),
		}
		cmd := exec.Command(ffmpeg, args...)
		_, err := cmd.Output()
		if err != nil {
			color.Red.Println("！！！！！！！！！！！！警告，flv视频转mp4失败！！！！！！！！！！！")
			return
			// panic(err)
		}
		os.Remove(path)
	}
	fmt.Println("", unitName+".mp4", " done\n")
}
