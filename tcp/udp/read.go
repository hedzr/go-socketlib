package udp

import (
	"context"
	"time"

	"gopkg.in/hedzr/go-ringbuf.v1/fast"

	"github.com/hedzr/go-socketlib/tcp/base"
)

func (s *Obj) readPump(ctx context.Context) {
	var (
		err       error
		it        interface{}
		processed bool
		retry     = 0
	)

	s.wg.Add(1)

	defer func() {
		if err == nil {
			s.Debugf("    .. readPump end.")
		} else {
			s.Errorf("    .. readPump end with error: %v", err)
		}
		s.wg.Done()
	}()

	for err == nil {
		select {
		case <-ctx.Done():
			return
		default:
		}

		it, err = s.rb.Dequeue()
		if err != nil {
			if err == fast.ErrQueueEmpty {
				// block till queue not empty
				time.Sleep(time.Duration(retry) * time.Microsecond)
				retry++
				if s.debugMode && retry > 2000 {
					s.Warnf("[udp.readPump] (retry %v). quantity = %v. EMPTY! block till queue not empty.", retry, s.rb.Quantity())
				}
				err = nil
				continue
			}
			s.Errorf("[udp.readPump] dequeue failed. err: %+v.", err)
			if s.InterceptorHolder != nil && s.InterceptorHolder.ProtocolInterceptor() != nil {
				s.InterceptorHolder.ProtocolInterceptor().OnError(ctx, s.baseConn, err)
			}
		}
		retry = 1

		if packet, ok := it.(*base.UdpPacket); ok {
			if s.InterceptorHolder != nil && s.InterceptorHolder.ProtocolInterceptor() != nil {
				if processed, err = s.InterceptorHolder.ProtocolInterceptor().OnUDPReading(ctx, s, packet); err != nil {
					s.Warnf("[udp.readPump] protocolInterceptor got error: %v", err)
					err = nil
					continue
				}
				if processed {
					continue
				}
			}

			s.Tracef("[udp.readPump] dequeued %v -> % x %q", packet.RemoteAddr, packet.Data, string(packet.Data))
			if s.IsConnected() == false { // while no interpreter defined,
				s.Write(packet) // run as an echo server
			}
		}
		// t.Logf("[GET] %5d. '%v' GOT, quantity = %v.", i, it, fast.Quantity())
	}
}
