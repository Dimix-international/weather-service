package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
)

const httpPort = ":3000"

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")

		w.Write([]byte(city))
	})

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		runCron()
	}()

	go func() {
		defer wg.Done()
		fmt.Println("start server")
		http.ListenAndServe(httpPort, r)
	}()

	wg.Wait()
}

func initJobs(s gocron.Scheduler) ([]gocron.Job, error) {
	j, err := s.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				fmt.Println("hello")
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return []gocron.Job{j}, nil
}

func runCron() {
	s, err := gocron.NewScheduler()
	if err != nil {
		return
	}

	jobs, err := initJobs(s)
	if err != nil {
		return
	}

	fmt.Printf("Started first job id %v\n", jobs[0].ID())

	s.Start()
}
