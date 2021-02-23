package main

import (
	"fmt"
	"log"
)

func main() {
	initDb()
	stops, err := getAllStops()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error getting all stops: %v", err))
	}
	fmt.Println(stops)
}
