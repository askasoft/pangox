package xjobs

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/askasoft/pango/gwp"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/ref"
	"github.com/askasoft/pango/sqx/sqlx"
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

func (jw *JobWorker[R]) SubmitWork(ctx context.Context, w func()) {
	jw.workerWait.Add(1)
	jw.workerPool.Submit(func() {
		defer func() {
			jw.workerWait.Add(-1)
			if err := recover(); err != nil {
				log.Errorf("Panic in JobWorker (%s): %v", ref.NameOfFunc(w), err)
			}
		}()

		select {
		case <-ctx.Done():
			return
		default:
			w()
		}
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

type IStreamRun[T any] interface {
	FindTarget() (T, error)
	IsStepLimited() bool
	StreamHandle(ctx context.Context, a T) error
}

func StreamRun[T any](ctx context.Context, sr IStreamRun[T]) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if sr.IsStepLimited() {
			return xjm.ErrJobComplete
		}

		a, err := sr.FindTarget()
		if err != nil {
			if errors.Is(err, sqlx.ErrNoRows) {
				return xjm.ErrJobComplete
			}
			return err
		}

		if err = sr.StreamHandle(ctx, a); err != nil {
			return err
		}
	}
}

type ISubmitRun[T any, R any] interface {
	FindTarget() (T, error)
	IsStepLimited() bool

	WorkerPool() *gwp.WorkerPool
	ResultChan() chan R
	WaitAndProcessResults(func(R) error) error

	ProcessResult(r R) error
	SubmitHandle(ctx context.Context, a T)
}

func SubmitRun[T any, R any](ctx context.Context, sr ISubmitRun[T, R]) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if sr.IsStepLimited() {
			return xjm.ErrJobComplete
		}

		a, err := sr.FindTarget()
		if err != nil {
			if errors.Is(err, sqlx.ErrNoRows) {
				return xjm.ErrJobComplete
			}
			return err
		}

		if err := submitTarget(ctx, a, sr); err != nil {
			return err
		}
	}
}

func submitTarget[T any, R any](ctx context.Context, a T, sr ISubmitRun[T, R]) error {
	for {
		select {
		case r := <-sr.ResultChan():
			if err := sr.ProcessResult(r); err != nil {
				return err
			}
		default:
			sr.SubmitHandle(ctx, a)
			return nil
		}
	}
}

type IStreamSubmitRun[T any, R any] interface {
	Running() (context.Context, context.CancelCauseFunc)

	IStreamRun[T]
	ISubmitRun[T, R]
}

func StreamOrSubmitRun[T any, R any](ssr IStreamSubmitRun[T, R]) (err error) {
	ctx, cancel := ssr.Running()
	defer cancel(nil)

	if ssr.WorkerPool() == nil {
		err = StreamRun(ctx, ssr)
		cancel(err)
	} else {
		err = SubmitRun(ctx, ssr)
		if errors.Is(err, xjm.ErrJobComplete) {
			if er := ssr.WaitAndProcessResults(ssr.ProcessResult); er != nil { //nolint: staticcheck
				err = er
			}
			cancel(err)
		} else {
			cancel(err)
			_ = ssr.WaitAndProcessResults(ssr.ProcessResult) //nolint: staticcheck
		}
	}

	return xerrs.ContextCause(ctx, err)
}
