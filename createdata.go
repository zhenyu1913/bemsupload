package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func getMeterMap() map[string]string {

	db, err := sql.Open("sqlite3", runstatePath)
	panicErr(err)

	rows, err := db.Query("SELECT * FROM MeterState")
	panicErr(err)

	db.Close()

	meterMap := make(map[string]string)
	for rows.Next() {
		var MeterCode string
		var ErrCount int
		var data string
		err = rows.Scan(&MeterCode, &ErrCount, &data)
		if ErrCount == 0 && data != "" {
			meterMap[MeterCode] = data
		}
	}

	return meterMap
}

func deleteDataOneMonthAgo() {

	db, err := sql.Open("sqlite3", runstatePath)
	panicErr(err)

	now := time.Now()
	d, err := time.ParseDuration("-720h")
	panicErr(err)

	oneMonthBefore := now.Add(d).Format("20060102150405")
	cmd := "DELETE FROM BemsUploadData WHERE CreatTime < " + oneMonthBefore

	_, err = db.Exec(cmd)
	panicErr(err)

	db.Close()
}

func saveData(myUploadData *uploadData) {

	db, err := sql.Open("sqlite3", runstatePath)
	panicErr(err)

	tx, err := db.Begin()
	panicErr(err)

	stmt, err := tx.Prepare("REPLACE INTO BemsUploadData values(?,?)")
	panicErr(err)

	dataJSON, err := json.Marshal(myUploadData.Meters)
	panicErr(err)

	_, err = stmt.Exec(myUploadData.Time, dataJSON)
	panicErr(err)

	err = tx.Commit()
	panicErr(err)

	db.Close()
}

func createTask() {
	configure := getConfigure()
	log.Printf("read configure :%+v", configure)
	interval := int64(configure.DataCenter[0].Freq)
	var doneTime int64
	for {
		now := time.Now().Unix() / 60
		if now%interval == 0 && now != doneTime {
			createData(configure)
			doneTime = now
		}
		time.Sleep(5 * time.Second)
	}
}

func createData(configure *configureStruct) {
	myMeterMap := getMeterMap()

	myMeters := []meter{}
	for _, meterItem := range configure.MeterItem {
		data, ok := myMeterMap[meterItem.MeterCode]
		myMeter := meter{}
		myMeter.ID = meterItem.MeterCodeAlia
		myMeter.FunctionID = meterItem.FuncID
		myMeter.Name = meterItem.Remark
		myMeter.EnergyItem = meterItem.EnergyItemCode
		if ok {
			var dataMap map[string]string
			json.Unmarshal([]byte(data), &dataMap)
			fieldValue, ok := dataMap[meterItem.FieldName]
			if ok {
				myMeter.Error = ""
				myMeter.Value = fieldValue
			}
		} else {
			myMeter.Error = "meter not vaild"
		}
		myMeters = append(myMeters, myMeter)
	}
	if len(myMeters) >= 1 {
		myUploadData := uploadData{}

		myUploadData.Time = time.Now().Format("20060102150405")
		myUploadData.Meters = myMeters

		log.Printf("make data :%+v", myUploadData)
		deleteDataOneMonthAgo()
		saveData(&myUploadData)
	}
}
