package parse

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestOperation_ParseAcceptComment(t *testing.T) {
	operation:= new(Operation)
	operation.ParseAcceptComment("json,xml,plain,html,mpfd")
	expected:=[]string{"application/json", "text/xml", "text/plain", "text/html", "multipart/form-data"}
	assert.Equal(t,expected,operation.Consumes)
}

func TestOperation_ParseProduceComment(t *testing.T) {
	operation:= new(Operation)
	operation.ParseProduceComment("json,xml,plain,html,mpfd")
	expected:=[]string{"application/json", "text/xml", "text/plain", "text/html", "multipart/form-data"}
	assert.Equal(t,expected,operation.Produces)
}
