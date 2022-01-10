/*
 * Copyright © 2020 Hedzr Yeh.
 */

package internal

import (
	"context"
	"io"
)

// here is some syntaxic sugar inspired by the Tomas Senart's video,
// it allows me to inline the Reader interface
type readerFunc func(p []byte) (n int, err error)

func (rf readerFunc) Read(p []byte) (n int, err error) { return rf(p) }

// Copy _
// slightly modified function signature:
// - context has been added in order to propagate cancelation
// - I do not return the number of bytes written, has it is not useful in my use case
func Copy(ctx context.Context, dst io.Writer, src io.Reader) error {

	// Copy will call the Reader and Writer interface multiple time, in order
	// to copy by chunk (avoiding loading the whole file in memory).
	// I insert the ability to cancel before read time as it is the earliest
	// possible in the call process.
	_, err := io.Copy(dst, readerFunc(func(p []byte) (int, error) {

		// golang non-blocking channel: https://gobyexample.com/non-blocking-channel-operations
		select {

		// if context has been canceled
		case <-ctx.Done():
			// stop process and propagate "context canceled" error
			return 0, ctx.Err()
		default:
			// otherwise just run default io.Reader implementation
			return src.Read(p)
		}
	}))
	return err
}

// Copyd _
func Copyd(done <-chan struct{}, dst io.Writer, src io.Reader) error {

	// Copy will call the Reader and Writer interface multiple time, in order
	// to copy by chunk (avoiding loading the whole file in memory).
	// I insert the ability to cancel before read time as it is the earliest
	// possible in the call process.
	_, err := io.Copy(dst, readerFunc(func(p []byte) (int, error) {

		// golang non-blocking channel: https://gobyexample.com/non-blocking-channel-operations
		select {

		// if context has been canceled
		case <-done:
			// stop process and propagate "context canceled" error
			return 0, nil
		default:
			// otherwise just run default io.Reader implementation
			return src.Read(p)
		}
	}))
	return err
}

type writerFunc func(p []byte) (n int, err error)

func (wf writerFunc) Write(p []byte) (n int, err error) { return wf(p) }

// OnRead _
func OnRead(done <-chan struct{}, src io.Reader, onRead writerFunc) error {

	// Copy will call the Reader and Writer interface multiple time, in order
	// to copy by chunk (avoiding loading the whole file in memory).
	// I insert the ability to cancel before read time as it is the earliest
	// possible in the call process.
	_, err := io.Copy(onRead, readerFunc(func(p []byte) (int, error) {

		// golang non-blocking channel: https://gobyexample.com/non-blocking-channel-operations
		select {

		// if context has been canceled
		case <-done:
			// stop process and propagate "context canceled" error
			return 0, nil
		default:
			// otherwise just run default io.Reader implementation
			return src.Read(p)
		}
	}))
	return err
}
