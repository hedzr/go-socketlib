package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	logz "log/slog"

	"github.com/hedzr/go-socketlib/net"
	"github.com/hedzr/go-socketlib/net/api"
)

func newPop3Server(opts ...net.ServerOpt) *pop3S {
	s := &pop3S{
		sessions: make(map[api.Response]*pop3Session),
	}

	s.onAuthenticate = s.defaultAuthenticate
	s.knownCommands = map[string]pop3Handler{
		"user": s.handleUser,
		"pass": s.handlePass,
		"apop": s.handleApop,
		"stat": s.handleStat,
		"uidl": s.handleUidl,
		"list": s.handleList,
		"retr": s.handleRetr,
		"dele": s.handleDele,
		"rest": s.handleRest,
		"top ": s.handleTop,
		"noop": s.handleNoop,
		"quit": s.handleQuit,
	}

	s.Server = net.NewServer(pop3serverAddress, append(opts,
		net.WithServerOnClientConnected(func(w api.Response, ss net.Server) {
			sess := s.newSession(w)
			s.Send(sess, "+OK pop3Session server (test, go-socketlib) ready.\r\n")
		}),
		net.WithServerOnProcessData(func(data []byte, w api.Response, r api.Request) (nn int, err error) {
			logz.Debug("[server] RECEIVED:", "data", string(data), "client.addr", w.RemoteAddr())
			nn, err = s.handleData(data, w, r)
			return
		}),
	)...)

	s.Logger = s.Server.(net.Logger)
	s.ExtraPrinters = s.Server.(net.ExtraPrinters)
	return s
}

type pop3S struct {
	net.Server
	sessions       map[api.Response]*pop3Session
	knownCommands  map[string]pop3Handler
	onAuthenticate AuthenticateHandler
	closed         int32

	net.ExtraPrinters // import ExtraPrinters
	net.Logger        // import Logger
}

type AuthenticateHandler func(sess *pop3Session, user, pass string) (passed bool)

type pop3Handler func(sess *pop3Session, data []byte, w api.Response) (ate int)

type pop3Session struct {
	api.Response

	server *pop3S

	cache *bytes.Buffer

	username   string
	password   string
	digestName string
	digest     string // md5 -> base64
	lastBeat   time.Time

	authenticating int32
	authenticated  int32
	mailboxes      map[string]*mailboxS
}

type mailboxS struct {
	messages      []*msgS
	totalCapacity int
}

type msgS struct {
	cap int
}

// Close cleanup itself, includes internal net connections, sessions, ...
func (s *pop3S) Close() {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		for w, sess := range s.sessions {
			w.Close()
			sess.Close()
		}
		s.sessions = nil

	}
	s.Server.Close()
}

func (s *pop3S) defaultAuthenticate(sess *pop3Session, user, pass string) (passed bool) {
	if atomic.CompareAndSwapInt32(&sess.authenticating, 0, 1) {
		defer func() {
			atomic.CompareAndSwapInt32(&sess.authenticating, 1, 0)
		}()
		if passed = user == "user"; passed {
			if atomic.CompareAndSwapInt32(&sess.authenticating, 0, 1) {
				return
			}
		}
	}
	return
}

func (s *pop3S) newSession(w api.Response) (p *pop3Session) {
	var ok bool
	if p, ok = s.sessions[w]; ok {
		return
	}

	p = &pop3Session{Response: w, server: s, username: ""}
	p.cache = bytes.NewBuffer([]byte{})
	// p.cacheWriter = bufio.NewWriterSize(p.cache, 4096)
	p.mailboxes = map[string]*mailboxS{
		"user": {
			messages: []*msgS{
				{2400},
				{5000},
				{5086},
			},
			totalCapacity: 12486,
		},
	}

	s.sessions[w] = p
	return p
}

func (s *pop3S) handleData(data []byte, w api.Response, r api.Request) (nn int, err error) {
	if sess, ok := s.sessions[w]; ok {
		nn, err = sess.handleData(data, w, r)
	}
	return
}

func (s *pop3Session) handleData(data []byte, w api.Response, r api.Request) (nn int, err error) {
	nn, err = s.cache.Write(data)
	// nn, err = s.cacheWriter.Write(data)
	if err != nil {
		return
	}
	// err = s.cacheWriter.Flush()

	// try dispatching each line as pop3 cmd
	scanner := bufio.NewScanner(s.cache)
	for scanner.Scan() {
		s.dispatchCmd(scanner.Text(), w)
	}
	return
}

