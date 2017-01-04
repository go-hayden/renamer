package main

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
	"regexp"
	renamer "renamer/rn"
	"strconv"
	"strings"

	"github.com/go-hayden/toolbox"
)

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	printSplitLine("第一步：设置路径及文件过滤")
	var files []toolbox.FileInfo
	for {
		files = getRenameFiles(dir)
		printSplitLine("需要替换的文件列表")
		for _, fileInfo := range files {
			println(fileInfo.Name())
		}
		tmp := getInput("继续请输入Y，重新设置请直接回车:")
		if tmp == "Y" {
			break
		}
	}

	println("")
	printSplitLine("第二步：设置替换规则")
	println("注：\n1、选择使用时间戳，则会在文件名之后扩展名之前添加文件的修改时间，例如xx-20060102150405.jpg")
	println("2、可添加多个匹配替换规则，支持正则表达式，匹配替换规则按输入顺序执行")
	var newFileNames []string
	for {
		newFileNames = getNewFileNames(files)
		printSplitLine("重命名信息")
		for idx, fileInfo := range files {
			nfn := newFileNames[idx]
			println(fileInfo.Name() + "  ->  " + nfn)
		}
		tmp := getInput("输入Y开始替换，直接按回车重新设置:")
		if tmp != "Y" {
			continue
		}

		doRename(files, newFileNames)
		break
	}
}

func getRenameFiles(dir string) []toolbox.FileInfo {
	root := dir

	tmp := getInput("请输入目录路径并回车，或直接回车使用当前路径[" + dir + "]:")
	if len(tmp) > 0 {
		root = tmp
	}

	includetype := 0
	tmp = getInput("请输入包含的文件类型(0:文件 1:文件夹 2:全部):")
	if len(tmp) > 0 {
		tmpi, err := strconv.Atoi(tmp)
		if err == nil {
			includetype = tmpi
		}
	}

	match := ""
	tmp = getInput("请输入匹配的文件名，支持正则表达式，直接回车则匹配全部文件:")
	if len(tmp) > 0 {
		match = tmp
	}

	files, err := renamer.List(root, includetype, match)
	if err != nil {
		panic(err)
	}
	return files
}

func getNewFileNames(files []toolbox.FileInfo) []string {
	tmp := getInput("是否使用时间戳(Y/n):")
	usetimestamp := tmp == "Y"

	regs := make([]*renamer.RenameReplaceInfo, 0, 10)
	for {
		tmp = getInput("输入要替换的文件名以及目标文件名，以空格分隔，若只输入一个，则替换全部文件名，不输入直接回车则退出替换设置:\n")
		if len(tmp) == 0 && len(regs) == 0 {
			println("请输入至少一个替换规则")
			continue
		} else if len(tmp) == 0 {
			break
		}

		info := getRenameReplaceInfo(tmp)
		regs = append(regs, info)
	}

	return renamer.GenerateNewNames(files, usetimestamp, regs)
}

func getRenameReplaceInfo(line string) *renamer.RenameReplaceInfo {
	info := new(renamer.RenameReplaceInfo)
	sp := strings.Split(line, " ")
	res := make([]string, 0, 2)
	for _, item := range sp {
		if len(res) >= 2 {
			break
		}
		if len(item) > 0 {
			res = append(res, item)
		}
	}
	if len(res) == 2 {
		info.Toreplace = regexp.MustCompile(res[0])
		info.Replaceto = res[1]
	} else if len(res) == 1 {
		info.Replaceto = res[0]
	}
	return info
}

func doRename(files []toolbox.FileInfo, newFileNames []string) {
	println("开始重命名...")
	for idx, fileInfo := range files {
		fileName := newFileNames[idx]
		filePath := path.Join(path.Dir(fileInfo.FilePath()), fileName)
		err := renamer.DoRename(fileInfo.FilePath(), filePath)
		msg := fileInfo.Name() + "  ->  " + fileName
		if err != nil {
			msg += "  Fail!"
		} else {
			msg += " Done!"
		}
		println(msg)
	}
}

func getInput(msg string) string {
	print(msg)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadBytes('\n')
	if err != nil || input == nil {
		return ""
	}

	tmp := string(input[0 : len(input)-1])
	return strings.TrimSpace(tmp)
}

func printSplitLine(msg string) {
	println("======== " + msg + " ========")
}
