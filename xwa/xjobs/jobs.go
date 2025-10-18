package xjobs

import (
	"fmt"
	"strings"
	"sync"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/cog/treemap"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox/xjm"
)

type Jobs struct {
	jobs []*xjm.Job
}

func (js *Jobs) Count() int {
	return len(js.jobs)
}

func (js *Jobs) AddJob(job *xjm.Job) {
	js.jobs = append(js.jobs, job)
}

func (js *Jobs) DelJob(job *xjm.Job) {
	js.jobs = asg.DeleteFunc(js.jobs, func(j *xjm.Job) bool { return j.ID == job.ID })
}

type JobsMap struct {
	mu sync.Mutex
	tm *treemap.TreeMap[string, *Jobs]
}

func NewJobsMap() *JobsMap {
	return &JobsMap{tm: treemap.NewTreeMap[string, *Jobs](strings.Compare)}
}

func (jm *JobsMap) Total() (total int) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	for it := jm.tm.Iterator(); it.Next(); {
		total += it.Value().Count()
	}
	return
}

func (jm *JobsMap) Count(key string) int {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if js, ok := jm.tm.Get(key); ok {
		return js.Count()
	}
	return 0
}

func (jm *JobsMap) AddJob(key string, job *xjm.Job) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	js, ok := jm.tm.Get(key)
	if !ok {
		js = &Jobs{}
		jm.tm.Set(key, js)
	}

	js.AddJob(job)
}

func (jm *JobsMap) DelJob(key string, job *xjm.Job) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if js, ok := jm.tm.Get(key); ok {
		js.DelJob(job)
	}
}

func (jm *JobsMap) Clean() {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if jm.tm.Len() > 0 {
		for it := jm.tm.Iterator(); it.Next(); {
			cnt := it.Value().Count()
			if cnt == 0 {
				it.Remove() // remove empty jobs
			}
		}
	}
}

func (jm *JobsMap) Stats() (int, string) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	total := 0
	for it := jm.tm.Iterator(); it.Next(); {
		total += it.Value().Count()
	}

	if total == 0 {
		return 0, ""
	}

	sb := &str.Builder{}

	_, _ = iox.RepeatWrite(sb, []byte{'-'}, 80)
	sb.WriteByte('\n')

	for it := jm.tm.Iterator(); it.Next(); {
		cnt := it.Value().Count()
		if cnt == 0 {
			continue
		}

		fmt.Fprintf(sb, "%32s: ", str.IfEmpty(it.Key(), "_"))
		for i, job := range it.Value().jobs {
			if i > 0 {
				fmt.Fprintf(sb, "%34s", " ")
			}
			fmt.Fprintf(sb, "%d. %s#%d\n", i+1, job.Name, job.ID)
		}
	}
	_, _ = iox.RepeatWrite(sb, []byte{'-'}, 80)

	return total, sb.String()
}
