package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

var _myXsCommon *xsCommon

func setXsCommon(myXsCommon *xsCommon) {
	_myXsCommon = myXsCommon
}

func getXsCommon() *xsCommon {
	if _myXsCommon == nil {
		panicErr(errors.New("get nil xscommon"))
	}
	return _myXsCommon
}

type xsCommon struct {
	XMLName    xml.Name `xml:"common"`
	BuildingID string   `xml:"building_id"`
	GatewayID  string   `xml:"gateway_id"`
	Type       string   `xml:"type"`
}

type xsValidate struct {
	XMLName    xml.Name `xml:"root"`
	Common     xsCommon
	IDValidate struct {
		XMLName   xml.Name `xml:"id_validate"`
		Operation string   `xml:"operation,attr"`
		Sequence  string   `xml:"sequence"`
		Md5       string   `xml:"md5"`
		Result    string   `xml:"result"`
	}
}

type xsEnergyItem struct {
	XMLName xml.Name `xml:"energy_item"`
	Code    string   `xml:"code,attr"`
	Value   string   `xml:",innerxml"`
}

type xsMeter struct {
	XMLName  xml.Name `xml:"meter"`
	ID       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	Function struct {
		XMLName xml.Name `xml:"function"`
		ID      string   `xml:"id,attr"`
		Error   string   `xml:"error,attr"`
		Value   string   `xml:",innerxml"`
	}
}

type xsData struct {
	XMLName xml.Name `xml:"root"`
	Common  xsCommon
	Data    struct {
		XMLName     xml.Name `xml:"data"`
		Operation   string   `xml:"operation,attr"`
		Time        string   `xml:"time"`
		EnergyItems struct {
			XMLName xml.Name `xml:"energy_items"`
			Items   []xsEnergyItem
		}
		Meters struct {
			XMLName xml.Name `xml:"meters"`
			Items   []xsMeter
		}
	}
}

type xsDataAck struct {
	XMLName xml.Name `xml:"root"`
	Common  xsCommon
	Data    struct {
		XMLName xml.Name `xml:"data"`
		Time    string   `xml:"time"`
		Ack     string   `xml:"ack"`
	}
}

func xsMarshal(xmlstruct interface{}) ([]byte, error) {
	text, err := xml.MarshalIndent(xmlstruct, "", "    ")
	if err != nil {
		return []byte(""), err
	}
	text = bytesCombine([]byte(`<?xml version="1.0" encoding="UTF-8"?>`+"\n"), text)
	return text, nil
}

func addValidateHead(data []byte) []byte {
	return bytesCombine([]byte("\x1F\x1F\x01"), intToBytes(len(data)), data)
}

func addDataHead(data []byte) []byte {
	return bytesCombine([]byte("\x1F\x1F\x03"), intToBytes(len(data)), data)
}

func sendXsValidate(host string, xs *xsValidate) ([]byte, error) {
	text, err := xsMarshal(xs)
	panicErr(err)

	// log.Println("TCP write:\n" + string(text))
	text = addValidateHead(text)
	text, err = tcpRW(host, text)
	if err != nil {
		return nil, err
	}
	// log.Println("TCP read:\n" + string(text))
	return text, nil
}

func validate(dataCenter *dataCenterStruct) error {

	myXsValidate := xsValidate{}
	myXsValidate.Common = *getXsCommon()
	myXsValidate.Common.Type = "id_validate"
	myXsValidate.IDValidate.Operation = "request"

	host := dataCenter.IP + ":" + dataCenter.PORT
	text, err := sendXsValidate(host, &myXsValidate)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(text, &myXsValidate)
	panicErr(err)

	sequence := myXsValidate.IDValidate.Sequence
	myMd5 := md5.Sum(bytesCombine([]byte(dataCenter.UploadSecretKey), []byte(sequence)))
	myXsValidate.IDValidate.Md5 = fmt.Sprintf("%x", myMd5)
	myXsValidate.IDValidate.Operation = "md5"

	text, err = sendXsValidate(host, &myXsValidate)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(text, &myXsValidate)
	panicErr(err)

	if string(myXsValidate.IDValidate.Result) == "pass" {
		log.Println("id validate: success")
		return nil
	}
	return errors.New("id validate not getting the correct reply")
}

