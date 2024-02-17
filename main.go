package main

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"os"
	"strconv"
	"strings"
)

var classes []string

func buildIndex() {
	// 用于构建搜索索引 防止重复IO读取
	files, err := os.ReadDir("data")
	if err != nil {
		fmt.Println("异常！无法获取课表数据 请存放在 ./data/目录下, 错误信息：", err)
		return
	}

	for fileIdx := range files {
		file := files[fileIdx]
		if !file.IsDir() {
			classes = append(classes, strings.Replace(file.Name(), ".wakeup_schedule", "", 1))
		}
	}
	idx, err := os.OpenFile(".index", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("创建搜索缓存失败. 错误: ", err)
		return
	}
	for classIdx := range classes {
		bytes := []byte(classes[classIdx] + "\n")
		_, _ = idx.Write(bytes)
	}
	_ = idx.Close()

	idx, err = os.OpenFile(".index", os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("读取搜索缓存失败, 错误: ", err)
		return
	}

	buf := make([]byte, 10240)
	_, err = idx.Read(buf)
	if err != nil {
		return
	}

	classes = strings.Split(string(buf), "\n")
}

func main() {
	buildIndex()

	r := gin.Default()

	r.GET("/", index)
	r.POST("/search", search)
	r.GET("/courseTable", courseTable)
	fmt.Println("http://127.0.0.1:5124/")
	err := r.Run(":5124")
	if err != nil {
		fmt.Println("异常退出：", err)
		return
	}
}

func index(ctx *gin.Context) {
	tmp, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Println("模板解析错误!", err)
		return
	}
	err = tmp.Execute(ctx.Writer, nil)
	if err != nil {
		return
	}
}

func search(ctx *gin.Context) {
	word := ctx.PostForm("word")
	var msg string
	results := make([]string, 0)
	resIsNull := false

	if word != "" {
		for classIdx := range classes {
			class := classes[classIdx]
			if strings.Contains(class, word) {
				results = append(results, class)
			}
		}

		if len(results) == 0 {
			resIsNull = true
			msg = "未找到搜索结果!"
		} else {
			msg = "搜索到" + strconv.Itoa(len(results)) + "条结果如下："
		}
	} else {
		msg = "请输入要搜索的班级, 不能为空!"
	}
	data := map[string]interface{}{
		"msg":       msg,
		"res":       results,
		"resIsNull": !resIsNull,
	}
	tmp, err := template.ParseFiles("templates/search.html")
	if err != nil {
		fmt.Println("模板解析错误!", err)
		return
	}
	err = tmp.Execute(ctx.Writer, data)
	if err != nil {
		fmt.Println("模板填充错误: ", err)
		return
	}
}

func courseTable(ctx *gin.Context) {
	tmp, err := template.ParseFiles("templates/courseTable.html")
	if err != nil {
		fmt.Println("异常！读取courseTable模板失败，", err)
		return
	}

	var msg string
	class := ctx.Query("class")
	if class == "" {
		msg = "不存在班级 请重新搜索!"
	} else {
		file, err := os.OpenFile("data/"+class+".wakeup_schedule", os.O_RDONLY, 0666)
		defer func(file *os.File) {
			_ = file.Close()
		}(file)
		if err != nil {
			msg = "不存在班级 请重新搜索!"
		}
		var buf string
		br := bufio.NewReader(file)

		for {
			data, _, err := br.ReadLine()
			if err == io.EOF {
				break
			}
			buf = buf + string(data)
		}

		if err != nil {
			msg = "不存在班级 请重新搜索!"
		} else {
			msg = ""
		}

		// Todo
	}
	data := map[string]interface{}{
		"msg": msg,
	}
	err = tmp.Execute(ctx.Writer, data)
	if err != nil {
		fmt.Println("courseTable 返回异常, ", err)
		return
	}
}
