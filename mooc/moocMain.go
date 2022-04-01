package mooc

import (
	"fmt"
	"os"
	"regexp"

	"github.com/gookit/color"
	"github.com/spf13/viper"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"

	"MoocDownload/mooc/model"
	"MoocDownload/mooc/utils"
)

var moocSession *MoocSession

// https://www.icourse163.org/learn/HIT-1002533005?tid=1467082464#/learn/announce
// 人家的，下面
// https://www.icourse163.org/learn/NJTU-1002080018?tid=1467143481#/learn/announce
func MoocMain() {
	cookieStr := utils.ReadCookie()
	color.Red.Printf("\n请把链接粘贴到处:")
	var Link string
	fmt.Scanln(&Link)
	re := regexp.MustCompile("tid=(\\d+)")
	tid := re.FindStringSubmatch(Link)[1]
	token := utils.CookieToMap(cookieStr)["NTESSTUDYSI"]
	moocSession = &MoocSession{
		Session:  requests.NewSession(),
		Cookie:   cookieStr,
		Token:    token,
		MemberId: tid,
	}
	moocSession.Session.Cookies = url.ParseCookies("https://www.icourse163.org/", cookieStr)
	moocSession.CheckStatus()
	jsonStr := moocSession.GetLastLearnedMocTermDto(tid)
	InfoStruct := utils.HandleJsonStr(jsonStr)
	Download(InfoStruct)
}

func Download(InfoStruct model.MyMocTermDto) {
	basePath, _ := os.Getwd()
	courseName := InfoStruct.CourseName
	courseName = utils.RemoveInvalidChar(courseName)
	courseNamePath := fmt.Sprintf("%s\\download\\%s", basePath, courseName)
	utils.PathExists(courseNamePath)
	videoBool := viper.GetInt("download.video")
	coursewareBool := viper.GetInt("download.courseware")
	paperBool := viper.GetInt("download.paper")
	for _, chapter := range InfoStruct.Chapters {
		chapterName := chapter.ChapterName
		chapterName = utils.RemoveInvalidChar(chapterName)
		chapterNamePath := fmt.Sprintf("%s\\%s", courseNamePath, chapterName)
		utils.PathExists(chapterNamePath)
		temPath := fmt.Sprintf("%s\\tem", chapterNamePath)
		utils.PathExists(temPath)
		for _, unit := range chapter.MyUnits {
			contentType := unit.ContentType
			UnitId := unit.UnitId
			ContentId := unit.ContentId
			unitName := unit.UnitName
			unitName = utils.RemoveInvalidChar(unitName)
			switch contentType {
			case 1:
				if videoBool == 1 {
					path := fmt.Sprintf("%s\\%s.mp4", chapterNamePath, unitName)
					_, err := os.Stat(path)
					if err != nil {
						moocSession.Video(UnitId, unitName, chapterNamePath)
					}
				}

			case 3:
				if coursewareBool == 1 {
					path := fmt.Sprintf("%s\\%s.pdf", chapterNamePath, unitName)
					_, err := os.Stat(path)
					if err != nil {
						moocSession.Courseware(ContentId, UnitId, path)
					}
				}
			case 5:
				if paperBool == 1 {

				}
			}
		}
		err := os.RemoveAll(temPath)
		if err != nil {
			panic(err)
		}
	}
}
