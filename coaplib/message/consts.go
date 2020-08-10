//go:generate stringer -type Type,Code,CodeCategory,OptionNumber,ValueFormat -linecomment -output consts_string.go

package message

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

const (
	KnownVer int = 1
)

type Type byte

const (
	CON Type = 0 // Confirmable
	NON Type = 1 //
	ACK Type = 2
	RST Type = 3
)

type TKL byte

const (
	CCReq          CodeCategory = 0
	CCAckOK        CodeCategory = 2
	CCClientBadAck CodeCategory = 4
	CCServerBadAck CodeCategory = 5

	CodeEmpty Code = 0 // 0.00 | empty

	MethodCodeGET    Code = 1 // 0.01 | get
	MethodCodePOST   Code = 2 // 0.02 | post
	MethodCodePUT    Code = 3 // 0.03 | put
	MethodCodeDELETE Code = 4 // 0.04 | delete

	ResponseCodeCreated                  Code = 65  // 2.01 | RFC7252
	ResponseCodeDeleted                  Code = 66  // 2.02 | RFC7252
	ResponseCodeValid                    Code = 67  // 2.03 | RFC7252
	ResponseCodeChanged                  Code = 68  // 2.04 | RFC7252
	ResponseCodeContent                  Code = 69  // 2.05 | RFC7252
	ResponseCodeBadRequest               Code = 128 // 4.00 | RFC7252
	ResponseCodeUnauthorized             Code = 129 // 4.01 | RFC7252
	ResponseCodeBadOption                Code = 130 // 4.02 | RFC7252
	ResponseCodeForbidden                Code = 131 // 4.03 | RFC7252
	ResponseCodeNotFound                 Code = 132 // 4.04 | RFC7252
	ResponseCodeMethodNotAllowed         Code = 133 // 4.05 | RFC7252
	ResponseCodeNotAcceptable            Code = 134 // 4.06 | RFC7252
	ResponseCodePreconditionFailed       Code = 140 // 4.12 | RFC7252
	ResponseCodeRequestEntityTooLarge    Code = 141 // 4.13 | RFC7252
	ResponseCodeUnsupportedContentFormat Code = 143 // 4.15 | RFC7252
	ResponseCodeInternalServerError      Code = 160 // 5.00 | RFC7252
	ResponseCodeNotImplemented           Code = 161 // 5.01 | RFC7252
	ResponseCodeBadGateway               Code = 162 // 5.02 | RFC7252
	ResponseCodeServiceUnavailable       Code = 163 // 5.03 | RFC7252
	ResponseCodeGatewayTimeout           Code = 164 // 5.04 | RFC7252
	ResponseCodeProxyingNotSupported     Code = 165 // 5.05 | RFC7252
)

type OptionNumber uint32

