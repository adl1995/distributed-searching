package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"bufio"
)

func main() {
	// Handle CLI parameters.
	searchText := flag.String("searchText", "", "a string")
	serverIPAndPort := flag.String("serverIPAndPort", "", "a string")

	flag.Parse()

	if (*searchText == "" || *serverIPAndPort == "") {
		log.Fatal("Please provide valid values for the flag CLI arguments. Program exiting...")
	}

	fmt.Println("searchText:", *searchText)
	fmt.Println("serverIPAndPort:", *serverIPAndPort)

	// Connect to the main server to get clients list.
	connection, _ := net.Dial("tcp", *serverIPAndPort)

	// Send job request to server.
	_, err := connection.Write([]byte(*searchText + "\n"))

	if (err == nil) {
		fmt.Println("Request sent to server. Awaiting response...")
	}

	// Receive greeting from server.
	message, _ := bufio.NewReader(connection).ReadString('\n')
	fmt.Print(message)

	// Receive response from server.
	message, _ = bufio.NewReader(connection).ReadString('\n')
	fmt.Print("Result: ", message)
}
