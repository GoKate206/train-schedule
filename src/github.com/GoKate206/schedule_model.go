package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/sdomino/scribble"
)

var (
	expectedHeaders = []string{"stopID", "route", "trainID", "time"}
	db              *scribble.Driver
	scheduleDbName  = "schedule"
)

type Schedule struct {
	StopID  int64  `json:"stopID"`
	Route   string `json:"route"`
	TrainID string `json:"trainID"`
	Time    string `json:"time"` // TODO: is this the right thing?
}

func csvHandler(schedule string) {
	schedules, err := readCsv(schedule)
	if err != nil {
		log.Panicf(fmt.Sprintf("Error reading CSV: %v", err))
		return
	}

	err = insertSchedules(schedules)
	if err != nil {
		log.Panicf(fmt.Sprintf("Insert Schedules error: %v", err))
		return
	}
}

func readCsv(schedule string) (schedules []Schedule, err error) {
	if schedule == "" {
		return schedules, err
	}

	r := csv.NewReader(strings.NewReader(schedule))
	records, err := r.ReadAll()
	if err != nil || len(records) == 0 {
		return schedules, err
	}

	err = verifyHeaders(records[0])
	if err != nil {
		return schedules, err
	}

	for _, record := range records[1:] {
		if len(record) != len(expectedHeaders) {
			return nil, fmt.Errorf("Incorrect number of columns. Expected: %d, Got: %d", len(record), len(expectedHeaders))
		}

		stopID, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return nil, err
		}

		trainID := record[2] // tood: Test for regex alphanumeric
		if len(trainID) < 4 {
			return nil, errors.New(fmt.Sprintf("Train Id is invalid, too few characters: %s", trainID))
		}

		schedule := Schedule{
			StopID:  stopID,
			Route:   record[1], // TODO: verify with route in db
			TrainID: trainID[:4],
			Time:    record[3],
		}
		schedules = append(schedules, schedule)
	}

	return schedules, err
}

func verifyHeaders(headers []string) error {
	if len(headers) != len(expectedHeaders) {
		return fmt.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(headers))
	}

	for i, header := range headers {
		if header != expectedHeaders[i] {
			return fmt.Errorf("Incorrect header. Expected %s, Got: %s", expectedHeaders[i], header)
		}
	}
	return nil
}

func insertSchedules(schedules []Schedule) error {
	all, _ := db.ReadAll(fmt.Sprintf("./%s", scheduleDbName))
	fmt.Println(len(all))

	id := len(all)
	for _, schedule := range schedules {
		id++
		if err := db.Write(scheduleDbName, fmt.Sprintf("%d", id), &schedule); err != nil {
			return err
		}
	}

	return nil
}

func initDb() {
	var err error
	db, err = scribble.New(fmt.Sprintf("./%s", scheduleDbName), nil)
	if err != nil {
		log.Fatal("Cannot create Schedule DB")
		return
	}

	fmt.Println(db)
}
