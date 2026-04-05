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

	WindowsChanged *[]struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		AppId string `json:"app_id"`
	} `json:"WindowsChanged"`
}

func saveStats(stats map[int]time.Duration) {
	printableStats := make(map[int]string)
	for id, duration := range stats {
		printableStats[id] = duration.String()
	}
	jsonData, err := json.MarshalIndent(printableStats, "", "  ")
	if err != nil {
		fmt.Println("Error encoding json :", err)
		return
	}
	err = os.WriteFile("stats.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing to the file : ", err)
		return
	}
}

func main() {
	socketPath := os.Getenv("NIRI_SOCKET")
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Println("Error Connecting :", err)
		return
	}

	var activeWindowId int
	var startTime time.Time = time.Now()

	windowStats := make(map[int]time.Duration)
	windowNames := make(map[int]string)

	scanner := bufio.NewScanner(conn)

	fmt.Fprint(conn, "\"Windows\"\n")
	if scanner.Scan() {
		var reply struct {
			Ok struct {
				Windows []struct {
					ID        int    `json:"id"`
					AppID     string `json:"app_id"`
					IsFocused bool   `json:"is_focused"`
				} `json:"Windows"`
			} `json:"Ok"`
		}

		json.Unmarshal(scanner.Bytes(), &reply)

		// Now we access it through reply.Ok.Windows
		for _, win := range reply.Ok.Windows {
			windowNames[win.ID] = win.AppID
			if win.IsFocused {
				activeWindowId = win.ID
			}
		}
	}

	fmt.Fprint(conn, "\"EventStream\"\n")
	defer conn.Close()

	ticker := time.NewTicker(30 * time.Second)

	go func() {
		for range ticker.C {
			windowStats[activeWindowId] += time.Since(startTime)
			startTime = time.Now()
			saveStats(windowStats)
			fmt.Println("--Auto saved stats--")
		}
	}()

	for scanner.Scan() {
		line := scanner.Bytes()

		var event NiriEvent

		err := json.Unmarshal(line, &event)
		if err != nil {
			continue
		}

		if event.WindowsChanged != nil {
			for _, win := range *event.WindowsChanged {
				windowNames[win.ID] = win.AppId
			}
		}

		if event.WindowFocusChanged != nil {
			duration := time.Since(startTime)
			windowStats[activeWindowId] += duration
			saveStats(windowStats)
			name, exists := windowNames[activeWindowId]
			if !exists {
				name = "unknown"
			}
			fmt.Printf("The total time for window %s : %v \n", name, windowStats[activeWindowId])
			activeWindowId = event.WindowFocusChanged.ID
			startTime = time.Now()
		}
	}
}
