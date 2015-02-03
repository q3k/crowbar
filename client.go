package crowbar

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
    "encoding/base64"
	"crypto/hmac"
	"crypto/sha256"
)

type ProxyConnection struct {
	uuid	string
	server	string
}


func Connect(server, username, password, remote string) (ProxyConnection, error) {
	if strings.HasSuffix(server, "/") {
		server = server[:len(server)-1]
	}
	conn := ProxyConnection{server: server}

	args := fmt.Sprintf("?username=%s", username)
	resp, err := http.Get(conn.server + EndpointAuth + args)
	if err != nil {
		return ProxyConnection{}, err
	}
	data_bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ProxyConnection{}, err
	}
	defer resp.Body.Close()
	data := string(data_bytes)
	if !strings.HasPrefix(data, PrefixData) {
		msg := fmt.Sprintf("crowbar: Invalid data returned by server: %s", data)
		return ProxyConnection{}, errors.New(msg)
	}
	nonce_b64 := data[len(PrefixData):]
	decodeLen := base64.StdEncoding.DecodedLen(len(nonce_b64))
	nonce := make([]byte, decodeLen)
	n, err := base64.StdEncoding.Decode(nonce, []byte(nonce_b64))
	if err != nil {
		return ProxyConnection{}, errors.New("crowbar: Invalid nonce")
	}
	nonce = nonce[:n]

	mac := hmac.New(sha256.New, []byte(password))
	mac.Write(nonce)
	hmac := mac.Sum(nil)

	v := url.Values{}
	v.Set("remote_host", strings.Split(remote, ":")[0])
	v.Set("remote_port", strings.Split(remote, ":")[1])
	v.Set("username", username)
	v.Set("proof", base64.StdEncoding.EncodeToString(hmac))
	resp, err = http.Get(conn.server + EndpointConnect + "?" + v.Encode())
	if err != nil {
		return ProxyConnection{}, err
	}
	data_bytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return ProxyConnection{}, err
	}
	defer resp.Body.Close()
	data = string(data_bytes)
	fmt.Println(data)
	if !strings.HasPrefix(data, PrefixOK) {
		return ProxyConnection{}, errors.New("crowbar: Authentication error")
	}
	conn.uuid = data[len(PrefixOK):]

	return conn, nil
}