func (s *pop3Session) dispatchCmd(cmd string, w api.Response) (n int) {
	println(" [pop3Session] dispatchCmd:", "cmd", cmd)
	if h, ok := s.server.knownCommands[cmd[:4]]; ok {
		n = h(s, []byte(cmd[4:]), w)
	}
	return
}

func (s *pop3S) handleUser(sess *pop3Session, data []byte, w api.Response) (ate int) {
	ate = len(data)
	// user<SP>username<CRLF>
	pos := skipws(data, 0)
	sess.username, pos = tillcr(data, pos)
	if pos != ate {
		panic(fmt.Sprintf("wrong command length: data-len=%d, parsed-pos=%d, data=%v", ate, pos, string(data)))
	}
	s.Send(sess, fmt.Sprintf("+OK hello %s!\r\n", sess.username))
	return
}

func (s *pop3S) handlePass(sess *pop3Session, data []byte, w api.Response) (ate int) {
	ate = len(data)
	// pass<SP>password<CRLF>
	pos := skipws(data, 0)
	sess.password, pos = tillcr(data, pos)
	if pos != ate {
		panic(fmt.Sprintf("wrong command length: ate=%d, parsed-pos=%d, data=%v", ate, pos, string(data)))
	}
	if sess.server.onAuthenticate(sess, sess.username, sess.password) {
		s.Send(sess, fmt.Sprintf("+OK %d message(s) [%d byte(s)]\r\n", len(sess.mailboxes["user"].messages), sess.mailboxes["user"].totalCapacity))
	} else {
		s.Send(sess, "-ERR -717 BAD authentication(s).\r\n")
	}
	return
}

func (s *pop3S) handleApop(sess *pop3Session, data []byte, w api.Response) (ate int) {
	ate = len(data)
	// apop<SP>name,digest<CRLF>
	pos := skipws(data, 0)
	sess.digestName, pos = tillChars(data, pos, ',')
	sess.digest, pos = tillcr(data, pos)
	if pos != ate {
		panic(fmt.Sprintf("wrong command length: ate=%d, parsed-pos=%d, data=%v", ate, pos, string(data)))
	}
	s.SendOK(sess)
	return
}

func (s *pop3S) handleStat(sess *pop3Session, data []byte, w api.Response) (ate int) {
	ate = len(data)
	// stat<CRLF>
	pos := skipws(data, 0)
	_, pos = tillcr(data, pos)
	if pos != ate {
		panic(fmt.Sprintf("wrong command length: ate=%d, parsed-pos=%d, data=%v", ate, pos, string(data)))
	}
	s.SendOK(sess)
	return
}

func (s *pop3S) handleUidl(sess *pop3Session, data []byte, w api.Response) (ate int) {
	ate = len(data)
	// uidl<SP>msg#<CRLF>
	pos := skipws(data, 0)
	var msgn string
	msgn, pos = tillcr(data, pos)
	if pos != ate {
		panic(fmt.Sprintf("wrong command length: ate=%d, parsed-pos=%d, data=%v", ate, pos, string(data)))
	}
	if num, err := strconv.Atoi(msgn); err != nil {
		logz.Error("cannot parse number of message", "err", err, "data", string(data))
	} else {
		logz.Debug("uidl msg#=<num>", "num", num)
	}
	s.SendOK(sess)
	return
}

func (s *pop3S) handleList(sess *pop3Session, data []byte, w api.Response) (ate int) {
	ate = len(data)
	// list<SP>[MSG#]<CRLF>
	pos := skipws(data, 0)
	var msgn string
	msgn, pos = tillcr(data, pos)
	if pos != ate {
		panic(fmt.Sprintf("wrong command length: ate=%d, parsed-pos=%d, data=%v", ate, pos, string(data)))
	}
	if num, err := strconv.Atoi(msgn); err != nil {
		logz.Error("cannot parse number of message", "err", err, "data", string(data))
	} else {
		logz.Debug("list msg#=<num>", "num", num)
	}
	s.SendOK(sess)
	for i, msg := range sess.mailboxes["user"].messages {
		s.Send(sess, fmt.Sprintf("%d %d\r\n", i+1, msg.cap))
	}
	return
}