const (
	/*
	   OptionID identifies an option in a message.

	   +-----+----+---+---+---+----------------+--------+--------+---------+
	   | No. | C  | U | N | R | Name           | Format | Length | Default |
	   +-----+----+---+---+---+----------------+--------+--------+---------+
	   |   1 | x  |   |   | x | If-Match       | opaque | 0-8    | (none)  |
	   |   3 | x  | x | - |   | Uri-Host       | string | 1-255  | (see    |
	   |     |    |   |   |   |                |        |        | below)  |
	   |   4 |    |   |   | x | ETag           | opaque | 1-8    | (none)  |
	   |   5 | x  |   |   |   | If-None-Match  | empty  | 0      | (none)  |
	   |   7 | x  | x | - |   | Uri-Port       | uint   | 0-2    | (see    |
	   |     |    |   |   |   |                |        |        | below)  |
	   |   8 |    |   |   | x | Location-Path  | string | 0-255  | (none)  |
	   |  11 | x  | x | - | x | Uri-Path       | string | 0-255  | (none)  |
	   |  12 |    |   |   |   | Content-Format | uint   | 0-2    | (none)  |
	   |  14 |    | x | - |   | Max-Age        | uint   | 0-4    | 60      |
	   |  15 | x  | x | - | x | Uri-Query      | string | 0-255  | (none)  |
	   |  17 | x  |   |   |   | Accept         | uint   | 0-2    | (none)  |
	   |  20 |    |   |   | x | Location-Query | string | 0-255  | (none)  |
	   |  23 | x  | x | - | - | Block2         | uint   | 0-3    | (none)  |
	   |  27 | x  | x | - | - | Block1         | uint   | 0-3    | (none)  |
	   |  28 |    |   | x |   | Size2          | uint   | 0-4    | (none)  |
	   |  35 | x  | x | - |   | Proxy-Uri      | string | 1-1034 | (none)  |
	   |  39 | x  | x | - |   | Proxy-Scheme   | string | 1-255  | (none)  |
	   |  60 |    |   | x |   | Size1          | uint   | 0-4    | (none)  |
	   +-----+----+---+---+---+----------------+--------+--------+---------+
	   C=Critical, U=Unsafe, N=NoCacheKey, R=Repeatable
	*/

	OptionNumberReserved      OptionNumber = 0
	OptionNumberIfMatch       OptionNumber = 1
	OptionNumberURIHost       OptionNumber = 3
	OptionNumberETag          OptionNumber = 4
	OptionNumberIfNoneMatch   OptionNumber = 5
	OptionNumberObserve       OptionNumber = 6
	OptionNumberURIPort       OptionNumber = 7
	OptionNumberLocationPath  OptionNumber = 8
	OptionNumberURIPath       OptionNumber = 11
	OptionNumberContentFormat OptionNumber = 12
	OptionNumberMaxAge        OptionNumber = 14
	OptionNumberURIQuery      OptionNumber = 15
	OptionNumberAccept        OptionNumber = 17
	OptionNumberLocationQuery OptionNumber = 20
	OptionNumberBlock2        OptionNumber = 23
	OptionNumberBlock1        OptionNumber = 27
	OptionNumberSize2         OptionNumber = 28
	OptionNumberProxyURI      OptionNumber = 35
	OptionNumberProxyScheme   OptionNumber = 39
	OptionNumberSize1         OptionNumber = 60
	OptionNumberR128          OptionNumber = 128
	OptionNumberR132          OptionNumber = 132
	OptionNumberR136          OptionNumber = 136
	OptionNumberR140          OptionNumber = 140
	OptionNumberNoResponse    OptionNumber = 258

	//ContentFormatTextPlain   ContentFormat = 0
	//ContentFormatLinkFormat  ContentFormat = 40
	//ContentFormatXml         ContentFormat = 41
	//ContentFormatOctetStream ContentFormat = 42
	//ContentFormatExi         ContentFormat = 47
	//ContentFormatJson        ContentFormat = 50
)

// Option value format (RFC7252 section 3.2)
type ValueFormat uint8

const (
	ValueUnknown ValueFormat = iota
	ValueEmpty
	ValueOpaque
	ValueUint
	ValueString
)

type OptionDef struct {
	ValueFormat ValueFormat
	MinLen      int
	MaxLen      int
}

var CoapOptionDefs = map[OptionNumber]OptionDef{
	OptionNumberIfMatch:       {ValueFormat: ValueOpaque, MinLen: 0, MaxLen: 8},
	OptionNumberURIHost:       {ValueFormat: ValueString, MinLen: 1, MaxLen: 255},
	OptionNumberETag:          {ValueFormat: ValueOpaque, MinLen: 1, MaxLen: 8},
	OptionNumberIfNoneMatch:   {ValueFormat: ValueEmpty, MinLen: 0, MaxLen: 0},
	OptionNumberObserve:       {ValueFormat: ValueUint, MinLen: 0, MaxLen: 3},
	OptionNumberURIPort:       {ValueFormat: ValueUint, MinLen: 0, MaxLen: 2},
	OptionNumberLocationPath:  {ValueFormat: ValueString, MinLen: 0, MaxLen: 255},
	OptionNumberURIPath:       {ValueFormat: ValueString, MinLen: 0, MaxLen: 255},
	OptionNumberContentFormat: {ValueFormat: ValueUint, MinLen: 0, MaxLen: 2},
	OptionNumberMaxAge:        {ValueFormat: ValueUint, MinLen: 0, MaxLen: 4},
	OptionNumberURIQuery:      {ValueFormat: ValueString, MinLen: 0, MaxLen: 255},
	OptionNumberAccept:        {ValueFormat: ValueUint, MinLen: 0, MaxLen: 2},
	OptionNumberLocationQuery: {ValueFormat: ValueString, MinLen: 0, MaxLen: 255},
	OptionNumberBlock2:        {ValueFormat: ValueUint, MinLen: 0, MaxLen: 3},
	OptionNumberBlock1:        {ValueFormat: ValueUint, MinLen: 0, MaxLen: 3},
	OptionNumberSize2:         {ValueFormat: ValueUint, MinLen: 0, MaxLen: 4},
	OptionNumberProxyURI:      {ValueFormat: ValueString, MinLen: 1, MaxLen: 1034},
	OptionNumberProxyScheme:   {ValueFormat: ValueString, MinLen: 1, MaxLen: 255},
	OptionNumberSize1:         {ValueFormat: ValueUint, MinLen: 0, MaxLen: 4},
	OptionNumberNoResponse:    {ValueFormat: ValueUint, MinLen: 0, MaxLen: 1},
}

