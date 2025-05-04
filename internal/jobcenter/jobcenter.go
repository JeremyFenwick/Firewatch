package jobcenter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

// Client represents a client connection to the job center.

type Client struct {
	Conn   net.Conn
	Reader *bufio.Reader
	Jobs   []*Job
}

func (c *Client) returnAllJobs(queueManager *QueueManager) {
	for _, job := range c.Jobs {
		if job != nil {
			queueManager.PutJob(job.Queue, job)
		}
	}
	c.Jobs = nil
}

func (c *Client) hasJob(id int) (*Job, bool) {
	for _, job := range c.Jobs {
		if job.Id == id {
			return job, true
		}
	}
	return nil, false
}

func (c *Client) deleteJob(id int) bool {
	for i, j := range c.Jobs {
		if j.Id == id {
			c.Jobs = append(c.Jobs[:i], c.Jobs[i+1:]...)
			return true
		}
	}
	return false
}

func (c *Client) addJob(job *Job) {
	c.Jobs = append(c.Jobs, job)
}

// Listen starts the job center server on the specified port.

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
	defer client.returnAllJobs(queueManager)

	log.Printf("Handling connection from %s", conn.RemoteAddr())
	for {
		clientData, err := client.Reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from client: %v", err)
			return
		}
		request := parseJsonLine([]byte(clientData))
		if request.Error != nil {
			err := fmt.Errorf("error parsing request: %v", request.Error)
			handleError(client, err)
			continue
		}
		err = messageDispatch(client, request, queueManager)
		if err != nil {
			log.Printf("Error dispatching message: %v", err)
			return
		}
	}
}

func messageDispatch(client *Client, request Request, queueManager *QueueManager) error {
	switch request.Type {
	case Put:
		handlePut(client, request.Request, queueManager)
	case Get:
		// If the request does not specify a wait, set it to false
		if request.Get.Wait == nil {
			request.Get.Wait = new(bool)
			*request.Get.Wait = false
		}
		handleGet(client, request.Get, queueManager)
	case Delete:
		handleDelete(client, request.Delete, queueManager)
	case Abort:
		handleAbort(client, request.Abort, queueManager)
	default:
		unknownRequest := fmt.Errorf("unknown request type: %d", request.Type)
		handleError(client, unknownRequest)
	}
	return nil
}

func handleError(client *Client, err error) {
	response := ErrorResponse{
		Status: "error",
		Error:  err.Error(),
	}
	responseData, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling error response: %v", err)
		return
	}
	client.Conn.Write(append(responseData, '\n'))
}

func handleAbort(client *Client, request *AbortRequest, queueManager *QueueManager) {
	log.Printf("Handling ABORT request for client: %s", client.Conn.RemoteAddr())
	abortResponse := AbortResponse{}
	job, jobFound := client.hasJob(request.Id)
	jobExists := queueManager.JobExists(request.Id)
	if jobFound {
		abortResponse.Status = "ok"
		queueManager.PutJob(job.Queue, job)
		client.deleteJob(job.Id)
	} else if jobExists {
		// If the job exists in the queue manager but not in the client's jobs
		// we need to send back an error response instead
		handleError(client, fmt.Errorf("job %d is not controlled by the client", request.Id))
		return
	} else {
		abortResponse.Status = "no-job"
	}

	responseData, err := json.Marshal(abortResponse)
	if err != nil {
		log.Printf("Error marshaling ABORT response: %v", err)
		return
	}
	client.Conn.Write(append(responseData, '\n'))
}

func handleDelete(client *Client, request *DeleteRequest, queueManager *QueueManager) {
	log.Printf("Handling DELETE request for client: %s", client.Conn.RemoteAddr())
	job, exists := client.hasJob(request.Id)
	deleted := false
	if exists {
		// Remove the job from the client's jobs
		deleted = client.deleteJob(job.Id)
	} else {
		// Else, try to delete the job from the queue manager
		deleted = queueManager.DeleteJob(request.Id)
	}
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
	client.Conn.Write(append(responseData, '\n'))
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
		Status: "ok",
		Id:     newJob.Id,
	}
	responseData, err := json.Marshal(putResponse)
	if err != nil {
		log.Printf("Error marshaling PUT response: %v", err)
		return
	}
	client.Conn.Write(append(responseData, '\n'))
}

func handleGet(client *Client, request *GetRequest, queueManager *QueueManager) {
	log.Printf("Handling GET request for client: %s", client.Conn.RemoteAddr())
	getResponse := GetResponse{}
	job, exists := queueManager.GetPriorityJob(request.Queues...)
	// If no job exists and wait is false, return "no-job" status
	if !exists && !*request.Wait {
		log.Printf("No jobs found in queues: %v", request.Queues)
		getResponse.Status = "no-job"
		responseData, err := json.Marshal(getResponse)
		if err != nil {
			log.Printf("Error marshaling GET response: %v", err)
			return
		}
		client.Conn.Write(append(responseData, '\n'))
		return
	} else if !exists && *request.Wait {
		var err error // Here we need to wait for a job
		job, err = waitForJob(client, queueManager, request.Queues)
		if err != nil {
			log.Printf("Error waiting for job: %v", err)
			return
		}
	}

	client.addJob(job)
	// Setup the response
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
	client.Conn.Write(append(responseData, '\n'))
}

func waitForJob(client *Client, queueManager *QueueManager, queues []string) (*Job, error) {
	job, exists := queueManager.GetPriorityJob(queues...)
	for !exists {
		time.Sleep(5 * time.Second)
		client.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		// Check to see if the client is still connected
		_, err := client.Conn.Write([]byte{})
		if err != nil {
			return nil, fmt.Errorf("client disconnected: %s", client.Conn.RemoteAddr())
		}
		job, exists = queueManager.GetPriorityJob(queues...)
	}
	return job, nil
}
