# jobqueue

- An orchestrator to schedule and synchronize jobs
- Handles recurring jobs
- Handles jobs with other jobs as dependancies
- Retries Failed jobs upto 3 times using a backoff


# How to run (will need go environment set up)

- go get github.com/aashishthite/jobqueue
- cd $GOPATH/src/github.com/aashishthite/jobqueue
- go run examples/main.go

# TODO

- Make hardcoded values configurable
- Use a database to store job configs to scale the jobqueue to work across multiple servers.
- Allow custom function formats for jobs
