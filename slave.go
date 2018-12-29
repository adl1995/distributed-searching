package main

import (
	"os"
	"flag"
	"fmt"
	"log"
	"net"
	"io/ioutil"
	"strings"
	"bufio"
)

func performSearching(fileNamesChan <-chan []string, tokenChan <-chan string,
					  terminateSearch chan<- bool, directory string, terminateProcess <-chan bool) {
	fileNames := <- fileNamesChan
	token := <- tokenChan

	for _, fileName := range fileNames {
		fmt.Println("Searching file " + fileName + "...")
		file, err := os.Open("../" + directory + "/" + fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		// Read lines separated by newline.
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines) 

		for scanner.Scan() {
			word := scanner.Text()
			if word == token {
				terminateSearch <- true
				return
			}

			if token == "ABORT" {
				terminateSearch <- true
				return
			}
			// Check if an ABORT message was received.
			// select {
			// case abort := <- terminateProcess:
			// 	if (abort) {
			// 		return
			// 	}
			// default:
			// 	// Do nothing...
			// 	continue
			// }
		}
	}
	terminateSearch <- false
}

func assembleSearching(connection net.Conn, fileNamesTokenChan <-chan string,
					   terminateSearch chan bool, fileNamesChan chan []string,
					   tokenChan chan string, directory string, terminateProcess <-chan bool) {
	fileNamesTokenAll := <- fileNamesTokenChan
	
	fileNamesToken := strings.Split(fileNamesTokenAll, "|")
	token := fileNamesToken[len(fileNamesToken) - 1]
	fileNames := fileNamesToken[:len(fileNamesToken) - 1]

	fmt.Println("Token:", token)

	if (token != "ABORT") {
		go performSearching(fileNamesChan, tokenChan, terminateSearch, directory, terminateProcess)

		fileNamesChan <- fileNames
		tokenChan <- token

		matchFound := <- terminateSearch

		if matchFound == true {
			fmt.Println("Match found! Terminating search...")
			connection.Write([]byte("found\n"))
		} else {
			fmt.Println("No match found.")
			connection.Write([]byte("not found\n"))
		}
	}
}

func receiveTasks(connection net.Conn, fileNamesToken chan string, terminateProcess chan bool,
				  terminateSearch chan bool, fileNamesChan chan []string, tokenChan chan string, directory string) {
	for {
		message, _ := bufio.NewReader(connection).ReadString('\n')
		message = strings.TrimSuffix(message, "\n")

		fmt.Println("Task received: " + message)
		if (message == "ABORT") {
			// Terminate the process
			// terminateProcess <- true
			fmt.Println("ABORT received! terminating search...")
			// break
		}
		go assembleSearching(connection, fileNamesToken, terminateSearch, fileNamesChan, tokenChan, directory, terminateProcess)

		fileNamesToken <- message
	}
}

func main() {
	// Handle CLI parameters.
	serverIPAndPort := flag.String("serverIPAndPort", "", "a string")
	directory := flag.String("directory", "", "a string")

	flag.Parse()

	if (*serverIPAndPort == "" || *directory == "") {
		log.Fatal("Please provide valid values for CLI flag arguments. Program exiting...")
	}

	fmt.Println("serverIPAndPort:", *serverIPAndPort)
	fmt.Println("directory:", *directory)

	// Connect to the main server to get clients list.
	connection, _ := net.Dial("tcp", *serverIPAndPort)

	// Read the list of available file(s).
	files, err := ioutil.ReadDir(".")

	if err != nil {
		log.Fatal(err)
	}

	var fileNames string
	for _, file := range files {
		fileNames += (file.Name() + "|")
	}
	fileNames += "\n"

	// Send file name(s) to server.
	connection.Write([]byte(fileNames))

	fileNamesToken := make(chan string)
	fileNamesChan := make(chan []string)
	tokenChan := make(chan string)
	terminateSearch := make(chan bool)
	terminateProcess := make(chan bool)

	go receiveTasks(connection, fileNamesToken, terminateProcess, terminateSearch, fileNamesChan, tokenChan, *directory)
	// go assembleSearching(connection, fileNamesToken, terminateSearch, fileNamesChan, tokenChan)
	// go performSearching(fileNamesChan, tokenChan, terminateSearch, *directory)

	<- terminateProcess
	fmt.Println("Process exiting...")
}