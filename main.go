package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
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

	var activeWindowId int
	var startTime time.Time = time.Now()
	windowStats := make(map[int]time.Duration)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()

		var event NiriEvent

		err := json.Unmarshal(line, &event)
		if err != nil {
			continue
		}

		if event.WindowFocusChanged != nil {
			duration := time.Since(startTime)
			windowStats[activeWindowId] += duration
			fmt.Printf("The total time for window %d : %v \n", activeWindowId, windowStats[activeWindowId])
			activeWindowId = event.WindowFocusChanged.ID
			startTime = time.Now()
		}
	}
}
