package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/sdomino/scribble"
)

var (
	db             *scribble.Driver
	scheduleDbName = "./schedule"
)

func tearDownDb() {
	if err := db.Delete(scheduleDbName, ""); err != nil {
		log.Fatal(fmt.Sprintf("Error deleting from db: %v", err))
	}
}

func initDb() {
	var err error
	db, err = scribble.New(scheduleDbName, nil)
	if err != nil {
		log.Fatal("Cannot create Schedule DB")
		return
	}
}

func bytesToSchedule(bytes []byte) (Schedule, error) {
	schedule := Schedule{}
	err := json.Unmarshal(bytes, &schedule)

	return schedule, err
}

func getScheduleByDate(givenDate time.Time) ([]Schedule, error) {
	schedules := []Schedule{}
	day := givenDate.Format(dateLayout)

	bytes, err := db.ReadAll(scheduleDbName)
	if err != nil {
		return schedules, err
	}

	for _, b := range bytes {
		schedule, err := bytesToSchedule(b)
		if err != nil {
			return schedules, err
		}

		trainDate, err := time.Parse(layout, schedule.Time)
		if err != nil {
			return schedules, err
		}

		if trainDate.Format(dateLayout) == day {
			schedules = append(schedules, schedule)
		}
	}

	sort.Slice(schedules, func(i int, j int) bool {
		return schedules[i].Time < schedules[j].Time
	})

	return schedules, nil
}

func getAllStops() ([]int64, error) {
	trainStops := []int64{}
	bytes, err := db.ReadAll(scheduleDbName)
	if err != nil {
		return trainStops, err
	}

	uniqueStops := map[int64]bool{}
	for _, b := range bytes {
		schedule, err := bytesToSchedule(b)
		if err != nil {
			return trainStops, err
		}

		if !uniqueStops[schedule.StopID] {
			trainStops = append(trainStops, schedule.StopID)
		}
	}

	return trainStops, nil
}
