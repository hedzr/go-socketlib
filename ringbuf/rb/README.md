
## For comparing with others

```bash
$ go test ./ringbuf/rb -v -race -run 'TestQueuePutGet'
=== RUN   TestQueuePutGet
    TestQueuePutGet: rb_test.go:225: Grp: 16, Times: 1360000, use: 16.019835259s, 11.779µs/op
    TestQueuePutGet: rb_test.go:226: Put: 1360000, use: 9.277736289s, 6.821µs/op
    TestQueuePutGet: rb_test.go:227: Get: 1360000, use: 3.963994253s, 2.914µs/op
--- PASS: TestQueuePutGet (16.02s)
```




## Bench

### Bench A


```bash
$ go test ./ringbuf/rb -race -bench='BenchmarkRingBuf' -run=none
goos: darwin
goarch: amd64
pkg: github.com/hedzr/socketlib/ringbuf/rb
BenchmarkRingBuf_Put16384-4   	1000000000	         0.201 ns/op
--- BENCH: BenchmarkRingBuf_Put16384-4
    rb_test.go:105: Ring Buffer created, size = 16384. maxN = 1, dbg: false
    rb_test.go:188: Waits: get: 0, put: 0
    rb_test.go:105: Ring Buffer created, size = 16384. maxN = 100, dbg: false
    rb_test.go:188: Waits: get: 0, put: 0
    rb_test.go:105: Ring Buffer created, size = 16384. maxN = 10000, dbg: false
    rb_test.go:188: Waits: get: 0, put: 0
    rb_test.go:105: Ring Buffer created, size = 16384. maxN = 596823, dbg: false
    rb_test.go:188: Waits: get: 0, put: 0
    rb_test.go:105: Ring Buffer created, size = 16384. maxN = 4861959, dbg: false
    rb_test.go:188: Waits: get: 1, put: 0
	... [output truncated]
BenchmarkRingBuf_Put1024-4    	1000000000	         0.0328 ns/op
--- BENCH: BenchmarkRingBuf_Put1024-4
    rb_test.go:105: Ring Buffer created, size = 1024. maxN = 1, dbg: false
    rb_test.go:188: Waits: get: 0, put: 0
    rb_test.go:105: Ring Buffer created, size = 1024. maxN = 100, dbg: false
    rb_test.go:188: Waits: get: 0, put: 0
    rb_test.go:105: Ring Buffer created, size = 1024. maxN = 10000, dbg: false
    rb_test.go:188: Waits: get: 1, put: 0
    rb_test.go:105: Ring Buffer created, size = 1024. maxN = 822657, dbg: false
    rb_test.go:188: Waits: get: 0, put: 0
    rb_test.go:105: Ring Buffer created, size = 1024. maxN = 14712006, dbg: false
    rb_test.go:188: Waits: get: 0, put: 0
	... [output truncated]
BenchmarkRingBuf_Put128-4     	1000000000	         0.0578 ns/op
--- BENCH: BenchmarkRingBuf_Put128-4
    rb_test.go:105: Ring Buffer created, size = 128. maxN = 1, dbg: false
    rb_test.go:188: Waits: get: 0, put: 0
    rb_test.go:105: Ring Buffer created, size = 128. maxN = 100, dbg: false
    rb_test.go:188: Waits: get: 0, put: 0
    rb_test.go:105: Ring Buffer created, size = 128. maxN = 10000, dbg: false
    rb_test.go:188: Waits: get: 16, put: 0
    rb_test.go:105: Ring Buffer created, size = 128. maxN = 479848, dbg: false
    rb_test.go:188: Waits: get: 6, put: 0
    rb_test.go:105: Ring Buffer created, size = 128. maxN = 7635192, dbg: false
    rb_test.go:188: Waits: get: 1, put: 0
	... [output truncated]
PASS
ok  	github.com/hedzr/socketlib/ringbuf/rb	1.718s
```


### Bench B


