package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"time"
)

func IntToBytes(n int) []byte {
	tmp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}

func BytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

func TCPread(tcpcon *net.TCPConn, readNum int, timeout time.Duration) ([]byte, error) {
	tcpcon.SetReadDeadline(time.Now().Add(timeout))
	result := make([]byte, readNum)
	byteRead, err := tcpcon.Read(result)
	if err != nil {
		return result, err
	}
	if byteRead != readNum {
		return result, errors.New("TCP read timout")
	}
	return result, nil
}

func TCPwr(networkName string, data []byte) ([]byte, error) {
	ip, err := net.ResolveTCPAddr("tcp", networkName)
	if err != nil {
		return []byte{}, err
	}
	tcpcon, err := net.DialTCP("tcp", nil, ip)
	if err != nil {
		return []byte{}, err
	}
	tcpcon.Write(data)

	result, err := TCPread(tcpcon, 7, 1000*time.Millisecond)
	if err != nil {
		return result, err
	}
	len := BytesToInt(result[3:7])

	result, err = TCPread(tcpcon, len, 1000*time.Millisecond)
	if err != nil {
		return result, err
	}
	return result, nil
}
