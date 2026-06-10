package m2cp_new

import (
	"github.com/stretchr/testify/assert"
	"m2cp"
	"testing"
)

func TestProtobufParserService(t *testing.T) {
	ctp := ContextPlus()
	var err error
	var parser m2cp.ProtobufParserService
	parser, err = ProtobufParserService(ctp, nil, nil)
	assert.Error(t, err)
	assert.Nil(t, parser)
	ctp.Cancel()

}
