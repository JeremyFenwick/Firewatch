package mobinthemiddle

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type contextPackage struct {
	ctx    context.Context
	cancel context.CancelFunc
	logger *log.Logger
	wg     *sync.WaitGroup
}

func Listen(port int, upstream string, upstreamPort int) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Could not start listener. REASON: %v", err)
	}
	log.Printf("Mob in the middle now listening on port %d\n", port)
	defer listener.Close()

	upstreamAddress := net.JoinHostPort(upstream, fmt.Sprintf("%d", upstreamPort))

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Encountered error accepting connection. REASON: %v", err)
			continue
		}
		go handleConnection(clientConn, upstreamAddress)
	}
}

func handleConnection(victimConn net.Conn, upstreamAddress string) {
	// Connect to upstream server
	upstreamConn, err := net.Dial("tcp", upstreamAddress)
	if err != nil {
		log.Println("Could not connect to upstream server")
		victimConn.Close()
		return
	}

	// Setup readers
	clientReader := bufio.NewReader(victimConn)
	upstreamReader := bufio.NewReader(upstreamConn)

	// Setup context package
	ctx, cancel := context.WithCancel(context.Background())
	logger := log.New(log.Writer(),
		fmt.Sprintf("[%s] ", victimConn.RemoteAddr().String()),
		log.Flags()|log.Lmsgprefix|log.Lshortfile)
	var wg sync.WaitGroup
	wg.Add(2)
	contextPackage := &contextPackage{
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
		wg:     &wg,
	}

	// Start bidirectional relay
	go relayMessages(clientReader, "victim", upstreamConn, "upstream", victimConn, contextPackage)
	go relayMessages(upstreamReader, "upstream", victimConn, "victim", upstreamConn, contextPackage)
}

func relayMessages(
	sourceReader *bufio.Reader,
	sourceName string,
	destConn net.Conn,
	destName string,
	sourceConn net.Conn,
	contextP *contextPackage) {

	// Defer closing connections
	defer sourceConn.Close()
	defer destConn.Close()
	defer contextP.wg.Done()

	for {
		select {
		case <-contextP.ctx.Done():
			return
		default:
			// Reset read deadline to prevent timeout
			sourceConn.SetReadDeadline(time.Now().Add(5 * time.Minute))

			// Read until newline
			message, err := sourceReader.ReadString('\n')
			if err != nil {
				contextP.logger.Printf("%s reader no longer active. Exiting", sourceName)
				contextP.cancel()
				return
			}

			// Trim the trailing newline
			messageString := strings.TrimSuffix(message, "\n")

			// Process message
			injectedMessage := MotmAttack(messageString)

			// Add the newline back
			_, err = fmt.Fprintln(destConn, injectedMessage)
			if err != nil {
				contextP.logger.Printf("Could not write to %s. Exiting...", destName)
				contextP.cancel()
				return
			}

			contextP.logger.Printf("Sent message from %s to %s: %q", sourceName, destName, injectedMessage)
		}
	}
}
