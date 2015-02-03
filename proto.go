package crowbar

import (
    "fmt"
    "net/http"
    "encoding/base64"
)

const PrefixError string = "ERROR:"
const PrefixOK string = "OK:"
const PrefixData string = "DATA:"
const PrefixQuit string = "QUIT:"

func WriteHTTPError(w http.ResponseWriter, message string) {
    body := fmt.Sprintf("%s%s", PrefixError, message)
    http.Error(w, body, http.StatusInternalServerError)
}

func WriteHTTPOK(w http.ResponseWriter, data string) {
    fmt.Fprintf(w, "%s:%s", PrefixOK, data)
}

func WriteHTTPData(w http.ResponseWriter, data []byte) {
    data_encoded := base64.StdEncoding.EncodeToString(data)
    fmt.Fprintf(w, "%s:%s", PrefixData, data_encoded)
}

func WriteHTTPQuit(w http.ResponseWriter, data string) {
    fmt.Fprintf(w, "%s:%s", PrefixQuit, data)
}

const EndpointConnect string = "/connect"
const EndpointSync string = "/sync"
