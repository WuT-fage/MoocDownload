package mooc

import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"

	"MoocDownload/mooc/download"
	"MoocDownload/mooc/js"
	"MoocDownload/mooc/model"
)

type MoocSession struct {
	Session  *requests.Session
	Cookie   string
	Token    string
	MemberId string
}

// 检查当前cookie状态
func (This *MoocSession) CheckStatus() bool {
	headers := url.NewHeaders()
	headers.Set("edu-script-token", This.Token)
	headers.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36")
	// headers.Set("cookie", This.Cookie)
	params := url.NewParams()
	params.Set("csrfKey", This.Token)
	data := url.NewData()
	data.Set("memberId", This.MemberId)
	req := url.NewRequest()
	req.Headers = headers
	req.Params = params
	req.Data = data
	res, _ := This.Session.Post("https://www.icourse163.org/web/j/memberBean.getMocMemberPersonalDtoById.rpc", req)

	var StatusStruct model.Status
	var err = json.Unmarshal([]byte(res.Text), &StatusStruct)
	if err != nil {
		panic(err)
	}
	if StatusStruct.Code == 0 {
		color.Cyan.Printf("\n欢迎您:%s\n", StatusStruct.Result.NickName)
		if len(StatusStruct.Result.Description) != 0 {
			color.Blue.Printf("\n你的座右铭:%s\n", StatusStruct.Result.Description)
		}
		return true
	} else {
		color.Red.Println("\ncookie已失效，请更新您的cookie!\n")
		time.Sleep(5 * time.Second)
		os.Exit(1)
		return false
	}
}

func (This *MoocSession) GetLastLearnedMocTermDto(tid string) string {
	headers := url.NewHeaders()
	headers.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36")
	headers.Set("edu-script-token", This.Token)
	params := url.NewParams()
	params.Set("csrfKey", This.Token)
	data := url.NewData()
	data.Set("termId", tid)
	req := url.NewRequest()
	req.Headers = headers
	req.Params = params
	req.Data = data
	res, err := This.Session.Post("https://www.icourse163.org/web/j/courseBean.getLastLearnedMocTermDto.rpc", req)
	if err != nil {
		panic(err)
	}
	jsonStr := res.Text
	return jsonStr
}

func (This *MoocSession) GetSignatureVideoId(UnitId int) (int, string) {
	headers := url.NewHeaders()
	headers.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36")
	data := url.NewData()
	data.Set("bizId", strconv.Itoa(UnitId))
	data.Set("bizType", "1")
	data.Set("contentType", "1")
	params := url.NewParams()
	params.Set("csrfKey", This.Token)
	req := url.NewRequest()
	req.Headers = headers
	req.Params = params
	req.Data = data
	res, _ := This.Session.Post("https://www.icourse163.org/web/j/resourceRpcBean.getResourceToken.rpc", req)
	var VideoStruct model.Video
	var err = json.Unmarshal([]byte(res.Text), &VideoStruct)
	if err != nil {
		panic(err)
	}
	signature := VideoStruct.Result.VideoSignDto.Signature
	videoId := VideoStruct.Result.VideoSignDto.VideoID
	return videoId, signature
}

// 视频下载
func (This *MoocSession) Video(UnitId int, unitName, chapterNamePath string) {
	videoId, signature := This.GetSignatureVideoId(UnitId)
	headers := url.NewHeaders()
	headers.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36")
	params := url.NewParams()
	params.Set("videoId", strconv.Itoa(videoId))
	params.Set("signature", signature)
	params.Set("clientType", "1")
	res, _ := This.Session.Get("https://vod.study.163.com/eds/api/v1/vod/video", &url.Request{
		Headers: headers,
		Params:  params,
	})
	var VodVideoStruct model.VodVideo
	var err = json.Unmarshal([]byte(res.Text), &VodVideoStruct)
	if err != nil {
		panic(err)
	}
	// var quality int
	var videoUrl string
	var k string
	var secondaryEncrypt bool
	var format string
	Count := len(VodVideoStruct.Result.Videos)
	if Count%3 == 0 {
		videos := VodVideoStruct.Result.Videos
		video := videos[2]
		format = video.Format
		videoUrl = video.VideoURL
		k = video.K
		secondaryEncrypt = video.SecondaryEncrypt

	} else {
		videos := VodVideoStruct.Result.Videos
		video := videos[Count%3-1]
		format = video.Format
		videoUrl = video.VideoURL
		k = video.K
		secondaryEncrypt = video.SecondaryEncrypt
	}
	if secondaryEncrypt {
		videoToken := js.Token(k)
		params := url.NewParams()
		params.Set("token", videoToken)
		params.Set("t", strconv.FormatInt(time.Now().UnixMilli(), 10))
		res1, _ := This.Session.Get(videoUrl, &url.Request{
			Params: params,
		})
		tsList, key := download.VipGetTsKey(res1.Text, videoId)
		download.VipVideo(tsList, key, unitName, chapterNamePath)
	} else {
		switch format {
		case "hls":
			var baseUrl string
			res0, _ := This.Session.Get(videoUrl, nil)
			if strings.Contains(videoUrl, "https") {
				baseUrl = videoUrl[:48]
			} else {
				baseUrl = videoUrl[:47]
				baseUrl = strings.Replace(baseUrl, "http", "https", 1)
			}
			tsList := download.FreeGetTs(res0.Text)
			for i, j := range tsList {
				tsList[i] = baseUrl + j
			}
			download.FreeVideo(tsList, unitName, chapterNamePath)
		case "mp4":
			download.Mp4Flv(videoUrl, unitName, chapterNamePath, false)
		case "flv":
			download.Mp4Flv(videoUrl, unitName, chapterNamePath, true)
		}

	}
}

// 文本资料下载
func (This *MoocSession) Courseware(ContentId int, UnitId int, path string) {
	headers := url.NewHeaders()
	headers.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36")
	data := url.NewData()
	data.Set("callCount", "1")
	data.Set("scriptSessionId", "${scriptSessionId}190")
	data.Set("httpSessionId", This.Token)
	data.Set("c0-scriptName", "CourseBean")
	data.Set("c0-methodName", "getLessonUnitLearnVo")
	data.Set("c0-id", "0")
	data.Set("c0-param0", "number:"+strconv.Itoa(ContentId))
	data.Set("c0-param1", "number:3")
	data.Set("c0-param2", "number:0")
	data.Set("c0-param3", "number:"+strconv.Itoa(UnitId))
	data.Set("batchId", strconv.Itoa(int(time.Now().UnixMilli())))
	res, _ := This.Session.Post("https://www.icourse163.org/dwr/call/plaincall/CourseBean.getLessonUnitLearnVo.dwr", &url.Request{
		Headers: headers,
		Data:    data,
	})
	cmp := regexp.MustCompile("textUrl:\"(http.*?)\"")
	textUrl := cmp.FindAllStringSubmatch(res.Text, 1)[0][1]
	download.Text(textUrl, path)
}
