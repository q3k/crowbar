// Copyright (c) 2015, Segiusz 'q3k' Bazanski <sergiusz@bazanski.pl>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package crowbar

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"encoding/base64"
	"crypto/hmac"
	"crypto/sha256"
)

type ProxyConnection struct {
	uuid		string
	server		string
	read_buffer	[]byte
	read_mutex	sync.Mutex
}

func (c *ProxyConnection) Write(b []byte) (int, error) {
	url_args := fmt.Sprintf("?uuid=" + c.uuid)
	post_args := url.Values{}
	post_args.Set("data", base64.StdEncoding.EncodeToString(b))

	resp, err := http.PostForm(c.server + EndpointSync + url_args, post_args)
	if err != nil {
		return 0, err
	}
	data_bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	data := string(data_bytes)

	if !strings.HasPrefix(data, PrefixOK) {
		msg := fmt.Sprintf("Could not send to server: %s", data)
		return 0, errors.New(msg)
	}

	return len(b), nil
}

func (c *ProxyConnection) FillReadBuffer() error {
	args := fmt.Sprintf("?uuid=" + c.uuid)
	resp, err := http.Get(c.server + EndpointSync + args)
	if err != nil {
		return err
	}
	data_bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data := string(data_bytes)

	if strings.HasPrefix(data, PrefixData) {
		data := data[len(PrefixData):]

		decodeLen := base64.StdEncoding.DecodedLen(len(data))
		bData := make([]byte, len(c.read_buffer) + decodeLen)
		n, err := base64.StdEncoding.Decode(bData[len(c.read_buffer):], []byte(data))
		if err != nil {
			return err
		}
		bData = bData[:len(c.read_buffer)+n]
		c.read_buffer = bData
	} else {
		return errors.New("Could not read from server")
	}
	return nil
}

func (c *ProxyConnection) Read(b []byte) (n int, err error) {
	c.read_mutex.Lock()
	// If local buffer is empty, get new data
	if len(c.read_buffer) == 0 {
		err := c.FillReadBuffer()
		if err != nil {
			c.read_mutex.Unlock()
			return 0, err
		}
	}
	// Return local buffer
	count := len(b)
	if count > len(c.read_buffer){
		count = len(c.read_buffer)
	}
	copy(b, c.read_buffer[:count])
	c.read_buffer = c.read_buffer[count:]

	c.read_mutex.Unlock()
	return count, nil
}

func Connect(server, username, password, remote string) (*ProxyConnection, error) {
	if strings.HasSuffix(server, "/") {
		server = server[:len(server)-1]
	}
	conn := ProxyConnection{server: server}

	args := fmt.Sprintf("?username=%s", username)
	resp, err := http.Get(conn.server + EndpointAuth + args)
	if err != nil {
		return &ProxyConnection{}, err
	}
	data_bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &ProxyConnection{}, err
	}
	defer resp.Body.Close()
	data := string(data_bytes)
	if !strings.HasPrefix(data, PrefixData) {
		msg := fmt.Sprintf("crowbar: Invalid data returned by server: %s", data)
		return &ProxyConnection{}, errors.New(msg)
	}
	nonce_b64 := data[len(PrefixData):]
	decodeLen := base64.StdEncoding.DecodedLen(len(nonce_b64))
	nonce := make([]byte, decodeLen)
	n, err := base64.StdEncoding.Decode(nonce, []byte(nonce_b64))
	if err != nil {
		return &ProxyConnection{}, errors.New("crowbar: Invalid nonce")
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
		return &ProxyConnection{}, err
	}
	data_bytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return &ProxyConnection{}, err
	}
	defer resp.Body.Close()
	data = string(data_bytes)
	if !strings.HasPrefix(data, PrefixOK) {
		return &ProxyConnection{}, errors.New("crowbar: Authentication error")
	}
	conn.uuid = data[len(PrefixOK):]

	return &conn, nil
}
