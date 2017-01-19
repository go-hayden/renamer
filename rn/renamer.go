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
	UseRegexp bool
	Toreplace string
	Replaceto string
}

func List(root string, includeType int, match string, useRegexp bool) ([]toolbox.FileInfo, toolbox.Err) {
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
	if useRegexp && len(match) > 0 {
		reg = regexp.MustCompile(match)
	}

	for _, fileinfo := range dir {
		if includeType == ListFileOnly && fileinfo.IsDir() {
			continue
		}

		if includeType == ListDirectoryOnly && !fileinfo.IsDir() {
			continue
		}

		if useRegexp && reg != nil && !reg.MatchString(fileinfo.Name()) {
			continue
		}

		if !useRegexp && len(match) > 0 && strings.Index(fileinfo.Name(), match) < 0 {
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
			if len(rep.Toreplace) == 0 {
				result[idx] = rep.Replaceto
			} else {
				if rep.UseRegexp {
					reg := regexp.MustCompile(rep.Toreplace)
					result[idx] = reg.ReplaceAllString(fn, rep.Replaceto)
				} else {
					result[idx] = strings.Replace(fn, rep.Toreplace, rep.Replaceto, -1)
				}
			}
		}
	}

	distinct := make(map[string]int)
	for idx, fi := range source {
		var timestamp string
		if usetimestamp {
			timestamp = "-" + fi.ModTime().Format("20060102150405")
		}
		fn := result[idx]
		fnabs, fnext := getNameAndExt(fn)
		newfn := fnabs + timestamp + fnext
		count, ok := distinct[newfn]
		if ok {
			distinct[newfn] = count + 1
			newfn = fnabs + timestamp + "-" + strconv.Itoa(count) + fnext
		} else {
			distinct[fn] = 1
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
