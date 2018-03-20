package main

import (
    "fmt"
    "encoding/xml"
    "crypto/md5"
)

type XmlCommon struct {
    XMLName xml.Name `xml:"common"`
    Building_id string `xml:"building_id"`
    Gateway_id string `xml:"gateway_id"`
    Type string `xml:"type"`
}

type Id_validate_quiest struct {
    Common XmlCommon
}

type XmlValidate struct {
    XMLName xml.Name `xml:"root"`
    Common XmlCommon
    Id_validate struct {
        XMLName xml.Name `xml:"id_validate"`
        Operation string `xml:"operation,attr"`
        Sequence string `xml:"sequence"`
        Md5 string `xml:"md5"`
        Result string `xml:"result"`
    }
}

type XmlEnergyItem struct {
    XMLName xml.Name `xml:"energy_item"`
    Code string `xml:"code,attr"`
    Value string `xml:",innerxml"`
}

type XmlMeter struct {
    XMLName xml.Name `xml:"meter"`
    Id string `xml:"id,attr"`
    Name string `xml:"name,attr"`
    Function struct {
        XMLName xml.Name `xml:"function"`
        Id string `xml:"id,attr"`
        Error string `xml:"error,attr"`
        Value string `xml:",innerxml"`
    }
}

type XmlData struct {
    XMLName xml.Name `xml:"root"`
    Common XmlCommon
    Data struct {
        XMLName xml.Name `xml:"data"`
        Operation string `xml:"operation,attr"`
        Time string `xml:"time"`
        Energy_items struct {
            XMLName xml.Name `xml:"energy_items"`
            Items []XmlEnergyItem
        }
        Meters struct {
            XMLName xml.Name `xml:"meters"`
            Items []XmlMeter
        }
    }
}


func BmesUploadAddHead(data []byte) []byte {
    return BytesCombine([]byte("\x1F\x1F\x01"),IntToBytes(len(data)),data)
}

func BmesUploadAddDataHead(data []byte) []byte {
    return BytesCombine([]byte("\x1F\x1F\x03"),IntToBytes(len(data)),data)
}

func BemsUploadSendXml(xml interface{}) []byte {
    xmltext, err := BemsUploadMarshal(xml)
    if err != nil {
        panic(err)
    }

    fmt.Println("TCP write:\n" + string(xmltext))
    xmltext = BmesUploadAddHead(xmltext)
    xmltext, err = TCPwr("hncj1.yeep.net.cn:7201",xmltext)
    if err != nil {
        panic(err)
    }
    fmt.Println("TCP read:\n" + string(xmltext))
    return xmltext
}

func main() {
    xmlstruct := XmlValidate{}
    xmlstruct.Common.Building_id = "JD310114BG0091"
    xmlstruct.Common.Gateway_id = "01"
    xmlstruct.Common.Type = "id_validate"
    xmlstruct.Id_validate.Operation = "request"

    xmltext := BemsUploadSendXml(xmlstruct)

    err := xml.Unmarshal(xmltext,&xmlstruct)
    if err != nil {
        panic(err)
    }

    secret := []byte("useruseruseruser")
    sequence := xmlstruct.Id_validate.Sequence
    myMd5 := md5.Sum(BytesCombine(secret,[]byte(sequence)))
    xmlstruct.Id_validate.Md5 = fmt.Sprintf("%x", myMd5)
    xmlstruct.Id_validate.Operation = "md5"

    xmltext = BemsUploadSendXml(xmlstruct)

    err = xml.Unmarshal(xmltext,&xmlstruct)
    if err != nil {
        panic(err)
    }

    fmt.Println("id validate result: " + string(xmlstruct.Id_validate.Result))

    dataStruct := XmlData{}
    dataStruct.Data.Operation = "report"
    dataStruct.Data.Time = "20170908010101"
    dataStruct.Common = xmlstruct.Common
    dataStruct.Common.Type = "energy_data"

    energy_item := XmlEnergyItem{}
    energy_item.Value = "1"
    energy_item.Code = "01000"
    meter := XmlMeter{}
    meter.Id = "A001"
    meter.Name = "1号电表"
    meter.Function.Id = "WPP"
    meter.Function.Value = "2"
    dataStruct.Data.Energy_items.Items = append(dataStruct.Data.Energy_items.Items, energy_item)
    dataStruct.Data.Meters.Items = append(dataStruct.Data.Meters.Items, meter)


    xmltext, err = BemsUploadMarshal(dataStruct)
    if err != nil {
        panic(err)
    }

    fmt.Println("TCP write:\n" + string(xmltext))
    xmltext, err = BemsUploadEncrypt(xmltext)
    if err != nil {
        panic(err)
    }
    xmltext = BmesUploadAddDataHead(xmltext)

    xmltext, err = TCPwr("hncj1.yeep.net.cn:7201",xmltext)
    if err != nil {
        panic(err)
    }
    fmt.Println("TCP read:\n" + string(xmltext))
}
