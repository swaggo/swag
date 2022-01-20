package swag

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
)

func formaterTimeMachine() {
	// reset format_test to format_src
	err := copy.Copy("./testdata/format_src", "./testdata/format_test")
	if err != nil {
		panic(err)
	}
}

const (
	SearchDir = "./testdata/format_test"
	Excludes  = "./testdata/format_test/web"
	MainFile  = "main.go"
)

func TestFormater_FormatAPI(t *testing.T) {
	t.Run("Format Test", func(t *testing.T) {
		formaterTimeMachine()
		formater := NewFormater()
		err := formater.FormatAPI(SearchDir, Excludes, MainFile)
		assert.NoError(t, err)
		parsedFile, err := ioutil.ReadFile("./testdata/format_test/api/api.go")
		assert.NoError(t, err)
		apiFile, err := ioutil.ReadFile("./testdata/format_dst/api/api.go")
		assert.NoError(t, err)
		assert.Equal(t, parsedFile, apiFile)

		parsedMainFile, err := ioutil.ReadFile("./testdata/format_test/main.go")
		assert.NoError(t, err)
		mainFile, err := ioutil.ReadFile("./testdata/format_dst/main.go")
		assert.NoError(t, err)
		assert.Equal(t, parsedMainFile, mainFile)
		formaterTimeMachine()
	})

	t.Run("TestWrongSearchDir", func(t *testing.T) {
		t.Parallel()
		formater := NewFormater()
		err := formater.FormatAPI("/dir_not_have", "", "")
		assert.Error(t, err)
	})

	t.Run("TestWithMonkeyFilepathAbs", func(t *testing.T) {
		formater := NewFormater()
		errFilePath := fmt.Errorf("file path error ")

		patches := gomonkey.ApplyFunc(filepath.Abs, func(_ string) (string, error) {
			return "", errFilePath
		})
		defer patches.Reset()

		err := formater.FormatAPI(SearchDir, Excludes, MainFile)
		assert.Equal(t, err, errFilePath)
		formaterTimeMachine()
	})

	t.Run("TestWithMonkeyFormatMain", func(t *testing.T) {
		formater := NewFormater()

		var s *Formater
		errFormatMain := fmt.Errorf("main format error ")
		patches := gomonkey.ApplyMethod(reflect.TypeOf(s), "FormatMain", func(_ *Formater, _ string) error {
			return errFormatMain
		})
		defer patches.Reset()

		err := formater.FormatAPI(SearchDir, Excludes, MainFile)
		assert.Equal(t, err, errFormatMain)
		formaterTimeMachine()
	})

	t.Run("TestWithMonkeyFormatFile", func(t *testing.T) {
		formater := NewFormater()

		var s *Formater
		errFormatFile := fmt.Errorf("file format error ")
		patches := gomonkey.ApplyMethod(reflect.TypeOf(s), "FormatFile", func(_ *Formater, _ string) error {
			return errFormatFile
		})
		defer patches.Reset()

		err := formater.FormatAPI(SearchDir, Excludes, MainFile)
		assert.Equal(t, err, fmt.Errorf("ParseFile error:%s", errFormatFile))
		formaterTimeMachine()
	})

}

func TestFormater_FormatMain(t *testing.T) {
	t.Run("TestWrongMainPath", func(t *testing.T) {
		t.Parallel()
		formater := NewFormater()
		err := formater.FormatMain("/dir_not_have/main.go")
		assert.Error(t, err)
	})
}

func TestFormater_FormatFile(t *testing.T) {
	t.Run("TestWrongFilePath", func(t *testing.T) {
		t.Parallel()
		formater := NewFormater()
		err := formater.FormatFile("/dir_not_have/api.go")
		assert.Error(t, err)
	})
}

func Test_writeFormatedComments(t *testing.T) {
	t.Run("TestWrongPath", func(t *testing.T) {
		t.Parallel()
		var (
			formatedComments = bytes.Buffer{}
			// CommentCache
			oldCommentsMap = make(map[string]string)
		)
		err := writeFormatedComments("/wrong_path", formatedComments, oldCommentsMap)
		assert.Error(t, err)
	})
}

func TestFormater_visit(t *testing.T) {
	formater := NewFormater()

	err := formater.visit("./testdata/test_test.go", &mockFS{}, nil)
	assert.NoError(t, err)
	err = formater.visit("/testdata/api.md", &mockFS{}, nil)
	assert.NoError(t, err)
	formater.mainFile = "main.go"
	err = formater.visit("/testdata/main.go", &mockFS{}, nil)
	assert.NoError(t, err)
}

