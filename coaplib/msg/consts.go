//go:generate stringer -type Type,Code,CodeCategory,OptionNumber,ContentFormat -linecomment -output consts_string.go

package msg

type Type uint8
type TKL uint8
type OptionNumber uint32
type ContentFormat uint8

const (
	CON Type = 0 // Confirmable
	NON Type = 1 //
	ACK Type = 2
	RST Type = 3

	KnownVer int = 1

	CCReq          CodeCategory = 0
	CCAckOK        CodeCategory = 2
	CCClientBadAck CodeCategory = 4
	CCServerBadAck CodeCategory = 5

	CodeEmpty Code = 0

	MethodCodeGET    Code = 1
	MethodCodePOST   Code = 2
	MethodCodePUT    Code = 3
	MethodCodeDELETE Code = 4

	ResponseCodeCreated                  Code = 65
	ResponseCodeDeleted                  Code = 66
	ResponseCodeValid                    Code = 67
	ResponseCodeChanged                  Code = 68
	ResponseCodeContent                  Code = 69
	ResponseCodeBadRequest               Code = 128
	ResponseCodeUnauthorized             Code = 129
	ResponseCodeBadOption                Code = 130
	ResponseCodeForbidden                Code = 131
	ResponseCodeNotFound                 Code = 132
	ResponseCodeMethodNotAllowed         Code = 133
	ResponseCodeNotAcceptable            Code = 134
	ResponseCodePreconditionFailed       Code = 140
	ResponseCodeRequestEntityTooLarge    Code = 141
	ResponseCodeUnsupportedContentFormat Code = 143
	ResponseCodeInternalServerError      Code = 160
	ResponseCodeNotImplemented           Code = 161
	ResponseCodeBadGateway               Code = 162
	ResponseCodeServiceUnavailable       Code = 163
	ResponseCodeGatewayTimeout           Code = 164
	ResponseCodeProxyingNotSupported     Code = 165

	OptionNumberReserved      OptionNumber = 0
	OptionNumberIfMatch       OptionNumber = 1
	OptionNumberUriHost       OptionNumber = 3
	OptionNumberETag          OptionNumber = 4
	OptionNumberIfNoneMatch   OptionNumber = 5
	OptionNumberUriPort       OptionNumber = 7
	OptionNumberLocationPath  OptionNumber = 8
	OptionNumberUriPath       OptionNumber = 11
	OptionNumberContentFormat OptionNumber = 12
	OptionNumberMaxAge        OptionNumber = 14
	OptionNumberUriQuery      OptionNumber = 15
	OptionNumberAccept        OptionNumber = 17
	OptionNumberLocationQuery OptionNumber = 20
	OptionNumberProxyUri      OptionNumber = 35
	OptionNumberProxyScheme   OptionNumber = 39
	OptionNumberSize1         OptionNumber = 60
	OptionNumberR128          OptionNumber = 128
	OptionNumberR132          OptionNumber = 132
	OptionNumberR136          OptionNumber = 136
	OptionNumberR140          OptionNumber = 140

	ContentFormatTextPlain   ContentFormat = 0
	ContentFormatLinkFormat  ContentFormat = 40
	ContentFormatXml         ContentFormat = 41
	ContentFormatOctetStream ContentFormat = 42
	ContentFormatExi         ContentFormat = 47
	ContentFormatJson        ContentFormat = 50
)
