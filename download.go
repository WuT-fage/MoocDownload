package main

import (
	"fmt"
	"github.com/spf13/viper"
	"moocVip/courseware"
	"moocVip/util"
	"moocVip/video"
	"os"
)

func Download(InfoStruct util.MyMocTermDto, cookieStr string, token string) {
	basePath, _ := os.Getwd()
	courseName := InfoStruct.CourseName
	courseName = util.RemoveInvalidChar(courseName)
	courseNamePath := fmt.Sprintf("%s\\download\\%s", basePath, courseName)
	util.PathExists(courseNamePath)
	videoBool := viper.GetInt("download.video")
	coursewareBool := viper.GetInt("download.courseware")
	paperBool := viper.GetInt("download.paper")
	for _, chapter := range InfoStruct.Chapters {
		chapterName := chapter.ChapterName
		chapterName = util.RemoveInvalidChar(chapterName)
		chapterNamePath := fmt.Sprintf("%s\\%s", courseNamePath, chapterName)
		util.PathExists(chapterNamePath)
		for _, unit := range chapter.MyUnits {
			contentType := unit.ContentType
			UnitId := unit.UnitId
			ContentId := unit.ContentId
			unitName := unit.UnitName
			unitName = util.RemoveInvalidChar(unitName)
			switch contentType {
			case 1:
				if videoBool == 1 {
					path := fmt.Sprintf("%s\\%s.mp4", chapterNamePath, unitName)
					_, err := os.Stat(path)
					if err != nil {
						video.Download(UnitId, token, cookieStr, unitName, chapterNamePath)
					}
				}

			case 3:
				if coursewareBool == 1 {
					path := fmt.Sprintf("%s\\%s.pdf", chapterNamePath, unitName)
					_, err := os.Stat(path)
					if err != nil {
						courseware.Download(ContentId, UnitId, token, cookieStr, path)
					}
				}
			case 5:
				if paperBool == 1 {

				}
			}
		}
	}

}
