package primetime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
)

type Request struct {
	Method string `json:"method"`
	Number any    `json:"number"`
}

type Response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func Listen(port int) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("Could not started listener. REASON: " + err.Error())
	}
	log.Printf("Prime time now listening on port %d\n", port)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Encountered error accepting connection. REASON: " + err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		requestBytes, err := reader.ReadBytes('\n')
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			return
		}
		if err != nil {
			log.Println("Failed to receive message from connection:", err)
			return
		}

		err = isPrimeResponse(requestBytes, conn)
		if err != nil {
			return
		}
	}
}

func isPrimeResponse(request []byte, conn net.Conn) error {
	requestStruct, err := decodeJson(request)
	if err != nil {
		responseError := fmt.Sprintf("Failed to parse recieved json. Closing connection. REASON: %s", err.Error())
		log.Println(responseError)
		conn.Write([]byte(responseError))
		return err
	}

	response, err := createResponse(requestStruct)
	if err != nil {
		log.Println("Failed to generate response. REASON: " + err.Error())
		return err
	}

	bytes, err := encodeJson(response)
	if err != nil {
		log.Println("Failed to encode json. Closing connection. REASON: " + err.Error())
		return err
	}

	_, err = conn.Write(append(bytes, []byte("\n")...))
	if err != nil {
		log.Println("Failed to send bytes. Closing connection. REASON: " + err.Error())
		return err
	}

	return nil
}

func decodeJson(data []byte) (*Request, error) {
	var request Request
	err := json.Unmarshal(data, &request)
	if err != nil {
		return nil, err
	}

	if request.Number == nil {
		return nil, fmt.Errorf("no number provided")
	}

	_, ok := request.Number.(float64)
	if !ok {
		return nil, fmt.Errorf("number field did not contain a number")
	}

	return &request, nil
}

func encodeJson(response *Response) ([]byte, error) {
	bytes, err := json.Marshal(&response)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func createResponse(request *Request) (*Response, error) {
	if request.Method != "isPrime" {
		return nil, fmt.Errorf("unknown request %s", request.Method)
	}

	primeResult := false
	number, _ := request.Number.(float64)

	if number > 0 && math.Trunc(number) == number {
		primeResult = IsPrime(int(number))
	}

	return &Response{
		Method: "isPrime",
		Prime:  primeResult,
	}, nil
}

// isPrime checks if a given integer n is a prime number.
func IsPrime(n int) bool {
	// Handle edge cases for numbers less than 2.
	if n <= 1 {
		return false
	}
	// 2 and 3 are prime numbers.
	if n <= 3 {
		return true
	}
	// If n is divisible by 2 or 3, it's not prime.
	if n%2 == 0 || n%3 == 0 {
		return false
	}

	// Optimization: We only need to check divisors up to the square root of n.
	// All other divisors will have a corresponding smaller divisor.
	sqrtN := int(math.Sqrt(float64(n)))

	// We only need to check divisors of the form 6k ± 1.
	for i := 5; i <= sqrtN; i = i + 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}

	return true
}
