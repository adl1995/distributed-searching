package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"bufio"
	"strings"
)

var slaveList []net.Conn
var clientList []net.Conn
var slaveFileMap = make(map[net.Conn] string)
var clientPasswordMap = make(map[string] net.Conn)

func handleClient(clients <-chan net.Conn, toSearch chan string) {
	for {
		// Call slave with the received search token.
		connection := <-clients
		fmt.Println("New client connected, connection: ", connection)
		// Receive token to search.
		searchToken, _ := bufio.NewReader(connection).ReadString('\n')
		
		go handleWorkLoad(toSearch)

		toSearch <- searchToken

		clientPasswordMap[searchToken] = connection
		fmt.Println("New client connected, connection: ", clientPasswordMap[searchToken])
	}
}

func handleWorkLoad(toSearch <-chan string) {
	wordToSearch := <-toSearch
	fmt.Println("Password to search is " + wordToSearch)

	// Send task to each worker process.
	for idx, connection := range(slaveList) {
		fileNames := slaveFileMap[connection]
		_, err := connection.Write([]byte(fileNames + wordToSearch + "\n"))

		if err == nil {
			fmt.Print("Worker", idx, "- files:", fileNames, "- target:", wordToSearch)
		}
	}

	var responses []string
	// Receive result from each worker process.
	for _, connection := range(slaveList) {
		message, _ := bufio.NewReader(connection).ReadString('\n')
		fmt.Print("Result: ", message)

		responses = append(responses, message)

		if (message == "found\n") {
			// Send ABORT to all slaves.
			for _, abort_conn := range(slaveList) {
				if (abort_conn != connection) {
					abort_conn.Write([]byte("ABORT\n"))
				}
			}
			break
		}
	}

	var isFound bool = false
	for _, response := range(responses) {
		if response == "found\n" {
			isFound = true
		}
	}

	// Write to client.
	if (isFound == true) {
		clientPasswordMap[wordToSearch].Write([]byte("found\n"))
	} else {
		clientPasswordMap[wordToSearch].Write([]byte("not found\n"))
	}

	// Go through the available list of files
	// and assign task to worker processes.
	
}

func handleSlave(slaves <-chan net.Conn, toSearch <-chan string) {
	for {
		// Wait for the list of files.
		connection := <-slaves
		message, _ := bufio.NewReader(connection).ReadString('\n')
		message = strings.TrimSuffix(message, "\n")

		slaveFileMap[connection] = message
		fileNames := strings.Split(message, "|")
		fmt.Println("File received:", fileNames)

		// go handleWorkLoad(toSearch)
	}
	// Wait for response (in a separate thread?)

	// Get the search token.

	// Assign task to worker process.

	// Alert client.
}

func listenClient(port string, clients chan net.Conn, toSearch chan string) {
	// Start a new TCP connection.
	ln, err := net.Listen("tcp", ":" + port)

	if err != nil {
		log.Fatal(err)
	}

	for {
		connection, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		// Welcome client and update status.
		greet := 
		"** Welcome to PasswordFinder! Your request has been passed to a worker process. Please wait... **\n"

		connection.Write([]byte(greet))

		// Create a separate thread for each connecting client.
		go handleClient(clients, toSearch)

		clients <- connection
		clientList = append(clientList, connection)
	}
}

func listenSlave(port string, slaves chan<- net.Conn) {
	// Start a new TCP connection.
	ln, err := net.Listen("tcp", ":" + port)

	if err != nil {
		log.Fatal(err)
	}

	for {
		connection, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		slaves <- connection
		slaveList = append(slaveList, connection)
		// Assign a task.
	}
}

func main() {
	// Handle CLI parameters.
	clientPort := flag.String("clientPort", "8090", "a string")
	slavePort := flag.String("slavePort", "9090", "a string")

	flag.Parse()

	fmt.Println("clientPort:", *clientPort)
	fmt.Println("slavePort:", *slavePort)

	clients := make(chan net.Conn)
	slaves := make(chan net.Conn)
	toSearch := make(chan string)

	go listenClient(*clientPort, clients, toSearch)
	go listenSlave(*slavePort, slaves)

	// Create a separate thread for each connecting slave.
	// go handleClient(clients, toSearch)
	go handleSlave(slaves, toSearch)

	for { }
}
