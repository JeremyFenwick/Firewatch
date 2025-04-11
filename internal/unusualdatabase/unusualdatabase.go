package unusualdatabase

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type udpMessage struct {
	message string
	sender  net.Addr
}

const protectedKey = "version"

func Listen(port int) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Create the database
	db := &weirdDatase{
		data: make(map[string]string, 0),
	}
	db.insert("version", "madvillains vault of villainy")
	// Determine if we are local or remote (fly io)
	bindingAddress := getBindingAddress(port)
	// Setup the udp listener
	udp, err := net.ListenPacket("udp", bindingAddress)
	if err != nil {
		log.Fatalf("can't listen on %d/udp: %s", port, err)
	}
	defer udp.Close()
	log.Printf("Unusual database listening on port %d", port)
	buffer := make([]byte, 999)
	// Start recieving messages
	for {
		n, senderAddress, err := udp.ReadFrom(buffer)
		if err != nil {
			log.Println("Could not recieve packet, continuing...")
			continue
		}
		message := string(buffer[:n])
		log.Printf("Recieved message: %s", message)
		request := &udpMessage{
			message: message,
			sender:  senderAddress,
		}
		go handleRequest(request, db, udp)
	}
}

func handleRequest(request *udpMessage, db *weirdDatase, conn net.PacketConn) {
	key, value, isInsert := strings.Cut(request.message, "=")
	if isInsert {
		handleInsert(key, value, db)
	} else {
		handleDataRequest(key, request, db, conn)
	}
}

func handleInsert(key, value string, db *weirdDatase) {
	// The version key is protected
	if key == protectedKey {
		return
	}
	db.insert(key, value)
}

func handleDataRequest(key string, request *udpMessage, db *weirdDatase, conn net.PacketConn) {
	exists, value := db.retrieve(key)
	if !exists {
		_, err := conn.WriteTo([]byte(key+"="), request.sender)
		if err != nil {
			log.Println("Could not send key not found message back to sender")
		}
		return
	}
	_, err := conn.WriteTo([]byte(request.message+"="+value), request.sender)
	if err != nil {
		log.Println("Could not send value back to client")
	}
}

func getBindingAddress(port int) string {
	_, exists := os.LookupEnv("FLY_APP_NAME")
	if exists {
		return fmt.Sprintf("fly-global-services:%d", port)
	} else {
		return fmt.Sprintf("0.0.0.0:%d", port)
	}
}
