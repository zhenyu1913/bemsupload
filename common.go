package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"runtime/debug"
	"time"
)

func intToBytes(n int) []byte {
	tmp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func bytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}

func bytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

func tcpRead(tcpcon *net.TCPConn, readNum int, timeout time.Duration) ([]byte, error) {
	tcpcon.SetReadDeadline(time.Now().Add(timeout))
	result := make([]byte, readNum)
	byteRead, err := tcpcon.Read(result)
	if err != nil {
		return result, err
	}
	if byteRead != readNum {
		return result, errors.New("TCP read timout" + "\n" + string(debug.Stack()))
	}
	return result, nil
}

func tcpRW(networkName string, data []byte) ([]byte, error) {
	ip, err := net.ResolveTCPAddr("tcp", networkName)
	if err != nil {
		return []byte{}, errors.New(err.Error() + "\n" + string(debug.Stack()))
	}
	tcpcon, err := net.DialTCP("tcp", nil, ip)
	if err != nil {
		return []byte{}, errors.New(err.Error() + "\n" + string(debug.Stack()))
	}
	tcpcon.Write(data)

	result, err := tcpRead(tcpcon, 7, 1000*time.Millisecond)
	if err != nil {
		return result, errors.New(err.Error() + "\n" + string(debug.Stack()))
	}
	len := bytesToInt(result[3:7])

	result, err = tcpRead(tcpcon, len, 1000*time.Millisecond)
	if err != nil {
		return result, errors.New(err.Error() + "\n" + string(debug.Stack()))
	}
	return result, nil
}

func panicErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}