func sendData(secret string, myUploadData *uploadData) error {

	myXsData := xsData{}
	myXsData.Data.Operation = "report"
	myXsData.Data.Time = myUploadData.Time
	myXsData.Common = *getXsCommon()
	myXsData.Common.Type = "energy_data"

	myXsData.Data.EnergyItems.Items, myXsData.Data.Meters.Items = createXs(myUploadData)

	text, err := xsMarshal(myXsData)
	panicErr(err)

	// log.Println("TCP write:\n" + string(text))
	text, err = bemsUploadEncrypt(text)
	panicErr(err)

	text = addDataHead(text)

	text, err = tcpRW("hncj1.yeep.net.cn:7201", text)
	if err != nil {
		return err
	}
	// log.Println("TCP read:\n" + string(text))

	myXsDataAck := xsDataAck{}
	err = xml.Unmarshal(text, &myXsDataAck)
	panicErr(err)

	if string(myXsDataAck.Data.Ack) == "OK" {
		log.Println("send data: success")
		return nil
	}
	return errors.New("send data not getting the correct reply")
}

func getData() *uploadData {
	db, err := sql.Open("sqlite3", runstatePath)
	panicErr(err)

	rows, err := db.Query("SELECT * FROM BemsUploadData LIMIT 1")
	panicErr(err)

	db.Close()

	var rlt *uploadData
	for rows.Next() {
		var Time string
		var data string
		err = rows.Scan(&Time, &data)
		myUploadData := uploadData{}
		myUploadData.Time = Time
		err := json.Unmarshal([]byte(data), &myUploadData.Meters)
		panicErr(err)
		rlt = &myUploadData
	}

	if rlt != nil {
		myXsCommon := xsCommon{}
		meterID := rlt.Meters[0].ID
		len := len(meterID)
		myXsCommon.BuildingID = meterID[0 : len-6]
		myXsCommon.GatewayID = meterID[len-6 : len-4]
		_myXsCommon = &myXsCommon
	}
	return rlt
}

func createXs(myUploadData *uploadData) ([]xsEnergyItem, []xsMeter) {
	myMeters := []xsMeter{}
	itemMap := make(map[string]float64)
	for _, meter := range myUploadData.Meters {
		myMeter := xsMeter{}
		myMeter.ID = meter.ID[len(meter.ID)-4 : len(meter.ID)]
		myMeter.Name = meter.Name
		myMeter.Function.ID = meter.FunctionID
		myMeter.Function.Value = meter.Value
		myMeters = append(myMeters, myMeter)

		meterValue, err := strconv.ParseFloat(meter.Value, 64)
		panicErr(err)
		value, ok := itemMap[meter.EnergyItem]
		if ok {
			itemMap[meter.EnergyItem] = meterValue + value
		} else {
			itemMap[meter.EnergyItem] = meterValue
		}
	}

	myEnergyItems := []xsEnergyItem{}

	for k, v := range itemMap {
		myXsEnergyItem := xsEnergyItem{}
		myXsEnergyItem.Code = k
		myXsEnergyItem.Value = strconv.FormatFloat(v, 'f', -1, 64)
		myEnergyItems = append(myEnergyItems, myXsEnergyItem)
	}

	return myEnergyItems, myMeters
}

func deleteData(myUploadData *uploadData) {

	db, err := sql.Open("sqlite3", runstatePath)
	panicErr(err)

	cmd := "DELETE FROM bemsUploadData where CreatTime == " + myUploadData.Time

	log.Println("delete data create by " + myUploadData.Time)

	_, err = db.Exec(cmd)

	panicErr(err)

	db.Close()
}

func uploadToDataCenter(dataCenter *dataCenterStruct) error {

	myUploadData := getData()
	if myUploadData == nil {
		return errors.New("no data found in BemsUploadData")
	}
	log.Printf("found data :%+v", myUploadData)

	err := validate(dataCenter)
	if err != nil {
		return err
	}

	err = sendData(dataCenter.AESVector, myUploadData)
	if err != nil {
		return err
	}

	deleteData(myUploadData)
	return nil
}

func uploadTask() {
	configure := getConfigure()

	if len(configure.DataCenter) < 1 {
		panicErr(errors.New("There's no data center"))
	} else {
		for {
			err := uploadToDataCenter(&configure.DataCenter[0])
			if err != nil {
				log.Println(err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}
