package swag

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestFormater_FormatAPI(t *testing.T) {
	t.Parallel()

	formater := NewFormater()

}

func TestFormater_FormatFile(t *testing.T) {
	type fields struct {
		debug    Debugger
		excludes map[string]bool
	}
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formater := &Formater{
				debug:    tt.fields.debug,
				excludes: tt.fields.excludes,
			}
			if err := formater.FormatFile(tt.args.filepath); (err != nil) != tt.wantErr {
				t.Errorf("FormatFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormater_formatMultiSearchDir(t *testing.T) {
	type fields struct {
		debug    Debugger
		excludes map[string]bool
	}
	type args struct {
		searchDirs []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Formater{
				debug:    tt.fields.debug,
				excludes: tt.fields.excludes,
			}
			if err := f.formatMultiSearchDir(tt.args.searchDirs); (err != nil) != tt.wantErr {
				t.Errorf("formatMultiSearchDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormater_skip(t *testing.T) {
	type fields struct {
		debug    Debugger
		excludes map[string]bool
	}
	type args struct {
		path     string
		fileInfo os.FileInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Formater{
				debug:    tt.fields.debug,
				excludes: tt.fields.excludes,
			}
			if err := f.skip(tt.args.path, tt.args.fileInfo); (err != nil) != tt.wantErr {
				t.Errorf("skip() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormater_visit(t *testing.T) {
	type fields struct {
		debug    Debugger
		excludes map[string]bool
	}
	type args struct {
		path     string
		fileInfo os.FileInfo
		err      error
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Formater{
				debug:    tt.fields.debug,
				excludes: tt.fields.excludes,
			}
			if err := f.visit(tt.args.path, tt.args.fileInfo, tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("visit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewFormater(t *testing.T) {
	tests := []struct {
		name string
		want *Formater
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFormater(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFormater() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_backupFile(t *testing.T) {
	type args struct {
		filename string
		data     []byte
		perm     os.FileMode
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := backupFile(tt.args.filename, tt.args.data, tt.args.perm)
			if (err != nil) != tt.wantErr {
				t.Errorf("backupFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("backupFile() got = %v, want %v", got, tt.want)
			}
		})
	}
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
				s:     []byte("// @ID  "),
				start: 6,
				end:   8,
				new:   '\t',
			},
			want: []byte("// @ID\t"),
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := separatorFinder(tt.args.comment); got != tt.want {
				t.Errorf("separatorFinder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_writeBack(t *testing.T) {
	type args struct {
		filepath string
		src      []byte
		old      []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := writeBack(tt.args.filepath, tt.args.src, tt.args.old); (err != nil) != tt.wantErr {
				t.Errorf("writeBack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
