package main

import (
    "fmt"
    "net"
    "time"
    "encoding/xml"
    "errors"
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

func TCPread(tcpcon *net.TCPConn, readNum int, timeout time.Duration) ([]byte, error) {
    tcpcon.SetReadDeadline(time.Now().Add(timeout))
    result := make([]byte,readNum)
    byteRead, err := tcpcon.Read(result)
    if err != nil {
        return result, err
    }
    if byteRead != readNum {
        return result, errors.New("TCP read timout")
    }
    return result, nil
}

func TCPwr(networkName string,data []byte) ([]byte, error) {
    ip, err := net.ResolveTCPAddr("tcp",networkName)
    if err != nil {
        return []byte{}, err
    }
    tcpcon, err := net.DialTCP("tcp", nil, ip)
    if err != nil {
        return []byte{}, err
    }
    tcpcon.Write(data)

    result, err := TCPread(tcpcon, 7, 1000 * time.Millisecond)
    if err != nil {
        return result, err
    }
    len := BytesToInt(result[3:7])

    result, err = TCPread(tcpcon, len, 1000 * time.Millisecond)
    if err != nil {
        return result, err
    }
    return result, nil
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

    // fmt.Println("TCP write:\n" + string(xmltext))
    xmltext = BmesUploadAddHead(xmltext)
    xmltext, err = TCPwr("hncj1.yeep.net.cn:7201",xmltext)
    if err != nil {
        panic(err)
    }
    // fmt.Println("TCP read:\n" + string(xmltext))

    err = xml.Unmarshal(xmltext,&xmlstruct)
    if err != nil {
        panic(err)
    }

    secret := []byte("useruseruseruser")
    sequence := xmlstruct.Id_validate.Sequence
    myMd5 := md5.Sum(BytesCombine(secret,[]byte(sequence)))
    xmlstruct.Id_validate.Md5 = fmt.Sprintf("%x", myMd5)
    xmlstruct.Id_validate.Operation = "md5"

    xmltext, err = BemsUploadMarshal(xmlstruct)
    if err != nil {
        panic(err)
    }

    // fmt.Println("TCP write:\n" + string(xmltext))
    xmltext = BmesUploadAddHead(xmltext)
    xmltext, err = TCPwr("hncj1.yeep.net.cn:7201",xmltext)
    if err != nil {
        panic(err)
    }
    // fmt.Println("TCP read:\n" + string(xmltext))

    err = xml.Unmarshal(xmltext,&xmlstruct)
    if err != nil {
        panic(err)
    }
    fmt.Println("id validate result: " + string(xmlstruct.Id_validate.Result))
}
