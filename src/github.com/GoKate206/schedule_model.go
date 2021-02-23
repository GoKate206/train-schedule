package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"regexp/syntax"
	"strconv"
	"strings"
	"time"
)

var (
	layout          = "Jan 02 2006 15:04"
	dateLayout      = "Jan 02 2006"
	timeLayout      = "15:04"
	expectedHeaders = []string{"stopID", "route", "trainID", "time"}
)

type Schedule struct {
	StopID  int64  `json:"stopID"`
	Route   string `json:"route"`
	TrainID string `json:"trainID"`
	Time    string `json:"time"` // TODO: is this the right thing?
	ID      string `json:"ID,omitempty"`
}

//*====================*
//    Verifications
//*====================*
func verifyHeaders(headers []string) error {
	// if the headers are not the expected length, raise an error
	if len(headers) != len(expectedHeaders) {
		return fmt.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(headers))
	}

	// if headers are not expected name % order, raise an error
	for i, header := range headers {
		if header != expectedHeaders[i] {
			return fmt.Errorf("Incorrect header. Expected %s, Got: %s", expectedHeaders[i], header)
		}
	}
	return nil
}

func validateTrainID(trainID string) (string, error) {
	const trainIDLength = 4
	if len(trainID) != trainIDLength {
		return "", fmt.Errorf("Train Id is invalid, too few characters: %s", trainID)
	}

	for _, r := range trainID {
		if !syntax.IsWordChar(r) {
			return "", fmt.Errorf("Train Id must be alphanumeric: %s", trainID)
		}
	}

	return trainID, nil
}

func validateAndParseTime(scheduledTime string) (string, error) {
	// convert to time.Time for comparisons and validation
	givenTime, err := time.Parse(layout, scheduledTime)
	if err != nil {
		return "", err
	}

	// Do not schedule for trains in the past
	timeNow := time.Now()
	if givenTime.Before(timeNow) {
		return "", fmt.Errorf("Scheduled time must be in the future")
	}

	return scheduledTime, nil
}

//*========================*
//    CSV Read & DB Write
//*========================*

func csvHandler(schedule string) {
	// get slice Schedule struct from csv
	schedules, err := readCsv(schedule)
	if err != nil {
		log.Panicf(fmt.Sprintf("Error reading CSV: %v", err))
		return
	}

	// insert struct values into db ( Scribble here, ideally PostgreSQL)
	err = insertSchedules(schedules)
	if err != nil {
		log.Panicf(fmt.Sprintf("Insert Schedules error: %v", err))
		return
	}
}

func readCsv(givenCsv string) (schedules []Schedule, err error) {
	if givenCsv == "" {
		return schedules, err
	}

	r := csv.NewReader(strings.NewReader(givenCsv))
	records, err := r.ReadAll()
	// if there is an error, or no csv rows return
	if err != nil || len(records) == 0 {
		return schedules, err
	}

	// Verify header matches expected
	err = verifyHeaders(records[0])
	if err != nil {
		return schedules, err
	}

	// First row is always headers, validate and populate slice
	for _, record := range records[1:] {
		if len(record) != len(expectedHeaders) {
			return nil, fmt.Errorf("Incorrect number of columns. Expected: %d, Got: %d", len(record), len(expectedHeaders))
		}

		stopID, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return nil, err
		}

		trainID, err := validateTrainID(record[2])
		if err != nil {
			return nil, err
		}

		// use go pkg time to validate and parse to expected string
		trainTime, err := validateAndParseTime(record[3])
		if err != nil {
			return nil, err
		}

		schedule := Schedule{
			StopID:  stopID,
			Route:   record[1], // TODO: verify with route in db
			TrainID: trainID,
			Time:    trainTime,
		}

		schedules = append(schedules, schedule)
	}

	return schedules, err
}

func insertSchedules(schedules []Schedule) (err error) {
	for i, schedule := range schedules {
		id := fmt.Sprintf("%d_%s", i, schedule.TrainID)
		schedule.ID = id
		if err = db.Write(scheduleDbName, id, &schedule); err != nil {
			return err
		}
	}

	return nil
}

//*========================*
//    Get Schedule
//*========================*
func getScheduleByStop(stopID int64, selectedTime string) (nextTrains []Schedule, err error) {
	// Parse given date to time.Time
	selectedDate, err := time.Parse(layout, selectedTime)
	if err != nil {
		return nextTrains, err
	}

	// Get all trains scheduled for all stops today
	// If this were sql the query would be easier / faster
	todaySchedule, err := getScheduleByDate(selectedDate)
	if err != nil {
		return nextTrains, err
	}

	// iterate through schedule for 2+ Trains arriving within range
	for _, scheduleData := range todaySchedule {
		if scheduleData.StopID != stopID {
			continue
		}

		// is the train arriving within five minutes?
		trainWithinSchedule, err := isTrainWithinSchedule(selectedTime, scheduleData.Time)
		if err != nil {
			return nextTrains, err
		}

		if trainWithinSchedule {
			nextTrains = append(nextTrains, scheduleData)
		}

	}

	var (
		lastOfToday        = todaySchedule[len(todaySchedule)-1]
		showTomorrowTrains = false
		nextDaySchedule    []Schedule
	)

	// If lastOfToday has data check if requested time
	// is after the last train of the day
	if lastOfToday.Route != "" {
		lastTrain, _ := time.Parse(layout, lastOfToday.Time)
		showTomorrowTrains = selectedDate.After(lastTrain)

		nextDaySchedule, err = getScheduleByDate(selectedDate.Add(time.Hour * 24))
		if err != nil {
			return nextTrains, err
		}
	}

	if showTomorrowTrains && len(nextDaySchedule) > 1 {
		return nextDaySchedule, nil
	}

	if len(nextTrains) < 2 {
		return []Schedule{}, nil
	}

	return nextTrains, nil
}

func isTrainWithinSchedule(requestedTime, scheduleTime string) (bool, error) {
	// TODO: allow different dateTime formats
	requested, err := time.Parse(layout, requestedTime)
	if err != nil {
		return false, err
	}

	scheduled, err := time.Parse(layout, scheduleTime)
	if err != nil {
		return false, err
	}

	return requested.Equal(scheduled), nil
}
