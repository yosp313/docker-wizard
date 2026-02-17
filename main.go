package main

import (
	"fmt"

	"docker-wizard/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
