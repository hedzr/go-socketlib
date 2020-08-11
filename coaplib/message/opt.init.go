package message

var optDecoder *optionDecoder

func init() {
	optDecoder = &optionDecoder{}
	optDecoder.decoders = map[OptionNumber]func(optNum OptionNumber, data []byte) Option{
		0:                         nil,
		OptionNumberIfMatch:       optDecoder.decodeOptionIfMatch,
		OptionNumberURIHost:       optDecoder.decodeOptionUriHost,
		OptionNumberETag:          optDecoder.decodeOptionETag,
		OptionNumberIfNoneMatch:   optDecoder.decodeOptionIfNoneMatch,
		OptionNumberObserve:       optDecoder.decodeOptionObserve,
		OptionNumberURIPort:       optDecoder.decodeOptionUriPort,
		OptionNumberLocationPath:  optDecoder.decodeOptionLocationPath,
		OptionNumberURIPath:       optDecoder.decodeOptionUriPath,
		OptionNumberContentFormat: optDecoder.decodeOptionContentFormat,
		OptionNumberMaxAge:        optDecoder.decodeOptionMaxAge,
		OptionNumberURIQuery:      optDecoder.decodeOptionUriQuery,
		OptionNumberAccept:        optDecoder.decodeOptionAccept,
		OptionNumberLocationQuery: optDecoder.decodeOptionLocationQuery,
		OptionNumberBlock2:        optDecoder.decodeOptionBlock2,
		OptionNumberBlock1:        optDecoder.decodeOptionBlock1, //setBlock1ToPayload,
		OptionNumberSize2:         optDecoder.decodeOptionSize2,
		OptionNumberProxyURI:      optDecoder.decodeOptionProxyUri,
		OptionNumberProxyScheme:   optDecoder.decodeOptionProxyScheme,
		OptionNumberSize1:         optDecoder.decodeOptionSize1,
		OptionNumberR128:          nil,
		OptionNumberR132:          nil,
		OptionNumberR136:          nil,
		OptionNumberR140:          nil,
		OptionNumberNoResponse:    nil,
	}

}
