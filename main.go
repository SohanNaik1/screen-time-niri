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

type WindowInfo struct {
	AppId string
	Title string
}

type WindowRecord struct {
	AppId    string        `json:"app_id"`
	Title    string        `json:"title"`
	Duration time.Duration `json:"duration"`
}

func saveStats(stats map[int]time.Duration, names map[int]WindowInfo) {
	saveData := make(map[int]WindowRecord)
	for id, dur := range stats {
		info := names[id]
		saveData[id] = WindowRecord{
			AppId:    info.AppId,
			Title:    info.Title,
			Duration: dur,
		}
	}
	jsonData, _ := json.MarshalIndent(saveData, "", "  ")
	err := os.WriteFile("stats.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing to the file : ", err)
		return
	}
}

func loadStats() map[int]time.Duration {
	stats := make(map[int]time.Duration)

	data, err := os.ReadFile("stats.json")
	if err != nil {
		return stats
	}
	// creating temporary map of int and string cause json can't read time.duration
	var tempMap map[int]string
	json.Unmarshal(data, &tempMap)

	for id, str := range tempMap {
		duration, _ := time.ParseDuration(str)
		stats[id] = duration
	}
	return stats
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

	windowStats := loadStats()
	windowNames := make(map[int]WindowInfo)

	scanner := bufio.NewScanner(conn)

	fmt.Fprint(conn, "\"Windows\"\n")
	if scanner.Scan() {
		var reply struct {
			Ok struct {
				Windows []struct {
					ID        int    `json:"id"`
					AppID     string `json:"app_id"`
					Title     string `json:title`
					IsFocused bool   `json:"is_focused"`
				} `json:"Windows"`
			} `json:"Ok"`
		}

		json.Unmarshal(scanner.Bytes(), &reply)

		// Now we access it through reply.Ok.Windows
		for _, win := range reply.Ok.Windows {
			windowNames[win.ID] = WindowInfo{
				AppId: win.AppID,
				Title: win.Title,
			}
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
			saveStats(windowStats, windowNames)
			fmt.Println("--Auto saved stats--")
		}
	}()

	go func() {
		socketFile := "/tmp/screen-time-niri-api.sock"
		os.Remove(socketFile)

		l, err := net.Listen("unix", socketFile)
		if err != nil {
			fmt.Println("Error: couldn't create socketFile. ", err)
			return
		}
		fmt.Println("Sever is now listening on", socketFile)

		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("SEVER ERROR: connection failed. ", err)
			}

			jsonData, _ := json.Marshal(windowStats)
			fmt.Fprintln(conn, string(jsonData))

			conn.Close()
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
				windowNames[win.ID] = WindowInfo{
					AppId: win.AppId,
					Title: win.Title,
				}
			}
		}

		if event.WindowFocusChanged != nil {
			duration := time.Since(startTime)
			windowStats[activeWindowId] += duration
			saveStats(windowStats, windowNames)
			info, exists := windowNames[activeWindowId]
			name := "unknown"
			title := "unknown"
			if exists {
				name = info.AppId
				title = info.Title
			}
			fmt.Printf("Name : %s | Title : %s | Total time : %v \n", name, title, windowStats[activeWindowId])
			activeWindowId = event.WindowFocusChanged.ID
			startTime = time.Now()
		}
	}
}
