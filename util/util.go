package util

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func Bs64Decode(Str string) string {
	resBytes, err := base64.StdEncoding.DecodeString(Str)
	if err != nil {
		panic("bs64无法解码")
	}
	return string(resBytes)
}

func Bs64Encode(Str string) string {
	return base64.StdEncoding.EncodeToString([]byte(Str))
}

func Iv(index int) []byte {
	indexStr := strconv.Itoa(index)
	length := len(indexStr)
	iv := make([]byte, 0, 16)
	for i := 0; i < 16-length; i++ {
		iv = append(iv, byte(0))
	}
	for i := 0; i < length; i++ {
		num, _ := strconv.Atoi(string(indexStr[i]))
		iv = append(iv, byte(num))
	}
	return iv
}

//PathExists 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		// 创建文件夹
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
		} else {
			return true, nil
		}
	}
	return false, err
}

func CookieToMap(CookieStr string) map[string]string {
	CookieMap := make(map[string]string)
	CookieList := strings.Split(CookieStr, "; ")
	for _, i := range CookieList {
		iList := strings.Split(i, "=")
		CookieMap[iList[0]] = iList[1]
	}
	//fmt.Println(CookieMap)
	return CookieMap
	//cmp := regexp.MustCompile("; ")
	//fmt.Println(cmp.FindAllString(CookieStr, -1))
}

func RemoveInvalidChar(oldStr string) string {
	cmp := regexp.MustCompile("]|\\[|：|:|\\*|\\?|\t|\\x0b|/|")
	newStr := cmp.ReplaceAllString(oldStr, "")
	newStr = strings.Replace(newStr, "/", "", -1)
	return newStr
}

func HttpCookieToMap(httpCookie []*http.Cookie) map[string]string {
	CookieMap := make(map[string]string)
	for _, cookie := range httpCookie {
		CookieMap[cookie.Name] = cookie.Value
	}
	return CookieMap
}

func CookieMapTOStr(CookieMap map[string]string) string {
	var CookieStr string
	for key, value := range CookieMap {
		CookieStr += key + "=" + value + "; "
	}
	fmt.Println(CookieStr)
	return CookieStr[:len(CookieStr)-2]
}
