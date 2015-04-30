package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorhill/cronexpr"
	reaper "github.com/ramr/go-reaper"
)

func main() {
	go reaper.Reap()
	signals := createSignalChannel()
	jobs, err := initializeJobs()
	if err != nil {
		fmt.Println("initializeJobs: " + err.Error())
		os.Exit(1)
	}
	fmt.Println(jobs)
	for {
		timeUntilNextJob := checkAndRun(jobs, signals)
		done := make(chan bool, 1)
		go func() {
			for {
				select {
				case <-signals:
					os.Exit(0)
				case <-done:
					return
				}
			}
		}()
		time.Sleep(timeUntilNextJob)
		done <- true
	}
}

func createSignalChannel() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	return c
}

func waitForProcessCompletion(signals chan os.Signal, command string, args []string) {
	execute := exec.Command(command, args...)
	done := make(chan error, 1)
	execute.Stdout = os.Stdout
	execute.Stderr = os.Stderr
	execute.Start()
	proc := execute.Process
	exit := false
	go func() {
		done <- execute.Wait()
	}()
	for {
		select {
		case err := <-done:
			if exit {
				os.Exit(0)
			}
			if err != nil {
				fmt.Println("Process finished with error: " + err.Error())
				return
			} else {
				fmt.Println("Process finished successfully.")
				return
			}
		case signal := <-signals:
			proc.Signal(signal)
			exit = true
		}
	}
}

func checkAndRun(jbs Jobs, signals chan os.Signal) time.Duration {
	jobs := jbs.Sorted()
	if time.Now().Before(jobs[0].NextTime) {
		return jobs[0].NextTime.Sub(time.Now())
	}
	fmt.Println("Running job " + jobs[0].Name)
	waitForProcessCompletion(signals, jobs[0].Command, jobs[0].Args)
	jobs[0].NextTime = jobs[0].Expression.Next(time.Now())
	jobs = jobs.Sorted()
	fmt.Println("The next job \"" + jobs[0].Name + "\" will run at time: " + jobs[0].NextTime.String())
	timeUntilExecution := math.Ceil(jobs[0].NextTime.Sub(time.Now()).Minutes())
	fmt.Printf("Next execution in ~%.0f minutes\n", timeUntilExecution)
	return jobs[0].NextTime.Sub(time.Now())
}

func initializeJobs() (Jobs, error) {
	file := flag.String("jobfile", "", "Path to json file containing jobs.")
	flag.Parse()
	if !flag.Parsed() || *file == "" {
		return nil, errors.New("Failed to parse filename for jobs.")
	}
	jbs, err := loadJobFile(*file)
	if err != nil {
		return nil, err
	}
	for i := range jbs {
		jbs[i].Expression = cronexpr.MustParse(jbs[i].Schedule)
		jbs[i].NextTime = jbs[i].Expression.Next(time.Now())
	}
	return jbs, nil
}
