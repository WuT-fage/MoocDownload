package video

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/panjf2000/ants/v2"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"moocVip/js"
	"moocVip/util"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

//付费视频
func VipVideo(TsList []string, key string, unitName string, chapterNamePath string) {
	temPath := fmt.Sprintf("%s\\tem", chapterNamePath)
	util.PathExists(temPath)
	var wg sync.WaitGroup
	pool, _ := ants.NewPool(5)
	defer pool.Release()

	barP := mpb.New(mpb.WithWidth(64))

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

	for index, TsUrl := range TsList {
		wg.Add(1)
		_ = pool.Submit(VipDecryptTs(chapterNamePath, TsUrl, key, index, &wg))
		bar.Increment()
	}
	barP.Wait()
	wg.Wait()
	MergeTs(len(TsList), unitName, chapterNamePath)
	err := os.RemoveAll(temPath)
	if err != nil {
		panic(err)
	}
}

//公开课视频
func FreeVideo(TsList []string, unitName string, chapterNamePath string) {
	temPath := fmt.Sprintf("%s\\tem", chapterNamePath)
	util.PathExists(temPath)
	var wg sync.WaitGroup
	pool, _ := ants.NewPool(5)
	defer pool.Release()

	barP := mpb.New(mpb.WithWidth(64))

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

	for index, TsUrl := range TsList {
		wg.Add(1)
		_ = pool.Submit(FreeTs(chapterNamePath, TsUrl, index, &wg))
		bar.Increment()
	}
	barP.Wait()
	wg.Wait()
	MergeTs(len(TsList), unitName, chapterNamePath)
	err := os.RemoveAll(temPath)
	if err != nil {
		panic(err)
	}
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
		panic(lookErr)
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
		panic(err)
	}
	//fmt.Println(string(r))
}

func GetSignatureVideoId(UnitId int, cookieStr string, token string) (int, string) {
	client := resty.New()
	client.SetHeaders(map[string]string{
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36",
		"cookie":     cookieStr,
	})
	res, _ := client.R().SetFormData(map[string]string{
		"bizId":       strconv.Itoa(UnitId),
		"bizType":     "1",
		"contentType": "1",
	}).SetQueryParams(map[string]string{
		"csrfKey": token,
	}).
		Post("https://www.icourse163.org/web/j/resourceRpcBean.getResourceToken.rpc")
	var VideoStruct util.Video
	var err = json.Unmarshal([]byte(res.String()), &VideoStruct)
	if err != nil {
		panic(err)
	}
	signature := VideoStruct.Result.VideoSignDto.Signature
	videoId := VideoStruct.Result.VideoSignDto.VideoID
	return videoId, signature
}

func VipGetTsKey(encryptStr string, videoId int) ([]string, string) {
	videoId_ := strconv.Itoa(videoId)
	m3u8 := js.M3u8(encryptStr, videoId_)
	tsCmp := regexp.MustCompile("http.*?ts")
	//获取ts列表
	tsList := tsCmp.FindAllString(m3u8, -1)
	//获取key
	keyCmp := regexp.MustCompile(`URI="(.*?)"`)
	keyUrl := keyCmp.FindStringSubmatch(m3u8)[1]
	client := resty.New()
	client.SetHeaders(map[string]string{
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36",
		"origin":     "https://www.icourse163.org",
		"referer":    "https://www.icourse163.org/",
		"authority":  "mooc2vod.stu.126.net",
	})
	res, _ := client.R().Get(keyUrl)
	text := res.String()
	key := js.Key(text, videoId_)
	return tsList, key
}

func FreeGetTs(M3u8Str string) []string {
	tsCmp := regexp.MustCompile("\\d+.*?ts")
	//获取ts列表
	tsList := tsCmp.FindAllString(M3u8Str, -1)
	return tsList
}

func Download(UnitId int, token string, cookieStr string, unitName string, chapterNamePath string) {
	videoId, signature := GetSignatureVideoId(UnitId, cookieStr, token)
	client := resty.New()
	client.SetHeader("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36")
	res, _ := client.R().SetQueryParams(map[string]string{
		"videoId":    strconv.Itoa(videoId),
		"signature":  signature,
		"clientType": "1",
	}).Get("https://vod.study.163.com/eds/api/v1/vod/video")
	var VodVideoStruct util.VodVideo
	var err = json.Unmarshal([]byte(res.String()), &VodVideoStruct)
	if err != nil {
		panic(err)
	}
	var quality int
	var videoUrl string
	var k string
	var secondaryEncrypt bool
	for _, video := range VodVideoStruct.Result.Videos {
		quality_ := video.Quality
		if quality_ >= quality {
			quality = quality_
		}
		videoUrl = video.VideoURL
		k = video.K
		secondaryEncrypt = video.SecondaryEncrypt
	}
	if secondaryEncrypt {
		videoToken := js.Token(k)
		res, _ = client.R().SetQueryParams(map[string]string{
			"token": videoToken,
			"t":     strconv.FormatInt(time.Now().UnixMilli(), 10),
		}).Get(videoUrl)
		tsList, key := VipGetTsKey(res.String(), videoId)
		VipVideo(tsList, key, unitName, chapterNamePath)
	} else {
		var baseUrl string
		res, _ = client.R().Get(videoUrl)
		if strings.Contains(videoUrl, "https") {
			baseUrl = videoUrl[:48]
		} else {
			baseUrl = videoUrl[:47]
			baseUrl = strings.Replace(baseUrl, "http", "https", -1)
		}
		tsList := FreeGetTs(res.String())
		for i, j := range tsList {
			tsList[i] = baseUrl + j
		}
		FreeVideo(tsList, unitName, chapterNamePath)
	}

}
