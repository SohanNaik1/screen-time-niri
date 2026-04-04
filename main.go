package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type NiriEvent struct {
	WindowFocusChanged *struct {
		ID int `json:"id"`
	} `json:"WindowFocusChanged"`
}

func main() {
	socketPath := os.Getenv("NIRI_SOCKET")
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Println("Error Connecting :", err)
		return
	}
	defer conn.Close()

	fmt.Fprint(conn, "\"EventStream\"\n")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()

		var event NiriEvent

		err := json.Unmarshal(line, &event)
		if err != nil {
			continue
		}

		if event.WindowFocusChanged != nil {
			fmt.Printf("The focus was switched to ID : %d \n", event.WindowFocusChanged.ID)
		}
	}
}
