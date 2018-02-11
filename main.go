package main

import (
    "fmt"
    "net"
    "time"
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

func TCPwr(networkName string,data []byte) ([]byte, error) {
    result := []byte{}
    ip, err := net.ResolveTCPAddr("tcp",networkName)
    if err != nil {
        return result, err
    }
    tcpcon, err := net.DialTCP("tcp", nil, ip)
    if err != nil {
        return result, err
    }
    tcpcon.Write(data)
    time.Sleep(100 * time.Millisecond)
    _, err = tcpcon.Read(result)
    if err != nil {
        return result, err
    }
    return result, nil
}

func ShaihaiBemsTCPgetData(data []byte) []byte {
    len := BytesToInt(data[3:7])
    return data[7:len+7]
}

func main() {
    secret := []byte("useruseruseruser")
    xmlstruct := XmlStruct{}
    xmlstruct.Common.Building_id = "JD310114BG0091"
    xmlstruct.Common.Gateway_id = "01"
    xmlstruct.Common.Type = "id_validate"
    xmlstruct.Id_validate.Operation = "request"

    xmltext, err := BemsUploadMarshal(xmlstruct)
    if err != nil {
        panic(err)
    }

    xmltext = BmesUploadAddHead(xmltext)
    fmt.Println("TCP write:\n" + fmt.Sprintf("%x",xmltext))
    xmltext, err = TCPwr("hncj1.yeep.net.cn:7201",xmltext)
    if err != nil {
        panic(err)
    }
    fmt.Println("TCP read:\n" + fmt.Sprintf("%x",xmltext))

    xmltext = ShaihaiBemsTCPgetData(xmltext)

    err = xml.Unmarshal(xmltext,&xmlstruct)
    if err != nil {
        panic(err)
    }


    sequence := xmlstruct.Id_validate.Sequence
    myMd5 := md5.Sum(BytesCombine(secret,[]byte(sequence)))
    xmlstruct.Id_validate.Md5 = fmt.Sprintf("%x", myMd5)
    xmlstruct.Id_validate.Operation = "md5"

    xmltext, err = BemsUploadMarshal(xmlstruct)
    if err != nil {
        panic(err)
    }

    fmt.Println("TCP write:\n" + string(xmltext))
    xmltext = BmesUploadAddHead(xmltext)
    xmltext, err = TCPwr("hncj1.yeep.net.cn:7201",xmltext)
    if err != nil {
        panic(err)
    }
    xmltext = ShaihaiBemsTCPgetData(xmltext)
    fmt.Println("TCP read:\n" + string(xmltext))

    err = xml.Unmarshal(xmltext,&xmlstruct)
    if err != nil {
        panic(err)
    }

    fmt.Println(xmlstruct)
}
