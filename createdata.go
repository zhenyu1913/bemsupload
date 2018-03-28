package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type configureStruct struct {
	DataCenter []struct {
		AESKey          string
		AESVector       string
		DCID            string
		Freq            int
		IP              string
		PORT            string
		UploadSecretKey string
	}
	MeterItem []struct {
		EnergyItemCode string
		FieldName      string
		FuncID         string
		MeterCode      string
		MeterCodeAlia  string
		TableName      string
		Remark         string
	}
}

type meter struct {
	ID         string
	Name       string
	FunctionID string
	Error      string
	Value      string
	EnergyItem string
}

type uploadData struct {
	Time   string
	Meters []meter
}

func getMeterMap() map[string]string {

	db, err := sql.Open("sqlite3", runstatePath)
	if err != nil {
		log.Panic(err)
	}

	rows, err := db.Query("SELECT * FROM MeterState")
	if err != nil {
		log.Panic(err)
	}

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

func getConfigure() *configureStruct {
	fileByteArray, err := ioutil.ReadFile(configurePath)
	if err != nil {
		log.Panic(err)
	}

	myConfigureStruct := configureStruct{}
	err = json.Unmarshal(fileByteArray, &myConfigureStruct)
	if err != nil {
		log.Panic(err)
	}

	return &myConfigureStruct
}

func panicErr(err error) {
	if err != nil {
		log.Panic(err)
	}
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

func createData() {
	myConfigureStruct := getConfigure()
	log.Printf("read getConfigure :%+v", myConfigureStruct)
	myMeterMap := getMeterMap()

	log.Println("read meter map :", myMeterMap)

	myMeters := []meter{}
	for _, meterItem := range myConfigureStruct.MeterItem {
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
		saveData(&myUploadData)
	}
}
