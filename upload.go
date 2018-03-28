package main

import (
	"crypto/md5"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
)

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

func getXsCommon() xsCommon {
	myXsCommon := xsCommon{}
	myXsCommon.BuildingID = "JD310114BG0091"
	myXsCommon.GatewayID = "01"
	return myXsCommon
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

func sendXsValidate(xs xsValidate) ([]byte, error) {
	text, err := xsMarshal(xs)
	panicErr(err)

	log.Println("TCP write:\n" + string(text))
	text = addValidateHead(text)
	text, err = tcpRW("hncj1.yeep.net.cn:7201", text)
	if err != nil {
		return nil, err
	}
	log.Println("TCP read:\n" + string(text))
	return text, nil
}

func validate(secret string) error {

	myXsValidate := xsValidate{}
	myXsValidate.Common = getXsCommon()
	myXsValidate.Common.Type = "id_validate"
	myXsValidate.IDValidate.Operation = "request"

	text, err := sendXsValidate(myXsValidate)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(text, &myXsValidate)
	panicErr(err)

	sequence := myXsValidate.IDValidate.Sequence
	myMd5 := md5.Sum(bytesCombine([]byte(secret), []byte(sequence)))
	myXsValidate.IDValidate.Md5 = fmt.Sprintf("%x", myMd5)
	myXsValidate.IDValidate.Operation = "md5"

	text, err = sendXsValidate(myXsValidate)
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

func sendData(secret string, energyItems []xsEnergyItem, meters []xsMeter) error {

	myXsData := xsData{}
	myXsData.Data.Operation = "report"
	myXsData.Data.Time = "20170908010101"
	myXsData.Common = getXsCommon()
	myXsData.Common.Type = "energy_data"

	myXsData.Data.EnergyItems.Items = energyItems
	myXsData.Data.Meters.Items = meters

	text, err := xsMarshal(myXsData)
	panicErr(err)

	log.Println("TCP write:\n" + string(text))
	text, err = bemsUploadEncrypt(text)
	panicErr(err)

	text = addDataHead(text)

	text, err = tcpRW("hncj1.yeep.net.cn:7201", text)
	if err != nil {
		return err
	}
	log.Println("TCP read:\n" + string(text))

	myXsDataAck := xsDataAck{}
	err = xml.Unmarshal(text, &myXsDataAck)
	panicErr(err)

	if string(myXsDataAck.Data.Ack) == "OK" {
		log.Println("send data: success")
		return nil
	}
	return errors.New("send data not getting the correct reply")
}

func upload() error {
	configure := getConfigure()

	for _, dataCenter := range configure.DataCenter {

		err := validate(dataCenter.UploadSecretKey)
		if err != nil {
			return err
		}

		myEnergyItem := xsEnergyItem{}
		myEnergyItem.Value = "1"
		myEnergyItem.Code = "01000"

		myMeter := xsMeter{}
		myMeter.ID = "A001"
		myMeter.Name = "1号电表"
		myMeter.Function.ID = "WPP"
		myMeter.Function.Value = "2"

		myEnergyItems := []xsEnergyItem{myEnergyItem}
		myMeters := []xsMeter{myMeter}
		err = sendData(dataCenter.AESVector, myEnergyItems, myMeters)
		if err != nil {
			return err
		}

	}
	return nil

}
