package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadCsv(t *testing.T) {
	csvHeaders := "stopID,route,trainID,time"
	invokeRead := func(csv string) []Schedule {
		schedules, err := readCsv(csv)
		require.Nil(t, err)
		return schedules
	}

	t.Run("when readCsv is invoked with an invalid csv", func(t *testing.T) {
		t.Run("when csv is not present", func(t *testing.T) {
			schedules := invokeRead("")

			t.Run("it will return an empty slice", func(t *testing.T) {
				assert.Len(t, schedules, 0)
			})
		})

		t.Run("when the csv has a header, but no rows", func(t *testing.T) {
			schedules := invokeRead(csvHeaders)

			t.Run("it will return an empty slice", func(t *testing.T) {
				assert.Len(t, schedules, 0)
			})
		})

		t.Run("when the csv has wrong header names", func(t *testing.T) {
			_, err := readCsv("stopId,route,trainID,time")

			t.Run("it will return an error", func(t *testing.T) {
				assert.NotNil(t, err)
				assert.EqualValues(t, err.Error(), "Incorrect header. Expected stopID, Got: stopId")
			})
		})

		t.Run("when there are too few headers", func(t *testing.T) {
			_, err := readCsv("stopId,route")

			t.Run("it will return an error", func(t *testing.T) {
				assert.NotNil(t, err)
				assert.EqualValues(t, err.Error(), "Expected 4 headers, got 2")
			})
		})

		t.Run("when stopId is not a valid number", func(t *testing.T) {
			csv := `stopID,route,trainID,time
"not-a-number","C","865a","Jul 05 2021 13:14"`
			_, err := readCsv(csv)
			assert.NotNil(t, err)
		})

		t.Run("when trainID is not 4 characters", func(t *testing.T) {
			csv := `stopID,route,trainID,time
1,"C","865","Jul 05 2021 13:14"`
			_, err := readCsv(csv)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), "Train Id is invalid, too few characters: 865")
		})

		t.Run("when trainID is not alphanumeric", func(t *testing.T) {
			csv := `stopID,route,trainID,time
1,"C","a_b@","Jul 05 2021 13:14"`
			_, err := readCsv(csv)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), "Train Id must be alphanumeric: a_b@")
		})
	})

	t.Run("when readCsv is invoked with a valid csv", func(t *testing.T) {
		csv := `stopID,route,trainID,time
1,"C","865a","Jul 05 2021 13:14"
1,"55","465a","Jul 05 2021 14:14"`
		schedules := invokeRead(csv)

		t.Run("it will return a slice with 2 schedules", func(t *testing.T) {
			assert.Len(t, schedules, 2)
		})
	})

}

func TestInsertSchedules(t *testing.T) {
	initDb()
	defer tearDownDb()

	t.Run("when a csv is processed and inserted to Schedule db", func(t *testing.T) {
		csv := `stopID,route,trainID,time
1,"C","865a","Jul 05 2021 13:14"
1,"55","465a","Jul 05 2021 14:14"`

		s, _ := readCsv(csv)
		err := insertSchedules(s)
		if err != nil {
			t.Fail()
		}

		t.Run("when we query the Schedule db", func(t *testing.T) {
			all, err := db.ReadAll(fmt.Sprintf("./%s", scheduleDbName))
			if err != nil {
				t.Fail()
			}

			var schedules = []Schedule{}
			for _, b := range all {
				schedule := Schedule{}
				err := json.Unmarshal(b, &schedule)
				require.NoError(t, err)
				schedules = append(schedules, schedule)
			}

			t.Run("it should have 2 rows with expected ids", func(t *testing.T) {
				require.Len(t, schedules, 2)

				var (
					first  = schedules[0]
					second = schedules[1]
				)

				assert.EqualValues(t, first.ID, "0_865a")
				assert.EqualValues(t, second.ID, "1_465a")
			})
		})
	})
}

