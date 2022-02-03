package swag

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/tabwriter"
)

const splitTag = "&*"

// Formater implements a formater for Go source files.
type Formater struct {
	// debugging output goes here
	debug Debugger

	// excludes excludes dirs and files in SearchDir
	excludes map[string]struct{}

	mainFile string
}

// NewFormater create a new formater instance.
func NewFormater() *Formater {
	formater := &Formater{
		debug:    log.New(os.Stdout, "", log.LstdFlags),
		excludes: make(map[string]struct{}),
	}
	return formater
}

// FormatAPI format the swag comment.
func (f *Formater) FormatAPI(searchDir, excludeDir, mainFile string) error {
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
			f.excludes[fi] = struct{}{}
		}
	}

	// parse main.go
	absMainAPIFilePath, err := filepath.Abs(filepath.Join(searchDirs[0], mainFile))
	if err != nil {
		return err
	}
	err = f.FormatMain(absMainAPIFilePath)
	if err != nil {
		return err
	}
	f.mainFile = mainFile

	err = f.formatMultiSearchDir(searchDirs)
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
	if err := walkWith(f.excludes, false)(path, fileInfo); err != nil {
		return err
	} else if fileInfo.IsDir() {
		// skip if file is folder
		return nil
	}

	if strings.HasSuffix(strings.ToLower(path), "_test.go") || filepath.Ext(path) != ".go" {
		// skip if file not has suffix "*.go"
		return nil
	}
	if strings.HasSuffix(strings.ToLower(path), f.mainFile) {
		// skip main file
		return nil
	}

	err = f.FormatFile(path)
	if err != nil {
		return fmt.Errorf("ParseFile error:%+v", err)
	}
	return nil
}

// FormatMain format the main.go comment.
func (f *Formater) FormatMain(mainFilepath string) error {
	fileSet := token.NewFileSet()
	astFile, err := goparser.ParseFile(fileSet, mainFilepath, nil, goparser.ParseComments)
	if err != nil {
		return fmt.Errorf("cannot format file, err: %w path : %s ", err, mainFilepath)
	}
	var (
		formatedComments = bytes.Buffer{}
		// CommentCache
		oldCommentsMap = make(map[string]string)
	)

	if astFile.Comments != nil {
		for _, comment := range astFile.Comments {
			formatFuncDoc(comment.List, &formatedComments, oldCommentsMap)
		}
	}

	return writeFormatedComments(mainFilepath, formatedComments, oldCommentsMap)
}

// FormatFile format the swag comment in go function.
func (f *Formater) FormatFile(filepath string) error {
	fileSet := token.NewFileSet()
	astFile, err := goparser.ParseFile(fileSet, filepath, nil, goparser.ParseComments)
	if err != nil {
		return fmt.Errorf("cannot format file, err: %w path : %s ", err, filepath)
	}

	var (
		formatedComments = bytes.Buffer{}
		// CommentCache
		oldCommentsMap = make(map[string]string)
	)

	for _, astDescription := range astFile.Decls {
		astDeclaration, ok := astDescription.(*ast.FuncDecl)
		if ok && astDeclaration.Doc != nil && astDeclaration.Doc.List != nil {
			formatFuncDoc(astDeclaration.Doc.List, &formatedComments, oldCommentsMap)
		}
	}

	return writeFormatedComments(filepath, formatedComments, oldCommentsMap)
}

func writeFormatedComments(filepath string, formatedComments bytes.Buffer, oldCommentsMap map[string]string) error {
	// Replace the file
	// Read the file
	srcBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("cannot open file, err: %w path : %s ", err, filepath)
	}
	replaceSrc := string(srcBytes)
	newComments := strings.Split(formatedComments.String(), "\n")
	for _, e := range newComments {
		commentSplit := strings.Split(e, splitTag)
		if len(commentSplit) == 2 {
			commentHash, commentContent := commentSplit[0], commentSplit[1]

			if !isBlankComment(commentContent) {
				replaceSrc = strings.Replace(replaceSrc, oldCommentsMap[commentHash], commentContent, 1)
			}
		}
	}
	return writeBack(filepath, []byte(replaceSrc), srcBytes)
}

