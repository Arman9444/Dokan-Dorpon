package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"
)
//global variable to store pretend date
var demoDay time.Time

func todayTime() time.Time{
	if demoDay.IsZero(){
		return time.Now()
	}
    return demoDay
}

func startDemo(date string) string {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Fatal("-today must look like 2026-05-14")
	}
	demoDay = day

	fileName := filepath.Join(os.TempDir(), "dokan-dorpon-demo.db")
	err = os.Remove(fileName)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
	fmt.Println("DEMO MODE: pretending today is", date)
	return fileName
}