func TestGetTrainsByStopAndTime(t *testing.T) {
	initDb()
	defer tearDownDb()

	t.Run("given a database with Schedules", func(t *testing.T) {
		stopId := int64(1)
		csv := `stopID,route,trainID,time
1,"C","865a","Jul 04 2021 07:14"
1,"C","865a","Jul 04 2021 07:42"
1,"C","865a","Jul 04 2021 08:10"
1,"C","865a","Jul 04 2021 08:34"
1,"C","865a","Jul 04 2021 09:04"
1,"C","865a","Jul 04 2021 09:20"
1,"C","865a","Jul 04 2021 09:50"
1,"C","kpr5","Jul 04 2021 10:14"
1,"C","kpr5","Jul 04 2021 10:35"
1,"C","kpr5","Jul 04 2021 10:55"
1,"C","kpr5","Jul 04 2021 11:34"
1,"C","kpr5","Jul 04 2021 11:55"
1,"C","kpr5","Jul 04 2021 12:02"
1,"C","kpr5","Jul 04 2021 12:18"
1,"55","465a","Jul 04 2021 07:42"
1,"55","465a","Jul 04 2021 12:30"
1,"55","465a","Jul 04 2021 12:50"
1,"55","465a","Jul 04 2021 13:12"
1,"55","465a","Jul 04 2021 13:35"
1,"55","465a","Jul 04 2021 14:14"
1,"55","465a","Jul 04 2021 20:30"
1,"55","465a","Jul 04 2021 21:05"
1,"55","465a","Jul 04 2021 21:40"
1,"55","465a","Jul 04 2021 22:12"
1,"55","465a","Jul 04 2021 22:50"
1,"55","465a","Jul 04 2021 23:35"
1,"55","465a","Jul 05 2021 02:30"
1,"C","314p","Jul 05 2021 02:30"
`
		csvHandler(csv)

		t.Run("given getTrainsByStopAndTime is invoked", func(t *testing.T) {
			t.Run("when invoked with an invalid time", func(t *testing.T) {
				_, err := getTrainsByStopAndTime(1, "07/04/21 7:42")
				assert.Error(t, err)
				assert.EqualValues(t, err.Error(), `parsing time "07/04/21 7:42" as "Jan 02 2006 15:04": cannot parse "07/04/21 7:42" as "Jan"`)
			})

			t.Run("when invoked with a valid time where there are 2 trains arriving on that minute", func(t *testing.T) {
				schedules, err := getTrainsByStopAndTime(stopId, "Jul 04 2021 07:42")
				require.Nil(t, err)

				t.Run("it will return 2 rows", func(t *testing.T) {
					assert.Len(t, schedules, 2)

					var (
						first  = schedules[0]
						second = schedules[1]
					)

					assert.EqualValues(t, "C", first.Route)
					assert.EqualValues(t, "Jul 04 2021 07:42", first.Time)

					assert.EqualValues(t, "55", second.Route)
					assert.EqualValues(t, "Jul 04 2021 07:42", second.Time)
				})
			})

			t.Run("when invoked with a valid time outside of schedule range", func(t *testing.T) {
				schedules, err := getTrainsByStopAndTime(stopId, "Jul 04 2021 06:30")
				require.Nil(t, err)

				t.Run("it will return 0 rows", func(t *testing.T) {
					assert.Len(t, schedules, 0)
				})
			})

			t.Run("when invoked where there are no available trains that day", func(t *testing.T) {
				schedules, _ := getTrainsByStopAndTime(stopId, "Jul 04 2021 23:36")
				assert.Len(t, schedules, 2)
			})
		})

	})
}

func TestGetFirstMultipleTrainsByDate(t *testing.T) {
	t.Run("given a database with two applicable trains", func(t *testing.T) {
		initDb()
		defer tearDownDb()

		csv := `stopID,route,trainID,time
1,"55","465a","Jul 05 2021 01:30"
1,"C","314p","Jul 05 2021 02:30"
2,"21","159t","Jul 05 2021 02:30"
1,"21x","159t","Jul 05 2021 02:30"`
		csvHandler(csv)

		t.Run("when getFirstTrainsOfDay is invoked", func(t *testing.T) {
			stopID := int64(1)
			d, _ := time.Parse(layout, "Jul 05 2021 12:00")
			trains, err := getFirstTrainsOfDay(d, stopID)
			require.Nil(t, err)

			t.Run("it will return all trains for the given stop", func(t *testing.T) {
				assert.Len(t, trains, 2)

				for i := range trains {
					assert.EqualValues(t, trains[i].StopID, stopID)
				}
			})
		})

	})

	t.Run("given a database with no applicable trains", func(t *testing.T) {
		initDb()
		defer tearDownDb()

		csv := `stopID,route,trainID,time
1,"55","465a","Jul 05 2021 01:30"
2,"C","314p","Jul 05 2021 02:30"
3,"21","159t","Jul 05 2021 02:30"
1,"55","159t","Jul 05 2021 02:30"`
		csvHandler(csv)

		t.Run("when getFirstTrainsOfDay is invoked", func(t *testing.T) {
			stopID := int64(1)
			d, _ := time.Parse(layout, "Jul 05 2021 12:00")
			trains, err := getFirstTrainsOfDay(d, stopID)
			require.Nil(t, err)

			t.Run("there will be no trains returned", func(t *testing.T) {
				assert.Len(t, trains, 0)
			})
		})

	})

}
