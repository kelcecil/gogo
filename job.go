package main

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"io/ioutil"
	"sort"
	"time"
)

type Jobs []Job

func (j Jobs) Len() int {
	return len(j)
}

func (jbs Jobs) Less(i, j int) bool {
	return jbs[i].NextTime.Before(jbs[j].NextTime)
}

func (jbs Jobs) Swap(i, j int) {
	jbs[i], jbs[j] = jbs[j], jbs[i]
}

func (jbs Jobs) Sorted() Jobs {
	sort.Sort(jbs)
	return jbs
}

type Job struct {
	Name     string
	Command  string
	Args     []string
	Schedule string

	NextTime   time.Time            `json:"-"`
	Expression *cronexpr.Expression `json:"-"`
}

type JobsFile struct {
	Jobs []Job
}

func loadJobFile(filename string) (Jobs, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var jobsInfo JobsFile
	err = json.Unmarshal(contents, &jobsInfo)
	if err != nil {
		return nil, err
	}
	return jobsInfo.Jobs, nil
}