func formatFuncDoc(commentList []*ast.Comment, formatedComments io.Writer, oldCommentsMap map[string]string) {
	tabw := tabwriter.NewWriter(formatedComments, 0, 0, 2, ' ', 0)

	for _, comment := range commentList {
		commentLine := comment.Text
		if isSwagComment(commentLine) || isBlankComment(commentLine) {
			cmd5 := fmt.Sprintf("%x", md5.Sum([]byte(commentLine)))

			// Find the separator and replace to \t
			c := separatorFinder(commentLine, '\t')
			oldCommentsMap[cmd5] = commentLine

			// md5 + splitTag + srcCommentLine
			// eg. xxx&*@Description get struct array
			_, _ = fmt.Fprintln(tabw, cmd5+splitTag+c)
		}
	}
	// format by tabwriter
	_ = tabw.Flush()
}

// Check of @Param @Success @Failure @Response @Header
var specialTagForSplit = map[string]byte{
	paramAttr:    1,
	successAttr:  1,
	failureAttr:  1,
	responseAttr: 1,
	headerAttr:   1,
}

var skipChar = map[byte]byte{
	'"': 1,
	'(': 1,
	'{': 1,
	'[': 1,
}

var skipCharEnd = map[byte]byte{
	'"': 1,
	')': 1,
	'}': 1,
	']': 1,
}

func separatorFinder(comment string, rp byte) string {
	commentBytes := []byte(comment)
	commentLine := strings.TrimSpace(strings.TrimLeft(comment, "/"))
	if len(commentLine) == 0 {
		return ""
	}
	attribute := strings.Fields(commentLine)[0]
	attrLen := strings.Index(comment, attribute) + len(attribute)
	attribute = strings.ToLower(attribute)
	var i = attrLen

	if _, ok := specialTagForSplit[attribute]; ok {
		var skipFlag bool
		for ; i < len(commentBytes); i++ {
			if !skipFlag && commentBytes[i] == ' ' {
				j := i
				for j < len(commentBytes) && commentBytes[j] == ' ' {
					j++
				}
				commentBytes = replaceRange(commentBytes, i, j, rp)
			}
			if _, ok := skipChar[commentBytes[i]]; ok && !skipFlag {
				skipFlag = true
			} else if _, ok := skipCharEnd[commentBytes[i]]; ok && skipFlag {
				skipFlag = false
			}
		}
	} else {
		for i < len(commentBytes) && commentBytes[i] == ' ' {
			i++
		}
		if i >= len(commentBytes) {
			return comment
		}
		commentBytes = replaceRange(commentBytes, attrLen, i, rp)
	}
	return string(commentBytes)
}

func replaceRange(s []byte, start, end int, new byte) []byte {
	if start > end || end < 1 {
		return s
	}
	if end > len(s) {
		end = len(s)
	}
	s = append(s[:start], s[end-1:]...)
	s[start] = new
	return s
}

var swagCommentExpression = regexp.MustCompile("@[A-z]+")

func isSwagComment(comment string) bool {
	return swagCommentExpression.MatchString(strings.ToLower(comment))
}

func isBlankComment(comment string) bool {
	lc := strings.TrimSpace(comment)
	return len(lc) == 0
}

// writeBack write to file
func writeBack(filepath string, src, old []byte) error {
	// make a temporary backup before overwriting original
	bakname, err := backupFile(filepath+".", old, 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, src, 0644)
	if err != nil {
		_ = os.Rename(bakname, filepath)
		return err
	}
	_ = os.Remove(bakname)
	return nil
}

const chmodSupported = runtime.GOOS != "windows"

// backupFile writes data to a new file named filename<number> with permissions perm,
// with <number randomly chosen such that the file name is unique. backupFile returns
// the chosen file name.
// copy from golang/cmd/gofmt
func backupFile(filename string, data []byte, perm os.FileMode) (string, error) {
	// create backup file
	f, err := ioutil.TempFile(filepath.Dir(filename), filepath.Base(filename))
	if err != nil {
		return "", err
	}
	if chmodSupported {
		_ = f.Chmod(perm)
	}

	// write data to backup file
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return f.Name(), err
}