func Test_isBlankComment(t *testing.T) {
	type args struct {
		comment string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				comment: " ",
			},
			want: true,
		},
		{
			name: "test2",
			args: args{
				comment: " A",
			},
			want: false,
		},
		{
			name: "test3",
			args: args{
				comment: " \t",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBlankComment(tt.args.comment); got != tt.want {
				t.Errorf("isBlankComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isSwagComment(t *testing.T) {
	type args struct {
		comment string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				comment: "@Param some_id ",
			},
			want: true,
		},
		{
			name: "test2",
			args: args{
				comment: "@ ",
			},
			want: false,
		},
		{
			name: "test3",
			args: args{
				comment: "@Success {object} ",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSwagComment(tt.args.comment); got != tt.want {
				t.Errorf("isSwagComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_replaceRange(t *testing.T) {
	type args struct {
		s     []byte
		start int
		end   int
		new   byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "test_replaceSuccess",
			args: args{
				s:     []byte("// @ID  get-ids"),
				start: 6,
				end:   8,
				new:   '\t',
			},
			want: []byte("// @ID\tget-ids"),
		},
		{
			name: "test1_replaceFail",
			args: args{
				s:     []byte("// @ID  A pet"),
				start: 6,
				end:   8,
				new:   '\t',
			},
			want: []byte("// @ID\tA pet"),
		},
		{
			name: "test1_replaceFail2",
			args: args{
				s:     []byte("// @ID  "),
				start: 6,
				end:   12,
				new:   '\t',
			},
			want: []byte("// @ID\t"),
		},
		{
			name: "test1_replaceFail3",
			args: args{
				s:     []byte("// @ID  "),
				start: 2,
				end:   1,
				new:   '\t',
			},
			want: []byte("// @ID  "),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replaceRange(tt.args.s, tt.args.start, tt.args.end, tt.args.new); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("replaceRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_separatorFinder(t *testing.T) {
	type args struct {
		comment string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				comment: `// @Param   some_id  query int  "some id  data" Enums(1, 2, 3)`,
			},
			want: `// @Param|some_id|query|int|"some id  data"|Enums(1, 2, 3)`,
		},
		{
			name: "test2",
			args: args{
				comment: `// @Summary   A pet store. `,
			},
			want: `// @Summary|A pet store. `,
		},
		{
			name: "test3",
			args: args{
				comment: `// @Summary    `,
			},
			want: `// @Summary    `,
		},
		{
			name: "test4",
			args: args{
				comment: `// @Failure      400       {object}  web.APIError{data=web.D ,data2=web.D2}  "We need ID!!"`,
			},
			want: `// @Failure|400|{object}|web.APIError{data=web.D ,data2=web.D2}|"We need ID!!"`,
		},
		{
			name: "test5",
			args: args{
				comment: `// `,
			},
			want: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := separatorFinder(tt.args.comment, '|')
			assert.Equal(t, got, tt.want)
		})
	}
}

func Test_writeBack(t *testing.T) {
	t.Run("Test", func(t *testing.T) {
		testFile, err := backupFile("test.go", []byte("package main \n"), 0644)
		assert.NoError(t, err)
		defer func() {
			_ = os.Remove(testFile)
		}()

		testBytes, err := ioutil.ReadFile(testFile)
		assert.NoError(t, err)
		newBytes := append(testBytes, []byte("import ()")...)

		err = writeBack(testFile, newBytes, testBytes)
		assert.NoError(t, err)

		newTestBytes, err := ioutil.ReadFile(testFile)
		assert.NoError(t, err)

		assert.Equal(t, newTestBytes, newBytes)
	})

	t.Run("TestWrongPathError", func(t *testing.T) {
		testFile, err := backupFile("test.go", []byte("package main \n"), 0644)
		assert.NoError(t, err)
		defer func() {
			_ = os.Remove(testFile)
		}()

		testBytes, err := ioutil.ReadFile(testFile)
		assert.NoError(t, err)
		newBytes := append(testBytes, []byte("import ()")...)

		err = writeBack("/not_found_file_path", testBytes, newBytes)
		assert.Error(t, err)
	})

	t.Run("TestWrongFile", func(t *testing.T) {
		testFile, err := backupFile("test.go", []byte("package main \n"), 0644)
		assert.NoError(t, err)
		defer func() {
			_ = os.Remove(testFile)
		}()

		testBytes, err := ioutil.ReadFile(testFile)
		assert.NoError(t, err)
		newBytes := append(testBytes, []byte("import ()")...)

		err = writeBack("", testBytes, newBytes)
		assert.Error(t, err)
	})
}
