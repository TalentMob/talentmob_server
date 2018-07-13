package talenthttp

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var Client = &http.Client{
	Timeout: time.Second * 10,
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	},
}

func Request(method string, url string, data io.Reader, v interface{}) error {
	req, err := http.NewRequest(method, url, data)

	if err != nil {
		return err
	}

	response, err := Client.Do(req)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {

		b, err := ioutil.ReadAll(response.Body)

		if err != nil {
			return err
		}

		return errors.New(string(b))

	}

	return json.NewDecoder(response.Body).Decode(v)
}
