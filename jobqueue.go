package jobqueue

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

const (
	_ int32 = iota
	statusProcessing
	statusPending
	statusCompleted
)

const (
	_ int32 = iota
	queueStopped
	queueStarting
	queueRunning
	queueStopping
)

type JobQueue struct {
	numWorkers  int
	configs     map[string]*Job
	status      *int32
	workersWg   sync.WaitGroup
	fetcherWg   sync.WaitGroup
	nextJob     chan string
	quit        chan bool
	maxAttempts int
	backoff     int
}

type Function func() error

type Job struct {
	Name          string
	Deps          []string
	Recurring     bool
	status        *int32
	Handler       Function
	attempts      int
	schedulableAt int64
}

func NewJobQueue() *JobQueue {
	status := queueStopped

	jq := JobQueue{
		configs:     make(map[string]*Job, 10),
		status:      &status,
		numWorkers:  5,
		maxAttempts: 3,
		backoff:     1,
	}
	return &jq
}

func (jq *JobQueue) Register(job *Job) error {
	if _, ok := jq.configs[job.Name]; ok {
		return fmt.Errorf("job %s: already registered", job.Name)
	}

	if st := atomic.LoadInt32(jq.status); st != queueStopped {
		return fmt.Errorf("job %s: unable to register once queue has started", job.Name)
	}
	job.schedulableAt = time.Now().Unix()
	jq.configs[job.Name] = job
	status := statusPending
	job.status = &status
	fmt.Println("Registering", job.Name)
	return nil
}

func (jq *JobQueue) fetcher() {
	defer jq.fetcherWg.Done()

	for {
		if st := atomic.LoadInt32(jq.status); st == queueStopping {
			return
		}
		var jobName string

		for k, v := range jq.configs {
			if v.schedulableAt <= time.Now().Unix() && v.attempts <= jq.maxAttempts {
				if st := atomic.LoadInt32(v.status); st != statusPending {
					continue
				}

				dependanciesDone := true
				for _, dep := range v.Deps {
					if depJob, ok := jq.configs[dep]; ok {
						depSt := atomic.LoadInt32(depJob.status)
						dependanciesDone = dependanciesDone && depSt == statusCompleted
					} else {
						dependanciesDone = false
						break
					}
				}

				if dependanciesDone {
					jobName = k
					atomic.StoreInt32(v.status, statusProcessing)
					break
				}
			}
		}

		if jobName == "" {
			time.Sleep(1 * time.Second)
			continue
		}
		//fmt.Println(jobName)
		select {
		case jq.nextJob <- jobName:
		case <-jq.quit:
		}
	}

}

func (jq *JobQueue) worker(workerIndex int) {
	defer jq.workersWg.Done()
	for {
		if st := atomic.LoadInt32(jq.status); st == queueStopping {
			return
		}

		jobName := <-jq.nextJob

		job, ok := jq.configs[jobName]
		if ok {

			jobErr := jq.safeProcess(job)

			if jobErr != nil {
				job.attempts++
				job.schedulableAt = time.Now().Unix() + int64(jq.backoff*job.attempts)
				atomic.StoreInt32(job.status, statusPending)
				continue
			}
			atomic.StoreInt32(job.status, statusCompleted)
			if job.Recurring {
				atomic.StoreInt32(job.status, statusPending)
				job.schedulableAt = time.Now().Unix() + int64(3) // do every 3 seconds
			}
		}

	}
}

func (jq *JobQueue) Start() error {
	if !atomic.CompareAndSwapInt32(jq.status, queueStopped, queueStarting) {
		return errors.New("can only start a queue which is stopped")
	}

	jq.nextJob = make(chan string)
	jq.quit = make(chan bool, 1)

	jq.fetcherWg.Add(1)
	go jq.fetcher()

	jq.workersWg.Add(jq.numWorkers)
	for i := 0; i < jq.numWorkers; i++ {
		go jq.worker(i)
	}

	atomic.StoreInt32(jq.status, queueRunning)
	return nil
}

func (jq *JobQueue) Stop() error {
	if !atomic.CompareAndSwapInt32(jq.status, queueRunning, queueStopping) {
		return errors.New("unable to stop")
	}

	jq.quit <- true
	jq.fetcherWg.Wait()
	close(jq.nextJob)

	jq.workersWg.Wait()
	atomic.StoreInt32(jq.status, queueStopped)
	jq.nextJob = nil
	jq.quit = nil

	return nil
}

func (jq *JobQueue) safeProcess(job *Job) error {
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("job %s panicked: %s, %s", job.Name, r, string(debug.Stack()))
			}
		}()

		err = job.Handler()
	}()
	return err
}
