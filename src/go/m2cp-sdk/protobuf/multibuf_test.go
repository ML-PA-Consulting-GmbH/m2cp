package protobuf

import (
	"encoding/base64"
)

const data = "EYsBCAAIAggICBIIIAgyCEgIYgiAAQiiAQjIAQjyAQigAgjSAgiIAwjCAwiABAjCBAiIBQjSBRAAEAEQCBAbEEAQfRgAGAQYEBgkGEAYZBiQARjEARiAAhjEAhiQAxjkAxjABBikBRiQBhiEBxiACBiECRiQChikCyCPTioQAQIDBAUGBwgJCgsMDQ4PEBOYAQjWurSpBhARGAIiiwEIAAgCCAgIEgggCDIISAhiCIABCKIBCMgBCPIBCKACCNICCIgDCMIDCIAECMIECIgFCNIFEAAQARAIEBsQQBB9GAAYBBgQGCQYQBhkGJABGMQBGIACGMQCGJADGOQDGMAEGKQFGJAGGIQHGIAIGIQJGJAKGKQLII9OKhABAgMEBQYHCAkKCwwNDg8QFgoIz7HqwbMxEOQZFQ0Iz7HqwbMxELgXGJADCxoIz7HqwbMxEIDVpQYYoJfEBiAFKNAPMBQ4CBIlCM+x6sGzMRACGhp0aGlzIGlzIGEgdGVzdCBsb2cgbWVzc2FnZRQOCM+x6sGzMRDQDxgKIDs="

func (t *TestSuite) TestParseConcatenatedProtobuf() {
	// decode base64
	decoded, _ := base64.StdEncoding.DecodeString(data)
	parsed, err := UnpackWrappedProtobufs(decoded)
	t.NoError(err)
	t.Len(parsed, 7)
	t.Equal(uint(17), parsed[0].Oid)
	t.Len(parsed[0].Protobuf, 139)
	t.Equal(uint(19), parsed[1].Oid)
	t.Len(parsed[1].Protobuf, 152)
	t.Equal(uint(22), parsed[2].Oid)
	t.Len(parsed[2].Protobuf, 10)
	t.Equal(uint(21), parsed[3].Oid)
	t.Len(parsed[3].Protobuf, 13)
	t.Equal(uint(11), parsed[4].Oid)
	t.Len(parsed[4].Protobuf, 26)
	t.Equal(uint(18), parsed[5].Oid)
	t.Len(parsed[5].Protobuf, 37)
	t.Equal(uint(20), parsed[6].Oid)
	t.Len(parsed[6].Protobuf, 14)
}

const dataMulti = "EVYKFAAAAAAAAAAAAAAAAAAAAAAAAAAAEgYAAAAAAAAaKP4//j/+P/4//j/+P/4//j/+P/4//j/+P/4//j/+P/4//j/+P/4//j8gACoE7iIAADDDgrDwBRFcChakA9YCAgAAAAAAAAAAAAAAAAAAAAAAEgoAAAAAAICAgIACGijyIYAm/j/+P/4//j/+P/4//j/+P/4//j/+P/4//j/+P/4//j/+P/4/IAAqBPQiAAAww4Kw8AURVgoUAAAAAAAAAAAAAAAAAAAAAAAAAAASBgAAAAAAABoo/j/+P/4//j/+P/4//j/+P/4//j/+P/4//j/+P/4//j/+P/4//j/+PyAAKgTwIgAAMMOCsPAFEVYKFAAAAAAAAAAAAAAAAAAAAAAAAAAAEgYAAAAAAAAaKP4//j/+P/4//j/+P/4//j/+P/4//j/+P/4//j/+P/4//j/+P/4//j8gACoE7SIAADDDgrDwBQ=="

func (t *TestSuite) TestParseConcatenatedProtobufMulti() {
	// decode base64
	decoded, _ := base64.StdEncoding.DecodeString(dataMulti)
	parsed, err := UnpackWrappedProtobufs(decoded)
	t.NoError(err)
	t.NotNil(parsed)
	t.Equal(4, len(parsed))
}

func (t *TestSuite) TestParseConcatenatedProtobufTruncated() {
	// decode base64
	decoded, _ := base64.StdEncoding.DecodeString(data)

	var err error
	var parsed []WrappedProtobuf

	parsed, err = UnpackWrappedProtobufs(decoded)
	t.NoError(err)
	t.NotNil(parsed)
	t.Equal(7, len(parsed))
}
