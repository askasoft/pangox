package xjobs

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/askasoft/pango/gwp"
	"github.com/askasoft/pango/ref"
	"github.com/askasoft/pangox/xjm"
	"github.com/askasoft/pangox/xwa/xerrs"
)

type JobWorker[R any] struct {
	workerPool *gwp.WorkerPool
	workerWait atomic.Int32
	resultChan chan R
}

func (jw *JobWorker[R]) WorkerPool() *gwp.WorkerPool {
	return jw.workerPool
}

func (jw *JobWorker[R]) WorkerRunning() int32 {
	return jw.workerWait.Load()
}

func (jw *JobWorker[R]) ResultChan() chan R {
	return jw.resultChan
}

func (jw *JobWorker[R]) SetWorkerPool(wp *gwp.WorkerPool) {
	jw.workerPool = wp
	if wp != nil {
		jw.resultChan = make(chan R, wp.MaxWorks())
	}
}

func (jw *JobWorker[R]) IsConcurrent() bool {
	return jw.workerPool != nil
}

func (jw *JobWorker[R]) SubmitWork(ctx JobContext, w func()) {
	jw.workerWait.Add(1)
	jw.workerPool.Submit(func() {
		defer func() {
			jw.workerWait.Add(-1)
			if r := recover(); r != nil {
				ctx.Cancel(fmt.Errorf("%s: %w", ref.NameOfFunc(w), xerrs.PanicError(r)))
			}
		}()

		w()
	})
}

func (jw *JobWorker[R]) WaitAndProcessResults(fp func(R) error) (err error) {
	timer := time.NewTimer(time.Millisecond * 100)
	defer timer.Stop()

	for {
		select {
		case r, ok := <-jw.resultChan:
			if !ok {
				return
			}
			if er := fp(r); er != nil {
				err = er
			}
		case <-timer.C:
			if jw.WorkerRunning() == 0 {
				close(jw.resultChan)
			} else {
				timer.Reset(time.Millisecond * 100)
			}
		}
	}
}

type IJobRun[T any] interface {
	FindTargets() ([]T, error)
	IsCompleted() bool
	Running() JobContext
}

type IStreamRun[T any] interface {
	IJobRun[T]
	StreamHandle(ctx JobContext, a T) error
}

func StreamRun[T any](sr IStreamRun[T]) (err error) {
	ctx := sr.Running()
	defer ctx.Cancel(nil)

	defer func() {
		if r := recover(); r != nil {
			err = xerrs.PanicError(r)
		}
	}()

	err = streamRun(ctx, sr)
	if err != nil {
		ctx.Cancel(err)
	}

	err = xerrs.ContextCause(ctx, err)
	return
}

func streamRun[T any](ctx JobContext, sr IStreamRun[T]) error {
	for {
		ts, err := sr.FindTargets()
		if err != nil {
			return err
		}

		if len(ts) == 0 {
			return nil
		}

		for _, t := range ts {
			if err = sr.StreamHandle(ctx, t); err != nil {
				return err
			}

			if sr.IsCompleted() {
				return xjm.ErrJobComplete
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
	}
}

type ISubmitRun[T any, R any] interface {
	IJobRun[T]

	WorkerPool() *gwp.WorkerPool
	ResultChan() chan R
	WaitAndProcessResults(func(R) error) error

	ProcessResult(r R) error
	SubmitHandle(ctx JobContext, a T) error
}

func SubmitRun[T any, R any](sr ISubmitRun[T, R]) error {
	ctx := sr.Running()
	defer ctx.Cancel(nil)

	err := submitRun(ctx, sr)
	if err == nil || errors.Is(err, xjm.ErrJobComplete) {
		if er := sr.WaitAndProcessResults(sr.ProcessResult); er != nil {
			err = er
		}
		if err != nil {
			ctx.Cancel(err)
		}
	} else {
		ctx.Cancel(err)
		_ = sr.WaitAndProcessResults(sr.ProcessResult)
	}

	return xerrs.ContextCause(ctx, err)
}

func submitRun[T any, R any](ctx JobContext, sr ISubmitRun[T, R]) error {
	for {
		ts, err := sr.FindTargets()
		if err != nil {
			return err
		}

		if len(ts) == 0 {
			return nil
		}

		for _, t := range ts {
			if err := submitTarget(ctx, t, sr); err != nil {
				return err
			}

			if sr.IsCompleted() {
				return xjm.ErrJobComplete
			}
		}
	}
}

func submitTarget[T any, R any](ctx JobContext, a T, sr ISubmitRun[T, R]) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case r := <-sr.ResultChan():
			if err := sr.ProcessResult(r); err != nil {
				return err
			}
		default:
			return sr.SubmitHandle(ctx, a)
		}
	}
}

type IStreamSubmitRun[T any, R any] interface {
	IStreamRun[T]
	ISubmitRun[T, R]
}

func StreamOrSubmitRun[T any, R any](ssr IStreamSubmitRun[T, R]) error {
	if ssr.WorkerPool() == nil {
		return StreamRun(ssr)
	}
	return SubmitRun(ssr)
}
