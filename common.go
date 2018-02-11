package main

import (
    "bytes"
    "encoding/binary"
)

func IntToBytes(n int) []byte {
    tmp := int32(n)
    bytesBuffer := bytes.NewBuffer([]byte{})
    binary.Write(bytesBuffer, binary.BigEndian, tmp)
    return bytesBuffer.Bytes()
}

func BytesToInt(b []byte) int  {
    bytesBuffer := bytes.NewBuffer(b)
    var x int32
    binary.Read(bytesBuffer, binary.BigEndian, &x)
    return int(x)
}

func BytesCombine(pBytes ...[]byte) []byte {
    return bytes.Join(pBytes, []byte(""))
}
