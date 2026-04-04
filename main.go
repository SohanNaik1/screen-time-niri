package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

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
		fmt.Println(scanner.Text())
	}
}
