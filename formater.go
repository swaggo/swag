package swag

import (
	"bytes"
	"crypto/md5"
	"fmt"
	goparser "go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/tabwriter"
)

const SplitTag = "&*"

type Formater struct {
	// debugging output goes here
	debug Debugger

	// excludes excludes dirs and files in SearchDir
	excludes map[string]bool
}

func NewFormater() *Formater {
	formater := &Formater{
		debug: log.New(os.Stdout, "", log.LstdFlags),
	}
	return formater
}

func (f *Formater) FormatAPI(searchDir, excludeDir string) error {
	searchDirs := strings.Split(searchDir, ",")
	for _, searchDir := range searchDirs {
		if _, err := os.Stat(searchDir); os.IsNotExist(err) {
			return fmt.Errorf("dir: %s does not exist", searchDir)
		}
	}
	for _, fi := range strings.Split(excludeDir, ",") {
		fi = strings.TrimSpace(fi)
		if fi != "" {
			fi = filepath.Clean(fi)
			f.excludes[fi] = true
		}
	}

	err := f.formatMultiSearchDir(searchDirs)
	if err != nil {
		return err
	}

	return nil
}

func (f *Formater) formatMultiSearchDir(searchDirs []string) error {
	for _, searchDir := range searchDirs {
		f.debug.Printf("Format API Info, search dir:%s", searchDir)

		err := filepath.Walk(searchDir, f.visit)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Formater) visit(path string, fileInfo os.FileInfo, err error) error {
	if err := f.skip(path, fileInfo); err != nil {
		return err
	} else if fileInfo.IsDir() {
		// skip if file is folder
		return nil
	}

	if strings.HasSuffix(strings.ToLower(path), "_test.go") || filepath.Ext(path) != ".go" {
		// skip if file not has suffix "*.go"
		return nil
	}

	err = f.FormatFile(path)
	if err != nil {
		return fmt.Errorf("ParseFile error:%+v", err)
	}
	return nil
}

// skip skip folder in ('vendor' 'docs' 'excludes' 'hidden folder')
func (f *Formater) skip(path string, fileInfo os.FileInfo) error {
	if fileInfo.IsDir() {
		if fileInfo.Name() == "vendor" || // ignore "vendor"
			fileInfo.Name() == "docs" || // exclude docs
			len(fileInfo.Name()) > 1 && fileInfo.Name()[0] == '.' { // exclude all hidden folder
			return filepath.SkipDir
		}

		if f.excludes != nil {
			if _, ok := f.excludes[path]; ok {
				return filepath.SkipDir
			}
		}
	}
	return nil
}

func (formater *Formater) FormatFile(filepath string) error {
	fileSet := token.NewFileSet()
	fileTree, err := goparser.ParseFile(fileSet, filepath, nil, goparser.ParseComments)
	if err != nil {
		return fmt.Errorf("cannot format file, err: %w path : %s ", err, filepath)
	}

	var (
		formatedComments = bytes.Buffer{}
		oldCommentsMap   = make(map[string]string)
	)

	tabw := tabwriter.NewWriter(&formatedComments, 0, 0, 3, ' ', 0)

	// Read the file
	srcBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("cannot open file, err: %w path : %s ", err, filepath)
	}
	src := string(srcBytes)

	if fileTree.Comments != nil {
		for _, comment := range fileTree.Comments {
			comments := strings.Split(comment.Text(), "\n")
			for _, commentLine := range comments {
				if IsSwagComment(commentLine) || IsBlankComment(commentLine) {
					cmd5 := MD5(commentLine)

					reg := regexp.MustCompile(` {3,}`)
					c := reg.ReplaceAllString(commentLine, "\t")
					oldCommentsMap[cmd5] = commentLine

					fmt.Fprintln(tabw, cmd5+SplitTag+c)
				}
			}
		}
		tabw.Flush()
	}

	// Replace old
	newComments := strings.Split(formatedComments.String(), "\n")
	for _, e := range newComments {
		commentSplit := strings.Split(e, SplitTag)
		if len(commentSplit) == 2 {
			commentHash := commentSplit[0]
			commentContent := commentSplit[1]

			if !IsBlankComment(commentContent) {
				oldComment := oldCommentsMap[commentHash]
				if strings.Contains(src, oldComment) {
					src = strings.Replace(src, oldComment, commentContent, 1)
				}
			}
		}
	}
	return WriteBack(filepath, []byte(src), srcBytes)
}

func WriteBack(filepath string, src, old []byte) error {
	// Write back (use golang/gofmt)
	// make a temporary backup before overwriting original
	bakname, err := backupFile(filepath+".", old, 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, src, 0644)
	if err != nil {
		os.Rename(bakname, filepath)
		return err
	}
	err = os.Remove(bakname)
	if err != nil {
		return err
	}
	return nil
}

func IsSwagComment(comment string) bool {
	lc := strings.ToLower(comment)
	return regexp.MustCompile("@[A-z]+").MatchString(lc)
}

func IsBlankComment(comment string) bool {
	lc := strings.TrimSpace(comment)
	return len(lc) == 0
}

func MD5(msg string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(msg)))
}

const chmodSupported = runtime.GOOS != "windows"

// backupFile writes data to a new file named filename<number> with permissions perm,
// with <number randomly chosen such that the file name is unique. backupFile returns
// the chosen file name.
func backupFile(filename string, data []byte, perm os.FileMode) (string, error) {
	// create backup file
	f, err := ioutil.TempFile(filepath.Dir(filename), filepath.Base(filename))
	if err != nil {
		return "", err
	}
	bakname := f.Name()
	if chmodSupported {
		err = f.Chmod(perm)
		if err != nil {
			f.Close()
			os.Remove(bakname)
			return bakname, err
		}
	}

	// write data to backup file
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}

	return bakname, err
}
