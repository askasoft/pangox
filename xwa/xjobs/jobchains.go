package xjobs

import (
	"errors"
	"fmt"
	"time"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox/xjm"
)

type IChainArg interface {
	GetChain() (int, bool)
	SetChain(chainSeq int, chainData bool)
}

type ChainArg struct {
	ChainSeq  int  `json:"chain_seq,omitempty" form:"-"`
	ChainData bool `json:"chain_data,omitempty" form:"chain_data"`
}

func (ca *ChainArg) GetChain() (int, bool) {
	return ca.ChainSeq, ca.ChainData
}

func (ca *ChainArg) SetChain(csq int, cdt bool) {
	ca.ChainSeq = csq
	ca.ChainData = cdt
}

func (ca *ChainArg) ShouldChainData() bool {
	return ca.ChainData && ca.ChainSeq > 0
}

type JobRunState struct {
	JID    int64    `json:"jid"`
	Name   string   `json:"name"`
	Status string   `json:"status"`
	Error  string   `json:"error"`
	State  JobState `json:"state"`
}

func JobChainDecodeStates(state string) (states []*JobRunState) {
	xjm.MustDecode(state, &states)
	return
}

func JobChainEncodeStates(states []*JobRunState) string {
	return xjm.MustEncode(states)
}

func JobChainInitStates(jns ...string) []*JobRunState {
	states := make([]*JobRunState, len(jns))
	for i, jn := range jns {
		js := &JobRunState{Name: jn, Status: xjm.JobStatusPending}
		states[i] = js
	}
	return states
}

func JobChainAbort(xjc xjm.JobChainer, tjm xjm.JobManager, jc *xjm.JobChain, reason string) error {
	return jobChainAbortCancel(xjc, tjm, jc, xjm.JobStatusAborted, reason, tjm.AbortJob)
}

func JobChainCancel(xjc xjm.JobChainer, tjm xjm.JobManager, jc *xjm.JobChain, reason string) error {
	return jobChainAbortCancel(xjc, tjm, jc, xjm.JobStatusCanceled, reason, tjm.CancelJob)
}

func jobChainAbortCancel(xjc xjm.JobChainer, tjm xjm.JobManager, jc *xjm.JobChain, status, reason string, funcAbortCancel func(int64, string) error) error {
	states := JobChainDecodeStates(jc.States)
	for _, sta := range states {
		if sta.JID != 0 && asg.Contains(xjm.JobUndoneStatus, sta.Status) {
			if err := funcAbortCancel(sta.JID, reason); err != nil && !errors.Is(err, xjm.ErrJobMissing) {
				return err
			}
			_ = tjm.AddJobLog(sta.JID, time.Now(), xjm.JobLogLevelWarn, reason)
		}
	}
	return xjc.UpdateJobChain(jc.ID, status)
}

func JobFindAndAbortChain(xjc xjm.JobChainer, cid, jid int64, jname, reason string) error {
	jc, err := xjc.GetJobChain(cid)
	if err != nil {
		return err
	}

	ok, err := jobAbortCancelChain(xjc, jc, jid, xjm.JobStatusAborted, reason)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return fmt.Errorf("unable to abort jobchain %s#%d for job %s#%d", jc.Name, jc.ID, jname, jid)
}

func JobFindAndCancelChain(xjc xjm.JobChainer, cid, jid int64, jname, reason string) error {
	jc, err := xjc.GetJobChain(cid)
	if err != nil {
		return err
	}

	ok, err := jobAbortCancelChain(xjc, jc, jid, xjm.JobStatusCanceled, reason)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return fmt.Errorf("unable to cancel jobchain %s#%d for job %s#%d", jc.Name, jc.ID, jname, jid)
}

func jobAbortCancelChain(xjc xjm.JobChainer, jc *xjm.JobChain, jid int64, status, reason string) (bool, error) {
	jcs := str.If(jc.IsDone(), "", status)

	states := JobChainDecodeStates(jc.States)
	for _, sta := range states {
		if sta.JID == jid {
			sta.Status = status
			if reason != "" {
				sta.Error = reason
			}
			state := JobChainEncodeStates(states)
			return true, xjc.UpdateJobChain(jc.ID, jcs, state)
		}
	}
	return false, nil
}

func JobCheckoutChain(xjc xjm.JobChainer, cid, jid int64, jname string) error {
	jc, err := xjc.GetJobChain(cid)
	if err != nil {
		return err
	}

	switch jc.States {
	case xjm.JobStatusAborted:
		return xjm.ErrJobAborted
	case xjm.JobStatusCanceled:
		return xjm.ErrJobCanceled
	case xjm.JobStatusFinished:
		return xjm.ErrJobComplete
	}

	states := JobChainDecodeStates(jc.States)
	for _, sta := range states {
		if sta.Name == jname && (sta.JID == 0 || sta.JID == jid) {
			sta.JID = jid
			sta.Status = xjm.JobStatusRunning
			state := JobChainEncodeStates(states)
			return xjc.UpdateJobChain(jc.ID, xjm.JobStatusRunning, state)
		}
	}
	return fmt.Errorf("unable to checkout jobchain %s#%d for job %s", jc.Name, jc.ID, jname)
}

func JobFindAndUpdateChainState(xjc xjm.JobChainer, cid, jid int64, jname string, state JobState) error {
	jc, err := xjc.GetJobChain(cid)
	if err != nil {
		return err
	}

	states := JobChainDecodeStates(jc.States)
	for _, sta := range states {
		if sta.JID == jid {
			sta.Status = xjm.JobStatusRunning
			sta.State = state
			return xjc.UpdateJobChain(jc.ID, "", JobChainEncodeStates(states))
		}
	}
	return fmt.Errorf("unable to set jobchain state %s#%d for job %s#%d", jc.Name, jc.ID, jname, jid)
}

func JobFindAndContinueChain(xjc xjm.JobChainer, cid, jid int64, jname string) (*JobRunState, error) {
	jc, err := xjc.GetJobChain(cid)
	if err != nil {
		return nil, err
	}

	var curr, next *JobRunState

	states := JobChainDecodeStates(jc.States)
	for i, sta := range states {
		if sta.JID == jid {
			curr = sta
			i++
			if i < len(states) {
				next = states[i]
			}
			break
		}
	}
	if curr == nil {
		return nil, fmt.Errorf("unable to continue jobchain %s#%d for job %s#%d", jc.Name, jc.ID, jname, jid)
	}

	curr.Status = xjm.JobStatusFinished
	status := str.If(next == nil, xjm.JobStatusFinished, xjm.JobStatusRunning)
	state := JobChainEncodeStates(states)

	if jc.IsDone() {
		// do not update already done job chain status
		status = ""
	}
	if err := xjc.UpdateJobChain(jc.ID, status, state); err != nil {
		return nil, err
	}
	if next != nil && status == xjm.JobStatusRunning {
		return next, nil
	}
	return nil, nil
}
