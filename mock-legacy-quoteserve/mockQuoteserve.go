package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

const (
	connHost = "172.20.0.1" // Run on the local machine
	connPort = "4444"       // Same port as on the regular system
	connType = "tcp"        // NOTE: not HTPP
)

func main() {
	// Listen for incoming connections.
	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + connHost + ":" + connPort)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	recv, err := conn.Read(buf)
	fmt.Println(string(buf[:recv-1]))
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	split := strings.Split(string(buf[:recv]), ",")
	time.Sleep(time.Duration(rand.Intn(30)) * time.Millisecond)
	// Send a response back to person contacting us.
	conn.Write([]byte(makeResponse(split[0], strings.TrimSpace(split[1]))))
	// Close the connection when you're done with it.
	conn.Close()
}

func makeResponse(stock string, username string) string {
	time := time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
	amount := fmt.Sprintf("%d.%d", rand.Intn(500)+1, rand.Intn(100))
	crypto := randSeq(25)
	// (?P<quote>.+),(?P<stock>.+),(?P<user>.+),(?P<time>.+),(?P<key>.+)
	output := fmt.Sprintf("%s,%s,%s,%d,%s\n",
		amount,
		stock,
		username,
		time,
		crypto)
	fmt.Println(output)
	return output
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
