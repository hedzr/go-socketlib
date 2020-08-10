package pi

import (
	"context"
	"github.com/hedzr/go-socketlib/coaplib/message"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
	"github.com/hedzr/log"
)

func NewCoAPClientInterceptor() protocol.ClientInterceptor {
	return &Pic{
		CoapCmd: NewCoapCmd(),
	}
}

type Pic struct {
	*CoapCmd
}

func (p *Pic) OnConnected(ctx context.Context, c base.Conn) {

	// update the default internal logger to user customized
	message.SetLogger(c.Logger())

	c.Logger().Debugf("OnConnected: %v", c)
	p.SetBase(c.RemoteAddr())
	// p.Get(ctx, "coap://coap.me/.well-known/core")
	return

}

func (p *Pic) OnClosing(c base.Conn, reason int) {
	c.Logger().Debugf("OnClosing")
	return
}

func (p *Pic) OnClosed(c base.Conn, reason int) {
	c.Logger().Debugf("OnClosed")
	return

}

func (p *Pic) OnError(ctx context.Context, c base.Conn, err error) {
	c.Logger().Errorf("OnError: %v", err)
	return

}

func (p *Pic) OnReading(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	// c.Logger().Debugf("OnReading")
	return
}

func (p *Pic) OnWriting(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	// c.Logger().Debugf("OnWriting")
	return
}

func (p *Pic) OnUDPReading(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {

	m := new(message.Message)
	if err = m.Decode(packet.Data); err != nil {
		c.Errorf("OnReading, decode CoAP message failed: %v", err) // don't break pump loopers
		return
	}

	// explain m
	//...
	c.Debugf("[COAP] ⬇ OnUDPReading: msg = %v", m.String())

	switch m.Type {
	case message.ACK:
		if sent, ok := p.sentMessages[m.MessageID]; ok {
			delete(p.sentMessages, m.MessageID)
			//c.Tracef("sent, recv := %v, %v", sent, m)
			if sent.OnACK != nil {
				if err = sent.OnACK(ctx, sent, m); err != nil {
					c.Errorf("OnReading, sent.OnACK() failed: %v", err)
				}
			} else if m.MediaType == message.AppLinkFormat {
				if err = p.OnAckLinkFormat(ctx, sent, m); err != nil {
					c.Errorf("OnReading, pic.OnAckLinkFormat() failed: %v", err)
				}
			}
		}

	case message.CON:
		if opt := m.FindOption(message.OptionNumberObserve); opt != nil {
			if _, ok := opt.(interface{ Uint64Data() uint64 }); ok {
				err = p.ACK(ctx, m.MessageID)
				token := m.Token
				if sent, ok := p.observed[token]; ok {
					sent.RaiseOnEvent(ctx, sent.Href(), token, m)
				}
			}
		}
	}

	processed = true
	return
}

func (p *Pic) OnUDPWriting(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	c.Debugf("OnUDPWriting to %s: %d bytes", packet.RemoteAddr.String(), len(packet.Data))

	// time.Sleep(100 & time.Millisecond)

	// processed = false
	return
}
