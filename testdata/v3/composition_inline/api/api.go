package api

import (
	"net/http"
)

// BaseMetadata contains common metadata fields
type BaseMetadata struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Status      string `json:"status" yaml:"status"`
}

// ServerMetadata embeds BaseMetadata with yaml inline tag
type ServerMetadata struct {
	BaseMetadata `yaml:",inline"`
	Image        string `json:"image" yaml:"image"`
}

// ServerMetadataJSONInline embeds BaseMetadata with both json and yaml inline tags
type ServerMetadataJSONInline struct {
	BaseMetadata `json:",inline" yaml:",inline"`
	Image        string `json:"image" yaml:"image"`
}

// ServerMetadataWithIgnored has a field that should be ignored via json:"-"
type ServerMetadataWithIgnored struct {
	BaseMetadata `yaml:",inline"`
	Image        string `json:"image" yaml:"image"`
	Secret       string `json:"-"` // This field should be skipped
}

// @Description get ServerMetadata
// @ID get-server-metadata
// @Accept json
// @Produce json
// @Success 200 {object} api.ServerMetadata
// @Router /testapi/get-server-metadata [get]
func GetServerMetadata(w http.ResponseWriter, r *http.Request) {
	var _ = ServerMetadata{}
}

// @Description get ServerMetadataJSONInline
// @ID get-server-metadata-json-inline
// @Accept json
// @Produce json
// @Success 200 {object} api.ServerMetadataJSONInline
// @Router /testapi/get-server-metadata-json-inline [get]
func GetServerMetadataJSONInline(w http.ResponseWriter, r *http.Request) {
	var _ = ServerMetadataJSONInline{}
}

// @Description get ServerMetadataWithIgnored
// @ID get-server-metadata-with-ignored
// @Accept json
// @Produce json
// @Success 200 {object} api.ServerMetadataWithIgnored
// @Router /testapi/get-server-metadata-with-ignored [get]
func GetServerMetadataWithIgnored(w http.ResponseWriter, r *http.Request) {
	var _ = ServerMetadataWithIgnored{}
}
