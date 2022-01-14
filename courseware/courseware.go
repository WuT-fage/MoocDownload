package courseware

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"regexp"
	"strconv"
	"time"
)

func Download(ContentId int, UnitId int, token string, cookieStr string, path string) {
	client := resty.New()
	res, _ := client.R().SetHeaders(map[string]string{
		"cookie":       cookieStr,
		"user-agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36",
		"content-type": "text/plain",
	}).SetFormData(map[string]string{
		"callCount":       "1",
		"scriptSessionId": "${scriptSessionId}190",
		"httpSessionId":   token,
		"c0-scriptName":   "CourseBean",
		"c0-methodName":   "getLessonUnitLearnVo",
		"c0-id":           "0",
		"c0-param0":       "number:" + strconv.Itoa(ContentId),
		"c0-param1":       "number:3",
		"c0-param2":       "number:0",
		"c0-param3":       "number:" + strconv.Itoa(UnitId),
		"batchId":         strconv.Itoa(int(time.Now().UnixMilli())),
	}).Post("https://www.icourse163.org/dwr/call/plaincall/CourseBean.getLessonUnitLearnVo.dwr")
	cmp := regexp.MustCompile("textUrl:\"(http.*?)\"")
	textUrl := cmp.FindAllStringSubmatch(res.String(), 1)[0][1]
	res, _ = client.R().Get(textUrl)

	ioutil.WriteFile(path, res.Body(), 0666)

	fmt.Println(path, " done")
}