// type ContentFormat byte

// MediaType specifies the content format of a message.
type MediaType uint16

// Content formats.
const (
	MediaTypeUndefined MediaType = 0xffff
	TextPlain          MediaType = 0     // text/plain;charset=utf-8
	AppCoseEncrypt0    MediaType = 16    // application/cose; cose-type="cose-encrypt0" (RFC 8152)
	AppCoseMac0        MediaType = 17    // application/cose; cose-type="cose-mac0" (RFC 8152)
	AppCoseSign1       MediaType = 18    // application/cose; cose-type="cose-sign1" (RFC 8152)
	AppLinkFormat      MediaType = 40    // application/link-format
	AppXML             MediaType = 41    // application/xml
	AppOctets          MediaType = 42    // application/octet-stream
	AppExi             MediaType = 47    // application/exi
	AppJSON            MediaType = 50    // application/json
	AppJSONPatch       MediaType = 51    // application/json-patch+json (RFC6902)
	AppJSONMergePatch  MediaType = 52    // application/merge-patch+json (RFC7396)
	AppCBOR            MediaType = 60    // application/cbor (RFC 7049)
	AppCWT             MediaType = 61    // application/cwt
	AppCoseEncrypt     MediaType = 96    // application/cose; cose-type="cose-encrypt" (RFC 8152)
	AppCoseMac         MediaType = 97    // application/cose; cose-type="cose-mac" (RFC 8152)
	AppCoseSign        MediaType = 98    // application/cose; cose-type="cose-sign" (RFC 8152)
	AppCoseKey         MediaType = 101   // application/cose-key (RFC 8152)
	AppCoseKeySet      MediaType = 102   // application/cose-key-set (RFC 8152)
	AppCoapGroup       MediaType = 256   // coap-group+json (RFC 7390)
	AppOcfCbor         MediaType = 10000 // application/vnd.ocf+cbor
	AppLwm2mTLV        MediaType = 11542 // application/vnd.oma.lwm2m+tlv
	AppLwm2mJSON       MediaType = 11543 // application/vnd.oma.lwm2m+json
)

var mediaTypeToString = map[MediaType]string{
	MediaTypeUndefined: "??",
	TextPlain:          "text/plain;charset=utf-8",
	AppCoseEncrypt0:    "application/cose; cose-type=\"cose-encrypt0\" (RFC 8152)",
	AppCoseMac0:        "application/cose; cose-type=\"cose-mac0\" (RFC 8152)",
	AppCoseSign1:       "application/cose; cose-type=\"cose-sign1\" (RFC 8152)",
	AppLinkFormat:      "application/link-format",
	AppXML:             "application/xml",
	AppOctets:          "application/octet-stream",
	AppExi:             "application/exi",
	AppJSON:            "application/json",
	AppJSONPatch:       "application/json-patch+json (RFC6902)",
	AppJSONMergePatch:  "application/merge-patch+json (RFC7396)",
	AppCBOR:            "application/cbor (RFC 7049)",
	AppCWT:             "application/cwt",
	AppCoseEncrypt:     "application/cose; cose-type=\"cose-encrypt\" (RFC 8152)",
	AppCoseMac:         "application/cose; cose-type=\"cose-mac\" (RFC 8152)",
	AppCoseSign:        "application/cose; cose-type=\"cose-sign\" (RFC 8152)",
	AppCoseKey:         "application/cose-key (RFC 8152)",
	AppCoseKeySet:      "application/cose-key-set (RFC 8152)",
	AppCoapGroup:       "coap-group+json (RFC 7390)",
	AppOcfCbor:         "application/vnd.ocf+cbor",
	AppLwm2mTLV:        "application/vnd.oma.lwm2m+tlv",
	AppLwm2mJSON:       "application/vnd.oma.lwm2m+json",
}

func (c MediaType) String() string {
	str, ok := mediaTypeToString[c]
	if !ok {
		return "MediaType(" + strconv.FormatInt(int64(c), 10) + ")"
	}
	return str
}

func ToMediaType(v string) (MediaType, error) {
	for key, val := range mediaTypeToString {
		if val == v {
			return key, nil
		}
	}
	return 0, fmt.Errorf("not found")
}

//

func GetOptionNumberMap() map[OptionNumber]string {
	return _OptionNumber_map
}

func init() {
	optionNumbersOnce.Do(func() {

		var keys []int
		for k := range _OptionNumber_map {
			keys = append(keys, int(k))
		}
		sort.Ints(keys)

		for _, k := range keys {
			OptionNumberSortedKeys = append(OptionNumberSortedKeys, OptionNumber(k))
		}
	})
}

var optionNumbersOnce sync.Once
var OptionNumberSortedKeys []OptionNumber
