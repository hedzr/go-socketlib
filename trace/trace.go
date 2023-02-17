// Copyright © 2020 Hedzr Yeh.

package trace

// import (
//	"sync"
//	"sync/atomic"
// )
//
// var tracing struct {
//	sync.Mutex
//	enabled int32
// }
//
// // Start starts the trace subsystem
// func Start() (err error) {
//	if atomic.CompareAndSwapInt32(&tracing.enabled, 0, 1) {
//		tracing.Lock()
//		defer tracing.Unlock()
//
//		// trace.Start()
//
//	}
//
//	return
// }
//
// // Stop stops the trace subsystem
// func Stop() {
//	if atomic.CompareAndSwapInt32(&tracing.enabled, 1, 0) {
//		tracing.Lock()
//		defer tracing.Unlock()
//
//		// ...
//
//	}
// }
//
// // IsEnabled return the state of this trace subsystem
// func IsEnabled() bool {
//	enabled := atomic.LoadInt32(&tracing.enabled)
//	return enabled == 1
// }
