package rb

import (
	"sync/atomic"
	"testing"
	"time"
)

const N = 100
const NLtd = 16

func TestRingBuf_Put_OneByOne(t *testing.T) {
	var err error
	rb := New(NLtd, WithDebugMode(true))
	size := rb.Cap() - 1
	// t.Logf("Ring Buffer created, capacity = %v, real size: %v", size+1, size)
	defer rb.Close()

	for i := uint32(0); i < size; i++ {
		err = rb.Enqueue(i)
		if err != nil {
			t.Fatalf("faild on i=%v. err: %+v", i, err)
		} else {
			// t.Logf("%5d. '%v' put, quantity = %v.", i, i, rb.Quantity())
		}
	}

	for i := uint32(size); i < uint32(size)+6; i++ {
		err = rb.Enqueue(i)
		if err != ErrQueueFull {
			t.Fatalf("> %3d. expect ErrQueueFull but no error raised. err: %+v", i, err)
		}
	}

	var it interface{}
	for i := 0; ; i++ {

		it, err = rb.Dequeue()
		if err != nil {
			if err == ErrQueueEmpty {
				break
			}
			t.Fatalf("faild on i=%v. err: %+v. item: %v", i, err, it)
		} else {
			// t.Logf("< %3d. '%v' GOT, quantity = %v.", i, it, rb.Quantity())
		}
	}
}

//
// go test ./... -race -run '^TestRingBuf_Put_Randomized$'
// go test ./ringbuf/rb -race -run '^TestRingBuf_Put_R.*'
//
func TestRingBuf_Put_Randomized(t *testing.T) {
	var maxN = NLtd * 1000
	putRandomized(t, maxN, NLtd, func(r RingBuffer) {
		// r.Debug(true)
	})
}

//
// go test ./ringbuf/rb -race -bench=. -run=none
// go test ./ringbuf/rb -race -bench '.*Put128' -run=none
//
// go test ./ringbuf/rb -race -bench=. -run=none -benchtime=3s
//
// go test ./ringbuf/rb -race -bench=. -benchmem -cpuprofile profile.out
// go test ./ringbuf/rb -race -bench=. -benchmem -memprofile memprofile.out -cpuprofile profile.out
// go tool pprof profile.out
//
// https://my.oschina.net/solate/blog/3034188
//
func BenchmarkRingBuf_Put16384(b *testing.B) {
	b.ResetTimer()
	putRandomized(b, b.N, 10000)
}

func BenchmarkRingBuf_Put1024(b *testing.B) {
	b.ResetTimer()
	putRandomized(b, b.N, 1000)
}

func BenchmarkRingBuf_Put128(b *testing.B) {
	b.ResetTimer()
	putRandomized(b, b.N, 100)
}

func putRandomized(t testing.TB, maxN int, queueSize uint32, opts ...func(r RingBuffer)) {
	var d1Closed int32
	d1, d2 := make(chan struct{}), make(chan struct{})

	_, noDebug := t.(*testing.B)
	rb := New(queueSize, WithDebugMode(!noDebug))
	for _, cb := range opts {
		cb(rb)
	}
	noDebug = !rb.Debug(!noDebug)
	t.Logf("Ring Buffer created, size = %v. maxN = %v", rb.Cap(), maxN)
	defer rb.Close()

	go func() {

		var err error
		var it interface{}
		var fetched []int
		// t.Logf("[GET] d1Closed: %v, err: %v", d1Closed, err)
		for i := 0; err != ErrQueueEmpty; i++ {
		retryGet:
			it, err = rb.Dequeue()
			if err != nil {
				if err == ErrQueueEmpty {
					// block till queue not empty
					if !noDebug {
						t.Logf("[GET] %5d. quantity = %v. EMPTY! block till queue not empty", i, rb.Quantity())
					}
					time.Sleep(1 * time.Microsecond)
					goto retryGet
				}
				t.Fatalf("[GET] failed on i=%v. err: %+v.", i, err)
			}

			fetched = append(fetched, it.(int))
			// t.Logf("[GET] %5d. '%v' GOT, quantity = %v.", i, it, rb.Quantity())
			// time.Sleep(50 * time.Millisecond)

			if atomic.LoadInt32(&d1Closed) == 1 {
				break
			}
		}
		close(d2)
		// t.Log("[GET] END")

		// checking the fetched
		last := fetched[0]
		for i := 1; i < len(fetched); i++ {
			ix := fetched[i]
			if ix != last+1 {
				t.Fatalf("[GET] the fetched sequence is wrong, expecting %v but got %v.", last+1, ix)
			}
			last = ix
		}
	}()

	go func() {

		var err error
		for i := 0; i < maxN; i++ {
		retryPut:
			err = rb.Enqueue(i)
			if err != nil {
				if err == ErrQueueFull {
					// block till queue not full
					if !noDebug {
						t.Logf("[PUT] %5d. quantity = %v. FULL! block till queue not full", i, rb.Quantity())
					}
					time.Sleep(1 * time.Microsecond)
					goto retryPut
				}
				t.Fatalf("[PUT] failed on i=%v. err: %+v.", i, err)
			}

			// t.Logf("[PUT] %5d. '%v' put, quantity = %v.", i, i, rb.Quantity())
			// time.Sleep(50 * time.Millisecond)
		}
		close(d1)
		atomic.StoreInt32(&d1Closed, 1)
		// t.Log("[PUT] END")

	}()

	<-d1
	<-d2
}
