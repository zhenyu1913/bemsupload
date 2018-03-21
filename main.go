package main

import (
	"crypto/md5"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
)

type XsCommon struct {
	XMLName     xml.Name `xml:"common"`
	Building_id string   `xml:"building_id"`
	Gateway_id  string   `xml:"gateway_id"`
	Type        string   `xml:"type"`
}

type XsValidate struct {
	XMLName     xml.Name `xml:"root"`
	Common      XsCommon
	Id_validate struct {
		XMLName   xml.Name `xml:"id_validate"`
		Operation string   `xml:"operation,attr"`
		Sequence  string   `xml:"sequence"`
		Md5       string   `xml:"md5"`
		Result    string   `xml:"result"`
	}
}

type XsEnergyItem struct {
	XMLName xml.Name `xml:"energy_item"`
	Code    string   `xml:"code,attr"`
	Value   string   `xml:",innerxml"`
}

type XsMeter struct {
	XMLName  xml.Name `xml:"meter"`
	Id       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	Function struct {
		XMLName xml.Name `xml:"function"`
		Id      string   `xml:"id,attr"`
		Error   string   `xml:"error,attr"`
		Value   string   `xml:",innerxml"`
	}
}

type XsData struct {
	XMLName xml.Name `xml:"root"`
	Common  XsCommon
	Data    struct {
		XMLName     xml.Name `xml:"data"`
		Operation   string   `xml:"operation,attr"`
		Time        string   `xml:"time"`
		EnergyItems struct {
			XMLName xml.Name `xml:"energy_items"`
			Items   []XsEnergyItem
		}
		Meters struct {
			XMLName xml.Name `xml:"meters"`
			Items   []XsMeter
		}
	}
}

type XsDataAck struct {
	XMLName xml.Name `xml:"root"`
	Common  XsCommon
	Data    struct {
		XMLName xml.Name `xml:"data"`
		Time    string   `xml:"time"`
		Ack     string   `xml:"ack"`
	}
}

func getXsCommon() XsCommon {
	xsCommon := XsCommon{}
	xsCommon.Building_id = "JD310114BG0091"
	xsCommon.Gateway_id = "01"
	return xsCommon
}

func XsMarshal(xmlstruct interface{}) ([]byte, error) {
	text, err := xml.MarshalIndent(xmlstruct, "", "    ")
	if err != nil {
		return []byte(""), err
	}
	text = BytesCombine([]byte(`<?xml version="1.0" encoding="UTF-8"?>`+"\n"), text)
	return text, nil
}

func AddValidateHead(data []byte) []byte {
	return BytesCombine([]byte("\x1F\x1F\x01"), IntToBytes(len(data)), data)
}

func AddDataHead(data []byte) []byte {
	return BytesCombine([]byte("\x1F\x1F\x03"), IntToBytes(len(data)), data)
}

func SendXsValidate(xs XsValidate) ([]byte, error) {
	text, err := XsMarshal(xs)
	if err != nil {
		panic(err)
	}

	log.Println("TCP write:\n" + string(text))
	text = AddValidateHead(text)
	text, err = TCPwr("hncj1.yeep.net.cn:7201", text)
	if err != nil {
		return nil, err
	}
	log.Println("TCP read:\n" + string(text))
	return text, nil
}

func Validate(secret string) error {

	xsValidate := XsValidate{}
	xsValidate.Common = getXsCommon()
	xsValidate.Common.Type = "id_validate"
	xsValidate.Id_validate.Operation = "request"

	text, err := SendXsValidate(xsValidate)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(text, &xsValidate)
	if err != nil {
		panic(err)
	}

	sequence := xsValidate.Id_validate.Sequence
	myMd5 := md5.Sum(BytesCombine([]byte(secret), []byte(sequence)))
	xsValidate.Id_validate.Md5 = fmt.Sprintf("%x", myMd5)
	xsValidate.Id_validate.Operation = "md5"

	text, err = SendXsValidate(xsValidate)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(text, &xsValidate)
	if err != nil {
		panic(err)
	}

	if string(xsValidate.Id_validate.Result) == "pass" {
		log.Println("id validate: success")
		return nil
	}
	return errors.New("id validate not getting the correct reply")
}

func SendData(secret string, energyItems []XsEnergyItem, meters []XsMeter) error {

	xsData := XsData{}
	xsData.Data.Operation = "report"
	xsData.Data.Time = "20170908010101"
	xsData.Common = getXsCommon()
	xsData.Common.Type = "energy_data"

	xsData.Data.EnergyItems.Items = energyItems
	xsData.Data.Meters.Items = meters

	text, err := XsMarshal(xsData)
	if err != nil {
		panic(err)
	}

	log.Println("TCP write:\n" + string(text))
	text, err = BemsUploadEncrypt(text)
	if err != nil {
		panic(err)
	}
	text = AddDataHead(text)

	text, err = TCPwr("hncj1.yeep.net.cn:7201", text)
	if err != nil {
		return err
	}
	log.Println("TCP read:\n" + string(text))

	xsDataAck := XsDataAck{}
	err = xml.Unmarshal(text, &xsDataAck)
	if err != nil {
		panic(err)
	}

	if string(xsDataAck.Data.Ack) == "OK" {
		log.Println("send data: success")
		return nil
	}
	return errors.New("send data not getting the correct reply")
}

func sendData() error {
	secret := "useruseruseruser"

	err := Validate(secret)
	if err != nil {
		return err
	}

	energyItem := XsEnergyItem{}
	energyItem.Value = "1"
	energyItem.Code = "01000"

	meter := XsMeter{}
	meter.Id = "A001"
	meter.Name = "1号电表"
	meter.Function.Id = "WPP"
	meter.Function.Value = "2"

	energyItems := []XsEnergyItem{energyItem}
	meters := []XsMeter{meter}
	err = SendData(secret, energyItems, meters)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	sendData()
}
