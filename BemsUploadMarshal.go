package main

import (
    "encoding/xml"
)

func BemsUploadMarshal(xmlstruct interface{}) ([]byte, error) {
    text, err := xml.MarshalIndent(xmlstruct, "", "    ")
    if err != nil {
        return []byte(""), err
    }
    text = BytesCombine([]byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n"),text)
    return text, nil
}
