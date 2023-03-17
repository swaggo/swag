package swag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpec_InstanceName(t *testing.T) {
	type fields struct {
		Version          string
		Host             string
		BasePath         string
		Schemes          []string
		Title            string
		Description      string
		InfoInstanceName string
		SwaggerTemplate  string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "TestInstanceNameCorrect",
			fields: fields{
				Version:          "1.0",
				Host:             "localhost:8080",
				BasePath:         "/",
				InfoInstanceName: "TestInstanceName1",
			},
			want: "TestInstanceName1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := Spec{
				Version:          tt.fields.Version,
				Host:             tt.fields.Host,
				BasePath:         tt.fields.BasePath,
				Schemes:          tt.fields.Schemes,
				Title:            tt.fields.Title,
				Description:      tt.fields.Description,
				InfoInstanceName: tt.fields.InfoInstanceName,
				SwaggerTemplate:  tt.fields.SwaggerTemplate,
			}

			assert.Equal(t, tt.want, doc.InstanceName())
		})
	}
}

func TestSpec_ReadDoc(t *testing.T) {
	type fields struct {
		Version          string
		Host             string
		BasePath         string
		Schemes          []string
		Title            string
		Description      string
		InfoInstanceName string
		SwaggerTemplate  string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "TestReadDocCorrect",
			fields: fields{
				Version:          "1.0",
				Host:             "localhost:8080",
				BasePath:         "/",
				InfoInstanceName: "TestInstanceName",
				SwaggerTemplate: `{
			"swagger": "2.0",
			"info": {
				"description": "{{escape .Description}}",
				"title": "{{.Title}}",
				"version": "{{.Version}}"
			},
			"host": "{{.Host}}",
			"basePath": "{{.BasePath}}",
		}`,
			},
			want: "{" +
				"\n\t\t\t\"swagger\": \"2.0\"," +
				"\n\t\t\t\"info\": {" +
				"\n\t\t\t\t\"description\": \"\",\n\t\t\t\t\"" +
				"title\": \"\"," +
				"\n\t\t\t\t\"version\": \"1.0\"" +
				"\n\t\t\t}," +
				"\n\t\t\t\"host\": \"localhost:8080\"," +
				"\n\t\t\t\"basePath\": \"/\"," +
				"\n\t\t}",
		},
		{
			name: "TestReadDocMarshalTrigger",
			fields: fields{
				Version:          "1.0",
				Host:             "localhost:8080",
				BasePath:         "/",
				InfoInstanceName: "TestInstanceName",
				SwaggerTemplate:  "{{ marshal .Version }}",
			},
			want: "\"1.0\"",
		},
		{
			name: "TestReadDocParseError",
			fields: fields{
				Version:          "1.0",
				Host:             "localhost:8080",
				BasePath:         "/",
				InfoInstanceName: "TestInstanceName",
				SwaggerTemplate:  "{{ ..Version }}",
			},
			want: "{{ ..Version }}",
		},
		{
			name: "TestReadDocExecuteError",
			fields: fields{
				Version:          "1.0",
				Host:             "localhost:8080",
				BasePath:         "/",
				InfoInstanceName: "TestInstanceName",
				SwaggerTemplate:  "{{ .Schemesa }}",
			},
			want: "{{ .Schemesa }}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := Spec{
				Version:          tt.fields.Version,
				Host:             tt.fields.Host,
				BasePath:         tt.fields.BasePath,
				Schemes:          tt.fields.Schemes,
				Title:            tt.fields.Title,
				Description:      tt.fields.Description,
				InfoInstanceName: tt.fields.InfoInstanceName,
				SwaggerTemplate:  tt.fields.SwaggerTemplate,
			}

			assert.Equal(t, tt.want, doc.ReadDoc())
		})
	}
}
