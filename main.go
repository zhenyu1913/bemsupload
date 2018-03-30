package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"runtime/debug"
)

type dataCenterStruct struct {
	AESKey          string
	AESVector       string
	DCID            string
	Freq            int
	IP              string
	PORT            string
	UploadSecretKey string
}

type configureStruct struct {
	DataCenter []dataCenterStruct
	MeterItem  []struct {
		EnergyItemCode string
		FieldName      string
		FuncID         string
		MeterCode      string
		MeterCodeAlia  string
		TableName      string
		Remark         string
	}
}

type uploadData struct {
	Time   string
	Meters []meter
}
type meter struct {
	ID         string
	Name       string
	FunctionID string
	Error      string
	Value      string
	EnergyItem string
}

func getConfigure() *configureStruct {
	fileByteArray, err := ioutil.ReadFile(configurePath)
	panicErr(err)

	myConfigureStruct := configureStruct{}
	err = json.Unmarshal(fileByteArray, &myConfigureStruct)
	panicErr(err)

	return &myConfigureStruct
}

func writeDCMLog(s string) {

}

func protectGo(f func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Fatal Error !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			log.Println(err)
			log.Println(debug.Stack())
		}
	}()
	f()
}

func main() {
	go protectGo(createTask)
	go protectGo(uploadTask)
	for {

	}
}
