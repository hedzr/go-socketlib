package pkg

func NewBuilder() Builder {
	b := &pkgBuilder{
		//reservedBytes: reservedBytes,
		//leadingBytes:  make([]byte, reservedBytes),
	}
	return b.New()
}

type pkgBuilder struct {
	reservedBytes int
	leadingBytes  []byte
	pkg           *Pkg
}

func (s *pkgBuilder) New() Builder {
	s.pkg = new(Pkg)
	return s
}

func (s *pkgBuilder) Build() Package {
	pkg := s.pkg
	s.pkg = nil
	return pkg
}

func (s *pkgBuilder) Bytes(reservedBytes int) []byte { return ToBytes(s.pkg, reservedBytes) }

func (s *pkgBuilder) Command(cmd Command) PackageBuilder {
	s.pkg.Command = cmd
	return s
}

func (s *pkgBuilder) Body(body []byte) PackageBuilder {
	s.pkg.Body = body
	return s
}

func (s *pkgBuilder) SendOOB(oob []byte) PackageBuilder {
	s.pkg.Command = PCOOb
	s.pkg.Body = oob
	return s
}
