package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
	"time"
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

func protectGo(f func()) {
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Println(string(debug.Stack()))
				}
			}()
			f()
		}()
		time.Sleep(3 * time.Minute)
	}
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "once" {
			configure := getConfigure()
			createData(configure)
			err := uploadToDataCenter(&configure.DataCenter[0])
			if err != nil {
				log.Println(err)
			}
			os.Exit(0)
		}
	}
	go protectGo(createTask)
	go protectGo(uploadTask)
	for {
		time.Sleep(1 * time.Second)
	}
}
