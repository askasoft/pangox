package xjobs

import (
	"fmt"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/gog"
)

type JobState struct {
	Step    int `json:"step,omitempty"`
	Count   int `json:"count,omitempty"`
	Total   int `json:"total,omitempty"`
	Exists  int `json:"exists,omitempty"`
	Skipped int `json:"skipped,omitempty"`
	Success int `json:"success,omitempty"`
	Failure int `json:"failure,omitempty"`
	Warning int `json:"warning,omitempty"`
}

func (js *JobState) GetStep() int {
	return js.Step
}

func (js *JobState) SetStep(step int) {
	js.Step = step
}

func (js *JobState) IsStepExceeded() bool {
	return js.Total > 0 && js.Step >= js.Total
}

func (js *JobState) IncSkipped() {
	js.Count++
	js.Skipped++
}

func (js *JobState) IncSuccess() {
	js.Count++
	js.Success++
}

func (js *JobState) IncFailure() {
	js.Count++
	js.Failure++
}

func (js *JobState) Progress() string {
	if js.Total > 0 {
		return fmt.Sprintf("[%d/%d]", js.Count, js.Total)
	}
	if js.Count > 0 {
		return fmt.Sprintf("[%d/%d]", js.Count, js.Step)
	}
	if js.Step > 0 {
		return fmt.Sprintf("[%d]", js.Step)
	}
	return ""
}

func (js *JobState) Counts() string {
	return fmt.Sprintf("[%d/%d] (-%d|+%d|!%d)", js.Step, js.Total, js.Skipped, js.Success, js.Failure)
}

func (js *JobState) State() JobState {
	return *js
}

type JobStateLx struct {
	JobState
	Limit int `json:"limit,omitempty"`
}

func (js *JobStateLx) SetTotalLimit(total, limit int) {
	js.Total = total
	js.Limit = gog.If(total > 0 && limit > total, total, limit)
}

func (js *JobStateLx) IsCompleted() bool {
	return js.IsStepLimited()
}

func (js *JobStateLx) IsStepLimited() bool {
	return js.Limit > 0 && js.Step >= js.Limit
}

func (js *JobStateLx) Progress() string {
	if js.Limit > 0 {
		return fmt.Sprintf("[%d/%d]", js.Count, js.Limit)
	}
	return js.JobState.Progress()
}

func (js *JobStateLx) Counts() string {
	return fmt.Sprintf("[%d/%d/%d] (-%d|+%d|!%d)", js.Step, js.Limit, js.Total, js.Skipped, js.Success, js.Failure)
}

type JobStateSx struct {
	JobStateLx
}

func (js *JobStateSix) IsCompleted() bool {
	return js.IsSuccessLimited()
}

func (js *JobStateSx) IsSuccessLimited() bool {
	return js.Limit > 0 && js.Success >= js.Limit
}

func (js *JobStateSx) IncSkipped() {
	js.Skipped++
}

func (js *JobStateSx) IncFailure() {
	js.Failure++
}

func (js *JobStateSx) Progress() string {
	if js.Limit > 0 {
		return fmt.Sprintf("[%d/%d]", js.Success, js.Limit)
	}
	if js.Total > 0 {
		return fmt.Sprintf("[%d/%d]", js.Success, js.Total)
	}
	if js.Success > 0 {
		return fmt.Sprintf("[%d/%d]", js.Success, js.Step)
	}
	if js.Step > 0 {
		return fmt.Sprintf("[%d]", js.Step)
	}
	return ""
}

type JobLastID struct {
	LastID int64 `json:"last_id,omitempty"`
}

// GetLastID get last id
func (jl *JobLastID) GetLastID() int64 {
	return jl.LastID
}

// SetLastID set last id
func (jl *JobLastID) SetLastID(id int64) {
	jl.LastID = id
}

type JobStateLix struct {
	JobStateLx
	JobLastID
}

type JobStateSix struct {
	JobStateSx
	JobLastID
}

type JobStateLixs struct {
	JobStateLix
	LastIDs []int64 `json:"last_ids,omitempty"`
}

func (jse *JobStateLixs) JobState() IState {
	return jse
}

// InitLastID keep last minimum id
func (jse *JobStateLixs) InitLastID() {
	if len(jse.LastIDs) > 0 {
		jse.LastID = asg.Min(jse.LastIDs)
		jse.LastIDs = jse.LastIDs[:0]
	}
}

// DelLastID remove id from LastIDs
func (jse *JobStateLixs) DelLastID(id int64) {
	jse.LastIDs = asg.DeleteEqual(jse.LastIDs, id)
}

// AddLastID increment step and add id to LastIDs
func (jse *JobStateLixs) AddLastID(id int64) {
	jse.Step++
	jse.LastID = id
	jse.LastIDs = append(jse.LastIDs, id)
}

// AddFailureID increment failure and remove id from LastIDs
func (jse *JobStateLixs) AddFailureID(id int64) {
	jse.IncFailure()
	jse.DelLastID(id)
}

// AddSuccessID increment success and remove id from LastIDs
func (jse *JobStateLixs) AddSuccessID(id int64) {
	jse.IncSuccess()
	jse.DelLastID(id)
}

// AddSkippedID increment skipped and remove id from LastIDs
func (jse *JobStateLixs) AddSkippedID(id int64) {
	jse.IncSkipped()
	jse.DelLastID(id)
}

type IJobStater interface {
	GetStep() int
	SetTotalLimit(total, limit int)
	CountTargets() (cnt int, err error)
	SaveState() error
}

func InitState(js IJobStater, limit int) error {
	if js.GetStep() == 0 {
		total, err := js.CountTargets()
		if err != nil {
			return err
		}

		js.SetTotalLimit(int(total), limit)
		return js.SaveState()
	}
	return nil
}
