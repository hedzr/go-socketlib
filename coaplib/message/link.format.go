package message

import (
	"fmt"
	"gopkg.in/hedzr/errors.v2"
	"strconv"
	"strings"
)

func NewLinkFormat() *LinkFormat {
	return &LinkFormat{
		Resources:   map[string]*lfResource{},
		Observables: map[string]*lfResource{},
	}
}

func NewLinkFormatParser() *LinkFormatParser {
	return &LinkFormatParser{}
}

type LinkFormatParser struct {
	cachedBin []byte
}

type LinkFormat struct {
	LinkFormatParser
	ResArray    []*lfResource
	Resources   map[string]*lfResource
	Observables map[string]*lfResource
}

type lfKV map[string]interface{}

type lfResource struct {
	src string
	lfKV
	Size        int
	Title       string
	ResType     string
	If          string
	ContentType MediaType
	Observable  bool
}

func (s *LinkFormat) Parse(data string) (err error) {
	s.ResArray, err = s.LinkFormatParser.Parse(data)
	if err == nil {
		for _, res := range s.ResArray {
			if res.Observable {
				s.Observables[res.src] = res
				continue
			}
			s.Resources[res.src] = res
		}
	}
	return
}

func (s *LinkFormat) ToBytes() []byte {
	return s.LinkFormatParser.ToBytes(s.ResArray)
}

func (s *LinkFormat) Bytes() []byte {
	return s.LinkFormatParser.ToBytes(s.ResArray)
}

func (s *LinkFormat) String() string {
	return string(s.Bytes())
}

func (s *LinkFormatParser) Parse(data string) (res []*lfResource, err error) {
	for _, it := range strings.Split(data, ",") {
		sl := strings.Split(it, ";")
		if len(sl) >= 1 {
			lfRes := &lfResource{
				src:         strings.Trim(sl[0], "<>"),
				Size:        -1,
				ContentType: MediaTypeUndefined,
				lfKV:        make(lfKV),
			}
			for i, z := range sl {
				if i > 0 {
					kv := strings.Split(z, "=")
					if len(kv) == 1 {
						if kv[0] == "obs" {
							//lfRes.lfKV[kv[0]] = true
							lfRes.Observable = true
						}
						continue
					}

					k, v := kv[0], kv[1]
					if v[0] == '"' {
						if t, e := strconv.Unquote(v); e == nil {
							if k == "title" {
								lfRes.Title = t
							} else if k == "rt" {
								lfRes.ResType = t
							} else if k == "if" {
								lfRes.If = t
							} else {
								lfRes.lfKV[k] = t
							}
						} else {
							err = errors.New("Unquote a string failed: Unquote(%q) -> %v", v, e)
							logger.Errorf("Unquote a string failed: Unquote(%q) -> %v", v, e)
						}
						// } else if i, err := strconv.ParseFloat(v, 64); err == nil {
						//	lfRes.lfKV[k] = i
					} else if i, err := strconv.ParseInt(v, 0, 64); err == nil {
						if k == "sz" {
							lfRes.Size = int(i)
						} else if k == "ct" {
							lfRes.ContentType = MediaType(i)
						} else {
							lfRes.lfKV[k] = i
						}
					} else if i, err := strconv.ParseBool(v); err == nil {
						lfRes.lfKV[k] = i
					} else {
						lfRes.lfKV[k] = v
					}
				}
			}
			res = append(res, lfRes)

		} else {
			//	err = errors.New("wrong link format: %q", data)
			//	logger.Errorf("wrong link format: %q", data)
		}
	}
	return
}

func (s *LinkFormatParser) ToBytes(res []*lfResource) []byte {
	if s.cachedBin == nil {
		var ss strings.Builder
		for i, rs := range res {
			if i > 0 {
				ss.WriteRune(',')
			}

			ss.WriteRune('<')
			ss.WriteString(rs.src)
			ss.WriteRune('>')

			if len(rs.ResType) > 0 {
				ss.WriteString(";rt=")
				ss.WriteString(strconv.Quote(rs.ResType))
			}
			if int(rs.ContentType) >= 0 && rs.ContentType != MediaTypeUndefined {
				ss.WriteString(";ct=")
				ss.WriteString(strconv.Itoa(int(rs.ContentType)))
			}
			if rs.Size >= 0 {
				ss.WriteString(";sz=")
				ss.WriteString(strconv.Itoa(rs.Size))
			}
			if rs.Observable {
				ss.WriteString(";obs")
			}
			if len(rs.If) > 0 {
				ss.WriteString(";if=")
				ss.WriteString(strconv.Quote(rs.If))
			}

			for k, v := range rs.lfKV {
				ss.WriteRune(';')
				ss.WriteString(k)
				if vs, ok := v.(string); ok {
					ss.WriteRune('=')
					ss.WriteString(strconv.Quote(vs))
				} else if _, ok := v.(bool); ok {
				} else {
					ss.WriteRune('=')
					ss.WriteString(fmt.Sprintf("%v", v))
				}
			}

			if len(rs.Title) > 0 {
				ss.WriteString(";title=")
				ss.WriteString(strconv.Quote(rs.Title))
			}
		}
		s.cachedBin = []byte(ss.String())
	}
	return s.cachedBin
}
