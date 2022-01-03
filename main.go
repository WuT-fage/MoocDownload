package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/gookit/color"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"moocVip/util"
	"os"
	"regexp"
	"syscall"
	"time"
)

//go build -o main.exe
func main() {
	Init()
	cookieStr := ReadCookie()
	CookieMap := util.CookieToMap(cookieStr)
	token := CookieMap["NTESSTUDYSI"]
	NETEASE_WDA_UID := CookieMap["NETEASE_WDA_UID"]
	cmp := regexp.MustCompile("(\\d+)#")
	memberId := cmp.FindStringSubmatch(NETEASE_WDA_UID)[1]
	CheckStatus(cookieStr, token, memberId)
	var Link string
	color.Red.Printf("\n请把链接粘贴到处:")
	//fmt.Printf("\n请把链接粘贴到处:")
	fmt.Scanln(&Link)
	cmp = regexp.MustCompile("tid=(\\d+)")
	tid := cmp.FindStringSubmatch(Link)[1]
	jsonStr := GetLastLearnedMocTermDto(cookieStr, token, tid)
	InfoStruct := HandleJsonStr(jsonStr)
	Download(InfoStruct, cookieStr, token)
}

func Init() {
	Table()
	basePath, _ := os.Getwd()
	Path := fmt.Sprintf("%s\\%s", basePath, "download")
	util.PathExists(Path)
}

func ReadCookie() string {
	basePath, _ := os.Getwd()
	cookiePath := fmt.Sprintf("%s\\cookie.txt", basePath)
	data, err := ioutil.ReadFile(cookiePath)
	if err != nil {
		os.Create(cookiePath)
		color.Red.Println("\n请把你的cookie复制到cookie.txt文件中并重新运行本程序!")
		time.Sleep(5 * time.Second)
		syscall.Exit(0)
	}
	if len(data) == 0 {
		color.Green.Println("\ncookie.txt文件中没有cookie,请您粘贴cookie并重新运行本程序!")
		time.Sleep(5 * time.Second)
		syscall.Exit(0)
	}
	return string(data)
}

//检查当前cookie状态
func CheckStatus(cookieStr string, token string, memberId string) {
	client := resty.New()
	client.SetHeaders(map[string]string{
		"user-agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36",
		"cookie":           cookieStr,
		"edu-script-token": token,
	})
	res, _ := client.R().SetQueryParam("csrfKey", token).SetFormData(map[string]string{
		"memberId": memberId,
	}).Post("https://www.icourse163.org/web/j/memberBean.getMocMemberPersonalDtoById.rpc")
	var StatusStruct util.Status
	var err = json.Unmarshal([]byte(res.String()), &StatusStruct)
	if err != nil {
		panic(err)
	}
	if StatusStruct.Code == 0 {
		color.Cyan.Printf("\n欢迎您:%s\n", StatusStruct.Result.NickName)
		if len(StatusStruct.Result.Description) != 0 {
			color.Blue.Printf("\n你的座右铭:%s\n", StatusStruct.Result.Description)
		}
	} else {
		color.Red.Println("\ncookie.txt文件中的cookie已失效，请更新您的cookie!\n")
		time.Sleep(5 * time.Second)
		syscall.Exit(0)
	}

}

func Table() {
	data := [][]string{
		[]string{"慕", "", "", ""},
		[]string{"课", "", "", ""},
		[]string{"下", "Esword", "Spiders and AI", "https://github.com/Esword618/moocVip"},
		[]string{"载", "", "", ""},
		[]string{"器", "", "", ""},
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"v2.7", "Author", "公众号", "Github", ""})
	table.SetFooter([]string{"", "", "♥ Cmf", "Super invincible little cute", ""}) // Add Footer
	table.SetBorder(false)                                                         // Set Border to false

	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.BgBlueColor, tablewriter.Bold, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.BgCyanColor, tablewriter.Bold, tablewriter.FgBlackColor},
		tablewriter.Colors{},
	)

	table.SetColumnColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiYellowColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{},
	)

	table.SetFooterColor(
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
		tablewriter.Colors{tablewriter.FgHiYellowColor},
		tablewriter.Colors{},
	)

	table.AppendBulk(data)
	table.Render()
}

func GetLastLearnedMocTermDto(cookieStr string, token string, tid string) string {

	client := resty.New()
	client.SetHeaders(map[string]string{
		"user-agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36",
		"cookie":           cookieStr,
		"edu-script-token": token,
	})
	res, err := client.R().SetQueryParam("csrfKey", token).
		SetFormData(map[string]string{
			"termId": tid,
		}).Post("https://www.icourse163.org/web/j/courseBean.getLastLearnedMocTermDto.rpc")
	if err != nil {
		panic(err)
	}
	jsonStr := res.String()
	return jsonStr
}

func HandleJsonStr(jsonStr string) util.MyMocTermDto {
	var STRUT util.LastLearnedMocTermDto
	var InfoStruct util.MyMocTermDto
	var contentType int
	var err = json.Unmarshal([]byte(jsonStr), &STRUT)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%+v", STRUT)
	courseName := STRUT.Result.MocTermDto.CourseName
	InfoStruct.CourseName = courseName
	chapters := STRUT.Result.MocTermDto.Chapters
	for _, chapter := range chapters {
		//fmt.Println(i, "------", chapter)
		var myChapter util.MyChapter
		lessons := chapter.Lessons
		contentType = chapter.ContentType
		chapterName := chapter.Name
		myChapter.ContentType = contentType
		myChapter.ChapterName = chapterName
		if contentType == 2 {
			var myUnit util.MyUnit
			myUnit.ContentType = contentType
			myUnit.UnitName = chapterName
			myUnit.UnitId = chapter.ID
			myUnit.ContentId = 0

			myChapter.MyUnits = append(myChapter.MyUnits, myUnit)

		} else {
			for _, lesson := range lessons {
				units := lesson.Units
				for _, unit := range units {
					var myUnit util.MyUnit
					contentType = unit.ContentType
					switch contentType {
					case 1:
						myUnit.ContentType = contentType
						myUnit.ContentId = 0
						myUnit.UnitName = unit.Name
						myUnit.UnitId = unit.ID
					case 3:
						myUnit.ContentType = contentType
						myUnit.ContentId = unit.ContentID
						myUnit.UnitName = unit.Name
						myUnit.UnitId = unit.ID
					case 5:
						myUnit.ContentType = contentType
						myUnit.ContentId = unit.ContentID
						myUnit.UnitName = unit.Name
						myUnit.UnitId = 0
					}

					myChapter.MyUnits = append(myChapter.MyUnits, myUnit)

				}
			}
			InfoStruct.Chapters = append(InfoStruct.Chapters, myChapter)
		}

	}
	//InfoStruct.Chapters = MyChaptersList
	//fmt.Println(InfoStruct)
	//for _, chapter := range InfoStruct.Chapters {
	//	fmt.Println(chapter.ChapterName)
	//	fmt.Println(chapter.ContentType)
	//	for _, i := range chapter.MyUnits {
	//		fmt.Println(i)
	//	}
	//	fmt.Println("\n\r---------\n\r")
	//}
	return InfoStruct
}

//goreleaser --snapshot --skip-publish --rm-dist
