package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gorhill/cronexpr"
	"os"
	"os/exec"
	"time"
)

func main() {
	jobs, err := initializeJobs()
	if err != nil {
		fmt.Println("initializeJobs: " + err.Error())
	}
	fmt.Println(jobs)
	for {
		timeUntilNextJob, _ := prepExecution(jobs)
		time.Sleep(timeUntilNextJob)
	}
}

func waitForProcessCompletion(command string, args []string) {
	execute := exec.Command(command, args...)
	done := make(chan error, 1)
	execute.Stdout = os.Stdout
	execute.Stderr = os.Stderr
	execute.Start()
	go func() {
		done <- execute.Wait()
	}()
	for {
		select {
		case err := <-done:
			if err != nil {
				fmt.Println("Process finished with error: " + err.Error())
				return
			} else {
				fmt.Println("Process finished successfully.")
				return
			}
		}
	}
}

func prepExecution(jbs Jobs) (time.Duration, error) {
	jobs := jbs.Sorted()
	if time.Now().Before(jobs[0].NextTime) {
		return jobs[0].NextTime.Sub(time.Now()), nil
	}
	fmt.Println("Running job " + jobs[0].Name)
	waitForProcessCompletion(jobs[0].Command, jobs[0].Args)
	jobs[0].NextTime = jobs[0].Expression.Next(time.Now())
	return jobs.Sorted()[0].NextTime.Sub(time.Now()), nil
}

func initializeJobs() (Jobs, error) {
	file := flag.String("jobfile", "", "Path to json file containing jobs.")
	flag.Parse()
	if !flag.Parsed() || file == nil {
		return nil, errors.New("Failed to parse filename for jobs.")
	}
	jbs, err := loadJobFile(*file)
	if err != nil {
		return nil, err
	}
	for i := range jbs {
		jbs[i].Expression = cronexpr.MustParse(jbs[i].Schedule)
	}
	return jbs, nil
}
