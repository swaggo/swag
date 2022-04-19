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
	"strings"
	"text/tabwriter"
)

const splitTag = "&*"

// Formatter implements a formatter for Go source files.
type Formatter struct {
	// debugging output goes here
	debug Debugger
}

// NewFormatter create a new formatter instance.
func NewFormatter() *Formatter {
	formatter := &Formatter{
		debug: log.New(os.Stdout, "", log.LstdFlags),
	}
	return formatter
}

// Format swag comments in given file.
func (f *Formatter) Format(filepath string) error {
	fileSet := token.NewFileSet()
	astFile, err := goparser.ParseFile(fileSet, filepath, nil, goparser.ParseComments)
	if err != nil {
		return err
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

	return writeFormattedComments(filepath, formatedComments, oldCommentsMap)
}

func writeFormattedComments(filepath string, formatedComments bytes.Buffer, oldCommentsMap map[string]string) error {
	// Replace the file
	// Read the file
	srcBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("cannot open file, err: %w path : %s ", err, filepath)
	}

	replaceSrc, newComments := string(srcBytes), strings.Split(formatedComments.String(), "\n")

	for _, e := range newComments {
		commentSplit := strings.Split(e, splitTag)
		if len(commentSplit) == 2 {
			commentHash, commentContent := commentSplit[0], commentSplit[1]

			if !isBlankComment(commentContent) {
				replaceSrc = strings.Replace(replaceSrc, oldCommentsMap[commentHash], commentContent, 1)
			}
		}
	}
	return writeBack(filepath, []byte(replaceSrc))
}

func formatFuncDoc(commentList []*ast.Comment, formattedComments io.Writer, oldCommentsMap map[string]string) {
	tabWriter := tabwriter.NewWriter(formattedComments, 0, 0, 2, ' ', 0)

	for _, comment := range commentList {
		commentLine := comment.Text
		if isSwagComment(commentLine) || isBlankComment(commentLine) {
			cmd5 := fmt.Sprintf("%x", md5.Sum([]byte(commentLine)))

			// Find the separator and replace to \t
			c := separatorFinder(commentLine, '\t')
			oldCommentsMap[cmd5] = commentLine

			// md5 + splitTag + srcCommentLine
			// eg. xxx&*@Description get struct array
			_, _ = fmt.Fprintln(tabWriter, cmd5+splitTag+c)
		}
	}
	// format by tabWriter
	_ = tabWriter.Flush()
}

func separatorFinder(comment string, replacer byte) string {
	commentBytes, commentLine := []byte(comment), strings.TrimSpace(strings.TrimLeft(comment, "/"))

	if len(commentLine) == 0 {
		return ""
	}

	attribute := strings.Fields(commentLine)[0]
	attrLen := strings.Index(comment, attribute) + len(attribute)
	attribute = strings.ToLower(attribute)

	var (
		length = attrLen

		// Check of @Param @Success @Failure @Response @Header.
		specialTagForSplit = map[string]byte{
			paramAttr:    1,
			successAttr:  1,
			failureAttr:  1,
			responseAttr: 1,
			headerAttr:   1,
		}
	)

	_, ok := specialTagForSplit[attribute]
	if ok {
		return splitSpecialTags(commentBytes, length, replacer)
	}

	for length < len(commentBytes) && commentBytes[length] == ' ' {
		length++
	}

	if length >= len(commentBytes) {
		return comment
	}

	commentBytes = replaceRange(commentBytes, attrLen, length, replacer)

	return string(commentBytes)
}

func splitSpecialTags(commentBytes []byte, length int, rp byte) string {
	var (
		skipFlag bool
		skipChar = map[byte]byte{
			'"': 1,
			'(': 1,
			'{': 1,
			'[': 1,
		}

		skipCharEnd = map[byte]byte{
			'"': 1,
			')': 1,
			'}': 1,
			']': 1,
		}
	)

	for ; length < len(commentBytes); length++ {
		if !skipFlag && commentBytes[length] == ' ' {
			j := length
			for j < len(commentBytes) && commentBytes[j] == ' ' {
				j++
			}

			commentBytes = replaceRange(commentBytes, length, j, rp)
		}

		_, found := skipChar[commentBytes[length]]
		if found && !skipFlag {
			skipFlag = true

			continue
		}

		_, found = skipCharEnd[commentBytes[length]]
		if found && skipFlag {
			skipFlag = false
		}
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
	return len(strings.TrimSpace(comment)) == 0
}

func writeBack(filename string, src []byte) error {
	f, err := ioutil.TempFile(filepath.Dir(filename), filepath.Base(filename))
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(src); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), filename); err != nil {
		return err
	}
	return nil
}
