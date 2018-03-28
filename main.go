package main

import (
	"encoding/json"
	"io/ioutil"
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

func getConfigure() *configureStruct {
	fileByteArray, err := ioutil.ReadFile(configurePath)
	panicErr(err)

	myConfigureStruct := configureStruct{}
	err = json.Unmarshal(fileByteArray, &myConfigureStruct)
	panicErr(err)

	return &myConfigureStruct
}

func main() {
	// err := upload()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	createData()
}
