package xjobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/askasoft/pango/log"
	"github.com/askasoft/pangox/xjm"
	"github.com/askasoft/pangox/xwa/xerrs"
)

var (
	ErrItemSkip = errors.New("item skip")
)

type IState interface {
	State() JobState
}

type FailedItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Error string `json:"error"`
}

func (si *FailedItem) Quoted() string {
	return fmt.Sprintf("%d\t%q\t%q\n", si.ID, si.Title, si.Error)
}

func (si *FailedItem) String() string {
	return fmt.Sprintf("#%d %s - %s", si.ID, si.Title, si.Error)
}

type IJobRunner interface {
	Checkout() error
	Run() error
	Done(error)
}

func RunJob(run IJobRunner) {
	run.Done(safeRun(run))
}

func safeRun(run IJobRunner) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
		}
	}()

	err = run.Checkout()
	if err != nil {
		return
	}

	err = run.Run()
	return
}

type JobRunner struct {
	xjc xjm.JobChainer

	*xjm.JobRunner
	ChainArg

	JobChainContinue func(next *JobRunState) error
}

func NewJobRunner(job *xjm.Job, xjc xjm.JobChainer, jmr xjm.JobManager, logger ...log.Logger) *JobRunner {
	jr := &JobRunner{
		xjc:       xjc,
		JobRunner: xjm.NewJobRunner(job, jmr, logger...),
	}

	return jr
}

func (jr *JobRunner) XJC() xjm.JobChainer {
	return jr.xjc
}

func (jr *JobRunner) AddFailedItem(id int64, title, reason string) {
	si := FailedItem{
		ID:    id,
		Title: title,
		Error: reason,
	}
	_ = jr.AddResult(si.Quoted())
}

func (jr *JobRunner) Checkout() error {
	if err := jr.JobRunner.Checkout(); err != nil {
		return err
	}

	return jr.jobChainCheckout()
}

func (jr *JobRunner) Running() (context.Context, context.CancelCauseFunc) {
	ctx, cancel := context.WithCancelCause(context.TODO())
	go func() {
		if err := jr.JobRunner.Running(ctx, time.Second, time.Minute); err != nil {
			cancel(err)
		}
	}()
	return ctx, cancel
}

func (jr *JobRunner) SetState(state IState) error {
	if err := jr.JobRunner.SetState(xjm.MustEncode(state)); err != nil {
		return err
	}

	return jr.jobChainSetState(state)
}

func (jr *JobRunner) Abort(reason string) {
	logger := jr.Log().GetLogger("JOB")

	if err := jr.JobRunner.Abort(reason); err != nil {
		if !errors.Is(err, xjm.ErrJobMissing) {
			logger.Error(err)
		}
	}

	// Abort job chain
	if err := jr.jobChainAbort(reason); err != nil {
		logger.Error(err)
	}

	logger.Warn("ABORTED.")
}

func (jr *JobRunner) Finish() {
	logger := jr.Log().GetLogger("JOB")

	if err := jr.JobRunner.Finish(); err != nil {
		if !errors.Is(err, xjm.ErrJobMissing) {
			logger.Error(err)
		}
		jr.Abort(err.Error())
		return
	}

	// Continue job chain
	if err := jr.jobChainContinue(); err != nil {
		logger.Error(err)
	}

	logger.Info("DONE.")
}

func (jr *JobRunner) Done(err error) {
	logger := jr.Log().GetLogger("JOB")

	defer jr.Log().Close()

	if errors.Is(err, xjm.ErrJobCheckout) {
		// do nothing, just log it
		logger.Warn(err)
		return
	}

	if err == nil || errors.Is(err, xjm.ErrJobComplete) {
		jr.Finish()
		return
	}

	if errors.Is(err, xjm.ErrJobMissing) {
		// job is missing, unable to do anything, just log error
		logger.Error(err)
		return
	}

	if errors.Is(err, xjm.ErrJobAborted) || errors.Is(err, xjm.ErrJobCanceled) || errors.Is(err, xjm.ErrJobPin) {
		job, err := jr.GetJob()
		if err != nil {
			logger.Error(err)
			return
		}

		switch job.Status {
		case xjm.JobStatusAborted:
			// NOTE:
			// It's necessary to call jobChainAbort() again.
			// The jobChainCheckout()/jobChainContinue() method may update job chain status to 'R' to a aborted job chain.
			if err := jr.jobChainAbort(job.Error); err != nil {
				logger.Error(err)
			}

			logger.Warn("ABORTED.")
			return
		case xjm.JobStatusCanceled:
			// NOTE:
			// It's necessary to call jobChainCancel() again.
			// The jobChainCheckout()/jobChainContinue() method may update job chain status to 'R' to a aborted job chain.
			if err := jr.jobChainCancel(job.Error); err != nil {
				logger.Error(err)
			}

			logger.Warn("CANCELED.")
			return
		default:
			logger.Errorf("Illegal job status %s#%d (%d): %s", jr.JobName(), jr.JobID(), jr.RunnerID(), xjm.MustEncode(job))
			return
		}
	}

	if xerrs.IsClientError(err) {
		logger.Warn(err)
	} else {
		logger.Error(err)
	}

	jr.Abort(err.Error())
}

// ---------------------------------------------------------------------
func (jr *JobRunner) jobChainCheckout() error {
	if jr.ChainID() == 0 {
		return nil
	}

	return JobCheckoutChain(jr.xjc, jr.ChainID(), jr.JobID(), jr.JobName())
}

func (jr *JobRunner) jobChainSetState(state IState) error {
	if jr.ChainID() == 0 {
		return nil
	}

	return JobFindAndUpdateChainState(jr.xjc, jr.ChainID(), jr.JobID(), jr.JobName(), state.State())
}

func (jr *JobRunner) jobChainAbort(reason string) error {
	if jr.ChainID() == 0 {
		return nil
	}

	return JobFindAndAbortChain(jr.xjc, jr.ChainID(), jr.JobID(), jr.JobName(), reason)
}

func (jr *JobRunner) jobChainCancel(reason string) error {
	if jr.ChainID() == 0 {
		return nil
	}

	return JobFindAndCancelChain(jr.xjc, jr.ChainID(), jr.JobID(), jr.JobName(), reason)
}

func (jr *JobRunner) jobChainContinue() error {
	if jr.ChainID() == 0 {
		return nil
	}

	next, err := JobFindAndContinueChain(jr.xjc, jr.ChainID(), jr.JobID(), jr.JobName())
	if err != nil {
		return err
	}
	if next != nil {
		return jr.JobChainContinue(next)
	}
	return nil
}
