package main

import (
	"log"

	internalapp "github.com/jsalvag/pomodify/internal/app"
)

func main() {
	if err := internalapp.Run(); err != nil {
		log.Fatal(err)
	}
}
