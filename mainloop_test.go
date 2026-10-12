package terminal

import (
	"bytes"
	"fmt"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// mainLoop stands in for the main goroutine of Fyne's desktop driver: fyne.Do and fyne.DoAndWait
// queue functions that one goroutine runs in order. Fyne's test driver runs them on the calling
// goroutine instead. Then the goroutine reading the output and the timers of the bell and of
// blinking text run widget code at the same time as each other and as the test, which the desktop
// driver never does, and the race detector cannot see code that runs on the wrong goroutine.
type mainLoop struct {
	fyne.Driver

	mu     sync.Mutex
	queue  []func()
	wake   chan struct{}
	loopID atomic.Uint64
	panics chan string
}

type mainLoopApp struct {
	fyne.App
	loop *mainLoop
}

func (a *mainLoopApp) Driver() fyne.Driver {
	return a.loop
}

// useMainLoop makes fyne.Do and fyne.DoAndWait run on a main loop until tb ends.
func useMainLoop(tb testing.TB) *mainLoop {
	prev := fyne.CurrentApp()
	app := test.NewApp()
	loop := &mainLoop{Driver: app.Driver(), wake: make(chan struct{}, 1), panics: make(chan string, 1)}
	fyne.SetCurrentApp(&mainLoopApp{App: app, loop: loop})
	go loop.run()
	tb.Cleanup(func() {
		time.Sleep(400 * time.Millisecond) // a bell clears itself after 300 ms
		loop.do(func() {})
		fyne.SetCurrentApp(prev)
	})
	return loop
}

// DoFromGoroutine queues f for the main loop and, if asked, waits for it to have run.
func (l *mainLoop) DoFromGoroutine(f func(), wait bool) {
	switch {
	case wait && l.onLoop():
		f() // the desktop driver warns and does not wait; waiting here would never end
	case wait:
		done := make(chan struct{})
		l.push(func() {
			defer close(done)
			f()
		})
		<-done
	default:
		l.push(f)
	}
}

// do runs f on the main loop and waits for it.
func (l *mainLoop) do(f func()) {
	l.DoFromGoroutine(f, true)
}

// flush waits until the main loop has run everything queued, including what that queued.
func (l *mainLoop) flush() {
	for {
		empty := false
		l.do(func() {
			l.mu.Lock()
			empty = len(l.queue) == 0
			l.mu.Unlock()
		})
		if empty {
			return
		}
	}
}

// panicked returns the panic of a function the main loop ran since the last call, if any.
func (l *mainLoop) panicked() string {
	select {
	case msg := <-l.panics:
		return msg
	default:
		return ""
	}
}

func (l *mainLoop) push(f func()) {
	l.mu.Lock()
	l.queue = append(l.queue, f)
	l.mu.Unlock()
	select {
	case l.wake <- struct{}{}:
	default:
	}
}

func (l *mainLoop) run() {
	l.loopID.Store(goroutineID())
	for range l.wake {
		for {
			l.mu.Lock()
			if len(l.queue) == 0 {
				l.mu.Unlock()
				break
			}
			f := l.queue[0]
			l.queue[0] = nil
			l.queue = l.queue[1:]
			l.mu.Unlock()
			l.call(f)
		}
	}
}

func (l *mainLoop) call(f func()) {
	defer func() {
		if r := recover(); r != nil {
			select {
			case l.panics <- fmt.Sprintf("panic on the main goroutine: %v\n%s", r, debug.Stack()):
			default:
			}
		}
	}()
	f()
}

func (l *mainLoop) onLoop() bool {
	return goroutineID() == l.loopID.Load()
}

// goroutineID returns the number the runtime gives the calling goroutine.
func goroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	fields := bytes.Fields(buf[:n]) // "goroutine 123 [running]: ..."
	if len(fields) < 2 {
		return 0
	}
	id, _ := strconv.ParseUint(string(fields[1]), 10, 64)
	return id
}
