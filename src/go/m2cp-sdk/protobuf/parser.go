package protobuf

import (
	"fmt"
	"m2cp"
)

var parsers map[uint]m2cp.ProtobufOidParser

func GetOidParser(oid uint) (m2cp.ProtobufOidParser, error) {
	if parsers == nil {
		return nil, fmt.Errorf("no parser for oid %d", oid)
	}
	if parser, ok := parsers[oid]; ok {
		return parser, nil
	}
	return nil, fmt.Errorf("no parser for oid %d", oid)
}

func RegisterParser(oid uint, parser m2cp.ProtobufOidParser) {
	if parsers == nil {
		parsers = make(map[uint]m2cp.ProtobufOidParser)
	}
	parsers[oid] = parser
}

func GetOidParsers() map[uint]m2cp.ProtobufOidParser {
	if parsers == nil {
		return map[uint]m2cp.ProtobufOidParser{}
	}
	copyOf := make(map[uint]m2cp.ProtobufOidParser, len(parsers))
	for oid, parser := range parsers {
		copyOf[oid] = parser
	}
	return copyOf
}
