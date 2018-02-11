package main

import (
    "fmt"
    "net"
    "time"
    "encoding/xml"
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

type XmlStruct struct {
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

func BmesUploadAddHead(data []byte) []byte {
    return BytesCombine([]byte("\x1F\x1F\x01"),IntToBytes(len(data)),data)
}

func main() {
    xmlstruct := XmlStruct{}
    xmlstruct.Common.Building_id = "JD310114BG0091"
    xmlstruct.Common.Gateway_id = "01"
    xmlstruct.Common.Type = "id_validate"
    xmlstruct.Id_validate.Operation = "request"

    xmltext, err := BemsUploadMarshal(xmlstruct)
    if err != nil {
        panic(err)
    }
    // fmt.Println(string(xmltext))
    xmltext = BmesUploadAddHead(xmltext)
    // fmt.Printf("%x\n",xmltext)
    ip, err := net.ResolveTCPAddr("tcp","hncj1.yeep.net.cn:7201")
    if err != nil {
        fmt.Println(err)
    }
    tcpcon, err := net.DialTCP("tcp", nil, ip)
    if err != nil {
        fmt.Println(err)
    }
    tcpcon.Write(xmltext)
    time.Sleep(100 * time.Millisecond)
    byteread, err := tcpcon.Read(xmltext)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(byteread)
    len := BytesToInt(xmltext[3:7])
    xmltext = xmltext[7:len+7]
    fmt.Println(string(xmltext))

    err = xml.Unmarshal(xmltext,&xmlstruct)
    if err != nil {
        panic(err)
    }
    sequence := xmlstruct.Id_validate.Sequence
    _ = sequence
    fmt.Println(xmlstruct)
}
