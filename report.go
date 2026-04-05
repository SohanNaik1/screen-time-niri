package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("unix", "/tmp/screen-time-niri-api.sock")
	if err != nil {
		fmt.Println("Is The tracker running?", err)
		return
	}
	defer conn.Close()

	result, _ := io.ReadAll(conn)

	var stats map[string]time.Duration
	json.Unmarshal(result, &stats)
	fmt.Printf("%-10s | %-20s\n", "Window ID", "Time Spent")
	fmt.Println("---------------------------------")

	for id, duration := range stats {
		fmt.Printf("%-10s | %-20s\n", id, duration.Round(time.Second).String())
	}
}
