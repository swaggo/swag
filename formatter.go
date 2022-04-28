package swag

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
)

const splitTag = "&*"

// Check of @Param @Success @Failure @Response @Header
var specialTagForSplit = map[string]bool{
	paramAttr:    true,
	successAttr:  true,
	failureAttr:  true,
	responseAttr: true,
	headerAttr:   true,
}

var skipChar = map[byte]byte{
	'"': '"',
	'(': ')',
	'{': '}',
	'[': ']',
}

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

// Format formats swag comments in contents. It uses fileName to report errors
// that happen during parsing of contents.
func (f *Formatter) Format(fileName string, contents []byte) ([]byte, error) {
	fileSet := token.NewFileSet()
	ast, err := goparser.ParseFile(fileSet, fileName, contents, goparser.ParseComments)
	if err != nil {
		return nil, err
	}
	formattedComments := bytes.Buffer{}
	oldComments := map[string]string{}

	if ast.Comments != nil {
		for _, comment := range ast.Comments {
			formatFuncDoc(comment.List, &formattedComments, oldComments)
		}
	}
	return formatComments(fileName, contents, formattedComments.Bytes(), oldComments), nil
}

func formatComments(fileName string, contents []byte, formattedComments []byte, oldComments map[string]string) []byte {
	for _, comment := range bytes.Split(formattedComments, []byte("\n")) {
		splits := bytes.SplitN(comment, []byte(splitTag), 2)
		if len(splits) == 2 {
			hash, line := splits[0], splits[1]
			contents = bytes.Replace(contents, []byte(oldComments[string(hash)]), line, 1)
		}
	}
	return contents
}

func formatFuncDoc(commentList []*ast.Comment, formattedComments io.Writer, oldCommentsMap map[string]string) {
	w := tabwriter.NewWriter(formattedComments, 0, 0, 2, ' ', 0)

	for _, comment := range commentList {
		text := comment.Text
		if attr, body, found := swagComment(text); found {
			cmd5 := fmt.Sprintf("%x", md5.Sum([]byte(text)))
			oldCommentsMap[cmd5] = text

			formatted := "// " + attr
			if body != "" {
				formatted += "\t" + splitComment2(attr, body)
			}
			// md5 + splitTag + srcCommentLine
			// eg. xxx&*@Description get struct array
			_, _ = fmt.Fprintln(w, cmd5+splitTag+formatted)
		}
	}
	// format by tabwriter
	_ = w.Flush()
}

func splitComment2(attr, body string) string {
	if specialTagForSplit[strings.ToLower(attr)] {
		for i := 0; i < len(body); i++ {
			if skipEnd, ok := skipChar[body[i]]; ok {
				if skipLen := strings.IndexByte(body[i+1:], skipEnd); skipLen > 0 {
					i += skipLen
				}
			} else if body[i] == ' ' {
				j := i
				for ; j < len(body) && body[j] == ' '; j++ {
				}
				body = replaceRange(body, i, j, "\t")
			}
		}
	}
	return body
}

func replaceRange(s string, start, end int, new string) string {
	return s[:start] + new + s[end:]
}

var swagCommentLineExpression = regexp.MustCompile(`^\/\/\s+(@[\S.]+)\s*(.*)`)

func swagComment(comment string) (string, string, bool) {
	matches := swagCommentLineExpression.FindStringSubmatch(comment)
	if matches == nil {
		return "", "", false
	}
	return matches[1], matches[2], true
}
