package framework

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	harness "github.com/dlespiau/kube-test-harness"
	appsv1 "k8s.io/api/apps/v1"
)

type Test struct {
	harness.Test

	operatorRunning bool
	harness         *Harness
}

func (t *Test) Setup() *Test {
	_ = t.Test.Setup()
	return t
}

func (t *Test) StartOperator() error {
	if t.operatorRunning == true {
		return fmt.Errorf("operator already started")
	}
	err := startOperator(t.Namespace, t.harness.Options, t.harness.Sinks)
	if err == nil {
		t.operatorRunning = true
	}
	return err
}

func (t *Test) StopOperator() {
	if t.operatorRunning {
		stopOperator(t.harness.Options, t.harness.Sinks)
		t.operatorRunning = false
	}
}

func (t *Test) DeleteDeployment(d *appsv1.Deployment, timeout time.Duration) {
	t.Test.DeleteDeployment(d)
	t.Test.WaitForDeploymentDeleted(d, timeout)
}

func (t *Test) Close() {
	// If panicking, let Test.Close() do its thing only and keep the operator running
	defer t.Test.Close()
	if r := recover(); r != nil {
		panic(r)
	} else {
		t.StopOperator()
	}
}

// One-shot channel for single error, safe to send() and close() concurrently
// or many times, but only first operation succeeds (others fail silently)
type chanError struct {
	data    chan error
	closing int32
}

func newChanError() *chanError {
	return &chanError{
		make(chan error, 1),
		0,
	}
}

func (c *chanError) send(err error) {
	if atomic.CompareAndSwapInt32(&c.closing, 0, 1) {
		c.data <- err
		// NOTE: we may send a nil error here, this is supported behaviour
		// NOTE: it is recipient responsibility to call close(c.data)
	}
}

func (c *chanError) close() {
	if atomic.CompareAndSwapInt32(&c.closing, 0, 1) {
		close(c.data)
	}
}

type asynchronousHandler struct {
	delay time.Duration
	result error
}

// The logic is not obvious, so some explanation follows:
// when we start the operator process for testing, it is possible that the process
// will fail right away, because of some early-manifest bug. It might also
// for some reason exit prematurely, without reporting an error.
// To discover when this happens, we will wait for the process to return (possibly
// with an error), and will also start a timer to close the channel for the status
// when operatorStartDelay has elapsed.
// If we have received anything on the channel (before it closed), it means that
// the program completed; otherwise we consider it running.
func(h* asynchronousHandler) Handle(cmd *exec.Cmd) {
	channel := newChanError()
	go func() {
		err := cmd.Wait()
		// will only succeed to send an error if completed before operatorStartDelay
		channel.send(err)
	}()
	go func() {
		time.Sleep(h.delay)
		// safe no-op if the channel closed earlier
		channel.close()
	}()

	// wait on channel.data will complete when either happens:
	// * channel.send(err), i.e. program completed, possibly with error
	// * channel.close(), i.e. Sleep(operatorStartDelay) elapsed
	if err, ok := <-channel.data; ok {
		if err == nil {
			// This will happen if channel.send(nil) was called above, which
			// indicates that the make target to start the operator has
			// exited prematurely, but with success status. This indicates
			// an unknown error, since we expect "make start operator" to block
			// while the operator is running
			err = fmt.Errorf("operator not running")
		}
		h.result = err
		// NOTE: it is recipient responsibility to call close(c.data)
		close(channel.data)
	}
}

func startOperator(namespace string, options Options, sinks Sinks) error {
	makefile := options.Makefile
	makedir := options.MakeDir
	_ = os.Setenv("TEST_OPERATOR_NS", namespace)
	args := []string{"make", "-s", "-f", makefile, "-C", makedir, options.operatorStart()}
	log.Printf("Starting %v ...", args)
	// let's use sinks.Operator as Stdout for operator output
	handler := asynchronousHandler{options.OperatorStartDelay, nil}
	if err := start(&handler, sinks.Operator,  nil, args); err != nil {
		_ = os.Unsetenv("TEST_OPERATOR_NS")
		return err
	}
	if handler.result != nil {
		_ = os.Unsetenv("TEST_OPERATOR_NS")
	}
	return handler.result
}

func stopOperator(options Options, sinks Sinks) {
	makefile := options.Makefile
	makedir := options.MakeDir
	args := []string{"make", "-s", "-f", makefile, "-C", makedir, options.operatorStop()}
	log.Printf("Running %v ...", args)
	_ = run(sinks.Stdout, sinks.Stderr, args)
	_ = os.Unsetenv("TEST_OPERATOR_NS")
	log.Print("... done")
}