func (s *pop3S) handleRetr(sess *pop3Session, data []byte, w api.Response) (ate int) {
	ate = len(data)
	// retr<SP>msg#<CRLF>
	pos := skipws(data, 0)
	var msgn string
	msgn, pos = tillcr(data, pos)
	if pos != ate {
		panic(fmt.Sprintf("wrong command length: ate=%d, parsed-pos=%d, data=%v", ate, pos, string(data)))
	}
	if num, err := strconv.Atoi(msgn); err != nil {
		logz.Error("cannot parse number of message", "err", err, "data", string(data))
	} else {
		logz.Debug("list msg#=<num>", "num", num)
		if num < len(sess.mailboxes["user"].messages) {
			s.Send(sess, fmt.Sprintf("+OK %d octets\r\n", sess.mailboxes["user"].messages[num].cap))
		} else {
			s.Send(sess, "-ERR unknown error\r\n")
		}
	}
	return
}

func (s *pop3S) handleDele(sess *pop3Session, data []byte, w api.Response) (ate int) {
	ate = len(data)
	// dele<SP>msg#<CRLF>
	pos := skipws(data, 0)
	var msgn string
	msgn, pos = tillcr(data, pos)
	if pos != ate {
		panic(fmt.Sprintf("wrong command length: ate=%d, parsed-pos=%d, data=%v", ate, pos, string(data)))
	}
	if num, err := strconv.Atoi(msgn); err != nil {
		logz.Error("cannot parse number of message", "err", err, "data", string(data))
	} else {
		logz.Debug("dele msg#=<num>", "num", num)
	}
	s.SendOK(sess)
	return
}

func (s *pop3S) handleRest(sess *pop3Session, data []byte, w api.Response) (ate int) {
	ate = len(data)
	// rest<CRLF>
	pos := skipws(data, 0)
	_, pos = tillcr(data, pos)
	if pos != ate {
		panic(fmt.Sprintf("wrong command length: ate=%d, parsed-pos=%d, data=%v", ate, pos, string(data)))
	}
	s.SendOK(sess)
	sess.lastBeat = time.Now().UTC()
	return
}

func (s *pop3S) handleTop(sess *pop3Session, data []byte, w api.Response) (ate int) {
	ate = len(data)
	// top<SP>msg#<SP>n<CRLF>
	pos := skipws(data, 0)
	_, pos = tillChars(data, pos, ' ')

	var msgn string
	msgn, pos = tillChars(data, pos, ' ')
	var msgn2 string
	msgn2, pos = tillcr(data, pos)
	if pos != ate {
		panic(fmt.Sprintf("wrong command length: ate=%d, parsed-pos=%d, data=%v", ate, pos, string(data)))
	}

	var num1, num2 int
	var err error
	if num1, err = strconv.Atoi(msgn); err != nil {
		logz.Error("cannot parse number of message", "err", err, "data", string(data))
		return
	}

	if num2, err = strconv.Atoi(msgn2); err != nil {
		logz.Error("cannot parse number of message", "err", err, "data", string(data))
		return
	}

	logz.Debug("top msg#=<num> n=<n>", "num", num1, "n", num2)
	s.SendOK(sess)
	return
}

func (s *pop3S) handleNoop(sess *pop3Session, data []byte, w api.Response) (ate int) {
	ate = len(data)
	// noop<CRLF>
	pos := skipws(data, 0)
	_, pos = tillcr(data, pos)
	if pos != ate {
		panic(fmt.Sprintf("wrong command length: ate=%d, parsed-pos=%d, data=%v", ate, pos, string(data)))
	}
	sess.lastBeat = time.Now().UTC()
	s.SendOK(sess)
	return
}

func (s *pop3S) handleQuit(sess *pop3Session, data []byte, w api.Response) (ate int) {
	ate = len(data)
	// quit<CRLF>
	pos := skipws(data, 0)
	_, pos = tillcr(data, pos)
	if pos != ate {
		panic(fmt.Sprintf("wrong command length: ate=%d, parsed-pos=%d, data=%v", ate, pos, string(data)))
	}

	println("[pop3S] Quiting...")
	time.Sleep(1 * time.Second)
	sess.Close()
	delete(s.sessions, w)
	w.Close()

	return
}

func (s *pop3S) Send(sess *pop3Session, msg string) {
	if _, err := sess.Write([]byte(msg)); err != nil {
		logz.Error("cannot write to connection", "remote.addr", sess.RemoteAddrString(), "msg", msg, "err", err)
	}
}

func (s *pop3S) SendOK(sess *pop3Session) {
	s.Send(sess, fmt.Sprintf("+OK %d %d\r\n", len(sess.mailboxes["user"].messages), sess.mailboxes["user"].totalCapacity))
}

func (s *pop3Session) Close() {
}

// var pop3serverMap = make(map[api.Response]*pop3Session)

const (
	pop3Port    = 110
	pop3TlsPort = 995
)
