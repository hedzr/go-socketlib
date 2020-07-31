package uri

import "net/url"

type URI struct {
	*url.URL
}

func (s *URI) Parse(raw string) (err error) {
	s.URL, err = url.Parse(raw)
	if err == nil {
		if s.Port() == "" {
			switch s.Scheme {
			case "coap":
				s.Host += ":5683"
			case "coaps":
				s.Host += ":5684"
			}
		}
	}
	return
}

func ParseURI(raw string) (uri URI, err error) {
	err = uri.Parse(raw)
	return
}

func ParseURIFast(raw string) (uri URI) {
	_ = uri.Parse(raw)
	return
}
