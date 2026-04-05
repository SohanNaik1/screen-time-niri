package main

import (
	"fmt"
	"io"
	"net"
)

func main() {
	conn, err := net.Dial("unix", "/tmp/screen-time-niri-api.sock")
	if err != nil {
		fmt.Println("Is The tracker running?", err)
		return
	}
	defer conn.Close()

	result, _ := io.ReadAll(conn)
	fmt.Println(string(result))
}
