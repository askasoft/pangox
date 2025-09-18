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

type JobStateLix struct {
	JobStateLx
	LastID int64 `json:"last_id,omitempty"`
}

type JobStateSix struct {
	JobStateSx
	LastID int64 `json:"last_id,omitempty"`
}

type JobStateLixs struct {
	JobStateLx
	LastIDs []int64 `json:"last_ids,omitempty"`
}

// InitLastID remain minimum id only
func (jse *JobStateLixs) InitLastID() {
	if len(jse.LastIDs) > 0 {
		jse.LastIDs[0] = asg.Min(jse.LastIDs)
		jse.LastIDs = jse.LastIDs[:1]
	}
}

func (jse *JobStateLixs) AddLastID(id int64) {
	jse.Step++
	jse.LastIDs = append(jse.LastIDs, id)
}

func (jse *JobStateLixs) AddFailureID(id int64) {
	jse.Count++
	jse.Failure++
	jse.LastIDs = asg.DeleteEqual(jse.LastIDs, id)
}

func (jse *JobStateLixs) AddSuccessID(id int64) {
	jse.Count++
	jse.Success++
	jse.LastIDs = asg.DeleteEqual(jse.LastIDs, id)
}

func (jse *JobStateLixs) AddSkippedID(id int64) {
	jse.Count++
	jse.Skipped++
	jse.LastIDs = asg.DeleteEqual(jse.LastIDs, id)
}