```bash
$ go test ./ringbuf/rb -race -bench 'BenchmarkPutGet' -run=none -benchtime=10s -v
goos: darwin
goarch: amd64
pkg: github.com/hedzr/socketlib/ringbuf/rb
BenchmarkPutGet
    BenchmarkPutGet: rb_test.go:257: Grp: 16, Times: 136, use: 2.279890383s, 16.763899ms/op | cnt = 1
    BenchmarkPutGet: rb_test.go:258: Put: 136, use: 14.786715ms, 108.725µs/op
    BenchmarkPutGet: rb_test.go:259: Get: 136, use: 11.828063ms, 86.971µs/op
    BenchmarkPutGet: rb_test.go:257: Grp: 16, Times: 544, use: 2.286623579s, 4.203352ms/op | cnt = 4
    BenchmarkPutGet: rb_test.go:258: Put: 544, use: 13.760842ms, 25.295µs/op
    BenchmarkPutGet: rb_test.go:259: Get: 544, use: 11.326448ms, 20.82µs/op
    BenchmarkPutGet: rb_test.go:257: Grp: 16, Times: 2720, use: 2.268485793s, 834.002µs/op | cnt = 20
    BenchmarkPutGet: rb_test.go:258: Put: 2720, use: 34.898867ms, 12.83µs/op
    BenchmarkPutGet: rb_test.go:259: Get: 2720, use: 16.811108ms, 6.18µs/op
    BenchmarkPutGet: rb_test.go:257: Grp: 16, Times: 14280, use: 2.346420828s, 164.315µs/op | cnt = 105
    BenchmarkPutGet: rb_test.go:258: Put: 14280, use: 119.552647ms, 8.372µs/op
    BenchmarkPutGet: rb_test.go:259: Get: 14280, use: 43.536215ms, 3.048µs/op
    BenchmarkPutGet: rb_test.go:257: Grp: 16, Times: 72896, use: 2.740715898s, 37.597µs/op | cnt = 536
    BenchmarkPutGet: rb_test.go:258: Put: 72896, use: 372.948351ms, 5.116µs/op
    BenchmarkPutGet: rb_test.go:259: Get: 72896, use: 201.965619ms, 2.77µs/op
    BenchmarkPutGet: rb_test.go:257: Grp: 16, Times: 319056, use: 4.714743406s, 14.777µs/op | cnt = 2346
    BenchmarkPutGet: rb_test.go:258: Put: 319056, use: 1.65687515s, 5.193µs/op
    BenchmarkPutGet: rb_test.go:259: Get: 319056, use: 814.404547ms, 2.552µs/op
    BenchmarkPutGet: rb_test.go:257: Grp: 16, Times: 811920, use: 8.80726798s, 10.847µs/op | cnt = 5970
    BenchmarkPutGet: rb_test.go:258: Put: 811920, use: 4.321487768s, 5.322µs/op
    BenchmarkPutGet: rb_test.go:259: Get: 811920, use: 2.112014177s, 2.601µs/op
    BenchmarkPutGet: rb_test.go:257: Grp: 16, Times: 1106088, use: 11.32722686s, 10.24µs/op | cnt = 8133
    BenchmarkPutGet: rb_test.go:258: Put: 1106088, use: 6.082846481s, 5.499µs/op
    BenchmarkPutGet: rb_test.go:259: Get: 1106088, use: 2.785739397s, 2.518µs/op
BenchmarkPutGet-4   	    8133	   1392783 ns/op
PASS
ok  	github.com/hedzr/socketlib/ringbuf/rb	37.138s
```

Note that the result of last group should be concerned:

```bash
    BenchmarkPutGet: rb_test.go:257: Grp: 16, Times: 1106088, use: 11.32722686s, 10.24µs/op | cnt = 8133
    BenchmarkPutGet: rb_test.go:258: Put: 1106088, use: 6.082846481s, 5.499µs/op
    BenchmarkPutGet: rb_test.go:259: Get: 1106088, use: 2.785739397s, 2.518µs/op
```



## END