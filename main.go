package main

import (
	"fmt"
	"os"

	"docker-wizard/internal/app"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--version" || arg == "-v" {
			fmt.Printf("docker-wizard:%s\n", version)
			return
		}
	}
	if err := app.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
