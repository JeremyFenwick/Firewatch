package jobcenter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

type Client struct {
	Conn   net.Conn
	Reader *bufio.Reader
	Jobs   []*Job
}

func Listen(port int) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		// Use Fatalf for formatted error messages
		log.Fatalf("Could not start listener. REASON: %v", err)
	}
	log.Printf("Job center now listening on port %d\n", port)
	defer listener.Close()

	queueManager := NewQueueManager()

	for {
		conn, err := listener.Accept()
		if err != nil {
			// Use Printf for formatted error messages
			log.Printf("Encountered error accepting connection. REASON: %v", err)
			continue
		}

		go handleConnection(conn, queueManager)
	}
}

func handleConnection(conn net.Conn, queueManager *QueueManager) {
	defer conn.Close()

	client := &Client{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
		Jobs:   make([]*Job, 0),
	}
	defer returnJobs(client, queueManager)

	log.Printf("Handling connection from %s", conn.RemoteAddr())
	clientData, err := client.Reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading from client: %v", err)
		return
	}
	request := parseJsonLine([]byte(clientData))
	if request.Error != nil {
		log.Printf("Error parsing request: %v", request.Error)
		return
	}
	// Process the request based on its type
	switch request.Type {
	case Put:
		handlePut(client, request.Request, queueManager)
	case Get:
		handleGet(client, request.Get, queueManager)
	case Delete:
		handleDelete(client, request.Delete, queueManager)
	case Abort:
		handleAbort(client, request.Abort, queueManager)
	case Error:
		log.Printf("Error in request: %v", request.Error)
	default:
		log.Printf("Unknown request type: %d", request.Type)
	}
}

func returnJobs(client *Client, queueManager *QueueManager) {
	for _, job := range client.Jobs {
		if job != nil {
			queueManager.PutJob(job.Queue, job)
		}
	}
	client.Jobs = nil
}

func jobsContains(jobs []*Job, id int) (*Job, bool) {
	for _, job := range jobs {
		if job.Id == id {
			return job, true
		}
	}
	return nil, false
}

func deleteJob(jobs []*Job, job *Job) []*Job {
	for i, j := range jobs {
		if j.Id == job.Id {
			return append(jobs[:i], jobs[i+1:]...)
		}
	}
	return jobs
}

func handleAbort(client *Client, request *AbortRequest, queueManager *QueueManager) {
	log.Printf("Handling ABORT request for client: %s", client.Conn.RemoteAddr())
	abortResponse := AbortResponse{}
	job, exists := jobsContains(client.Jobs, request.Id)
	if exists {
		abortResponse.Status = "ok"
		queueManager.PutJob(job.Queue, job)
		client.Jobs = deleteJob(client.Jobs, job)
	} else {
		abortResponse.Status = "no-job"
	}
	responseData, err := json.Marshal(abortResponse)
	if err != nil {
		log.Printf("Error marshaling ABORT response: %v", err)
		return
	}
	client.Conn.Write(responseData)
}

func handleDelete(client *Client, request *DeleteRequest, queueManager *QueueManager) {
	log.Printf("Handling DELETE request for client: %s", client.Conn.RemoteAddr())
	deleted := queueManager.DeleteJob(request.Id)
	deleteResponse := DeleteResponse{}
	if deleted {
		deleteResponse.Status = "ok"
	} else {
		deleteResponse.Status = "no-job"
	}
	responseData, err := json.Marshal(deleteResponse)
	if err != nil {
		log.Printf("Error marshaling DELETE response: %v", err)
		return
	}
	client.Conn.Write(responseData)
}

func handlePut(client *Client, request *PutRequest, queueManager *QueueManager) {
	log.Printf("Handling PUT request for client: %s", client.Conn.RemoteAddr())
	newJob := &Job{
		Priority: request.Priority,
		Id:       queueManager.GetNextId(),
		Content:  request.Job,
		Queue:    request.Queue,
	}
	queueManager.PutJob(request.Queue, newJob)
	putResponse := PutResponse{
		Request: "get",
		Id:      newJob.Id,
	}
	responseData, err := json.Marshal(putResponse)
	if err != nil {
		log.Printf("Error marshaling PUT response: %v", err)
		return
	}
	client.Conn.Write(responseData)
}

func handleGet(client *Client, request *GetRequest, queueManager *QueueManager) {
	log.Printf("Handling GET request for client: %s", client.Conn.RemoteAddr())
	getResponse := GetResponse{}
	wait := false
	if request.Wait != nil {
		wait = *request.Wait
	}
	job, exists := queueManager.GetPriorityJob(request.Queues...)
	if !exists && !wait {
		log.Printf("No jobs found in queues: %v", request.Queues)
		getResponse.Status = "no-job"
		return
	} else if !exists && wait {
		for !exists {
			time.Sleep(5 * time.Second)
			client.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			// Check to see if the client is still connected
			_, err := client.Conn.Write([]byte{})
			if err != nil {
				log.Printf("Client disconnected: %s", client.Conn.RemoteAddr())
				return
			}
			job, exists = queueManager.GetPriorityJob(request.Queues...)
		}
	}

	client.Jobs = append(client.Jobs, job)
	getResponse.Status = "ok"
	getResponse.ID = &job.Id
	getResponse.Job = job.Content
	getResponse.Queue = &job.Queue
	getResponse.Priority = &job.Priority

	responseData, err := json.Marshal(getResponse)
	if err != nil {
		log.Printf("Error marshaling GET response: %v", err)
		return
	}
	client.Conn.Write(responseData)
}
