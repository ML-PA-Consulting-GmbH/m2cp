package protobuf

import (
	"errors"
	"fmt"
	"m2cp"
	"maps"
)

var ErrNoParserForOid = errors.New("no parser for oid")

var parsers map[uint]m2cp.ProtobufOidParser

func GetOidParser(oid uint) (m2cp.ProtobufOidParser, error) {
	if parsers == nil {
		return nil, fmt.Errorf("%w %d", ErrNoParserForOid, oid)
	}
	if parser, ok := parsers[oid]; ok {
		return parser, nil
	}
	return nil, fmt.Errorf("%w %d", ErrNoParserForOid, oid)
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
	maps.Copy(copyOf, parsers)
	return copyOf
}
