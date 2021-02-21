package main

import (
	"fmt"
	"testing"

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
"not-a-number","C","865a","13:14"`
			_, err := readCsv(csv)
			assert.NotNil(t, err)
		})

		t.Run("when trainID has less than", func(t *testing.T) {
			csv := `stopID,route,trainID,time
1,"C","865","13:14"`
			_, err := readCsv(csv)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), "Train Id is invalid, too few characters: 865")
		})
	})

	t.Run("when readCsv is invoked with an expected csv", func(t *testing.T) {
		csv := `stopID,route,trainID,time
1,"C","865a","13:14"
1,"55","465a","14:14"`
		schedules := invokeRead(csv)

		t.Run("it will return a slice with 2 schedules", func(t *testing.T) {
			assert.Len(t, schedules, 2)
		})
	})

}

func TestInsertSchedules(t *testing.T) {
	initDb()
	csv := `stopID,route,trainID,time
1,"C","865a","13:14"
1,"55","465a","14:14"`
	schedules, _ := readCsv(csv)
	err := insertSchedules(schedules)
	fmt.Println(err)

	if err != nil {
		t.Fail()
	}
}
