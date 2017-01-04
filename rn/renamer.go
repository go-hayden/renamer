package rn

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-hayden/toolbox"
)

const (
	ListFileOnly = iota
	ListDirectoryOnly
	ListAll
)

type RenameReplaceInfo struct {
	Toreplace *regexp.Regexp
	Replaceto string
}

func List(root string, includeType int, matchreg string) ([]toolbox.FileInfo, toolbox.Err) {
	if !toolbox.DirectoryExists(root) {
		msg := "Directory path '" + root + "' is not exist!"
		return nil, toolbox.NewErrWithMessage(toolbox.ErrCodeNotExist, msg)
	}

	dir, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, toolbox.NewErr(toolbox.ErrCodeUnknown, err)
	}

	result := make([]toolbox.FileInfo, 0, len(dir))
	var reg *regexp.Regexp
	if len(matchreg) > 0 {
		reg = regexp.MustCompile(matchreg)
	}

	for _, fileinfo := range dir {
		if includeType == ListFileOnly && fileinfo.IsDir() {
			continue
		}

		if includeType == ListDirectoryOnly && !fileinfo.IsDir() {
			continue
		}

		if reg != nil && !reg.MatchString(fileinfo.Name()) {
			continue
		}

		newFI := new(toolbox.FileInfoBase)
		newFI.FileInfo = fileinfo
		newFI.Path = path.Join(root, fileinfo.Name())
		result = append(result, newFI)
	}
	return result, nil
}

func GenerateNewNames(source []toolbox.FileInfo, usetimestamp bool, replace []*RenameReplaceInfo) []string {
	result := make([]string, len(source), len(source))

	for idx, fi := range source {
		result[idx] = fi.Name()
	}

	for _, rep := range replace {
		if len(strings.TrimSpace(rep.Replaceto)) == 0 {
			continue
		}
		for idx, fn := range result {
			if rep.Toreplace == nil {
				result[idx] = rep.Replaceto
			} else {
				result[idx] = rep.Toreplace.ReplaceAllString(fn, rep.Replaceto)
			}
		}
	}

	dupmap := make(map[string]int)
	for idx, fi := range source {
		var timestamp string
		if usetimestamp {
			timestamp = "-" + fi.ModTime().Format("20060102150405")
		}
		fn := result[idx]
		fnabs, fnext := getNameAndExt(fn)
		newfn := fnabs + timestamp + fnext
		count, ok := dupmap[newfn]
		if ok {
			dupmap[newfn] = count + 1
			newfn = fnabs + timestamp + "-" + strconv.Itoa(count) + fnext
		} else {
			dupmap[fn] = 1
		}
		result[idx] = newfn
	}
	return result
}

func DoRename(srcpath string, destpath string) toolbox.Err {
	err := os.Rename(srcpath, destpath)
	if err != nil {
		return toolbox.NewErr(toolbox.ErrCodeUnknown, err)
	}
	return nil
}

func getNameAndExt(filename string) (string, string) {
	ext := path.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	if len(name) == 0 && len(ext) > 0 {
		name = ext
		ext = ""
	}
	return name, ext
}
