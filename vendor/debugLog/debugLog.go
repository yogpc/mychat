package debugLog

import (
	"os"
	"log"
)

func info(msg string, filename string) error {
	var f *os.File
	var errF error
	if checkFileIsExist(filename) { //如果文件存在
		f, errF = os.OpenFile(filename, os.O_APPEND, 0777) //打开文件
		//fmt.Println("文件存在")
	} else {
		f, errF = os.Create(filename) //创建文件
		//fmt.Println("文件不存在")
	}
	if errF != nil {
		return errF
	}
	defer f.Close()
	// 创建一个日志对象
	debugLog := log.New(f,"[Info]",log.Lmicroseconds)
	debugLog.Println(msg)
	return nil
}


/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}