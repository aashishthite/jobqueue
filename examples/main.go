package main

import (
	"fmt"
	"time"

	"github.com/aashishthite/jobqueue"
)

func main() {

	fmt.Println("Hello World!")
	jq := jobqueue.NewJobQueue()

	//Register jobs A through H
	//A depends on B,C & G
	//D is a recurring job
	//E depends on H and A
	jq.Register(&jobqueue.Job{
		Name: "A",
		Deps: []string{"B", "C", "G"},
		Handler: func() error {
			fmt.Println("Doing A")
			time.Sleep(1 * time.Second)
			return nil
		},
	})

	jq.Register(&jobqueue.Job{
		Name: "B",
		Deps: []string{},
		Handler: func() error {
			fmt.Println("Doing B")
			time.Sleep(1 * time.Second)
			return nil
		},
	})

	jq.Register(&jobqueue.Job{
		Name: "C",
		Deps: []string{},
		Handler: func() error {
			fmt.Println("Doing C")
			time.Sleep(1 * time.Second)
			return nil
		},
	})

	jq.Register(&jobqueue.Job{
		Name:      "D",
		Deps:      []string{},
		Recurring: true,
		Handler: func() error {
			fmt.Println("Doing D")
			time.Sleep(1 * time.Second)
			return nil
		},
	})

	jq.Register(&jobqueue.Job{
		Name: "E",
		Deps: []string{"H", "A"},
		Handler: func() error {
			fmt.Println("Doing E")
			time.Sleep(1 * time.Second)
			return nil
		},
	})
	jq.Register(&jobqueue.Job{
		Name: "F",
		Deps: []string{},
		Handler: func() error {
			fmt.Println("Doing F")
			time.Sleep(1 * time.Second)
			return nil
		},
	})
	jq.Register(&jobqueue.Job{
		Name: "G",
		Deps: []string{},
		Handler: func() error {
			fmt.Println("Doing G")
			time.Sleep(1 * time.Second)
			return nil
		},
	})
	jq.Register(&jobqueue.Job{
		Name: "H",
		Deps: []string{},
		Handler: func() error {
			fmt.Println("Doing H")
			time.Sleep(1 * time.Second)
			return nil
		},
	})

	jq.Start()
	time.Sleep(10 * time.Second)
	jq.Stop()
}
