package talentmobtranscoding

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	AdminToken = os.Getenv("ADMIN_TOKEN")
)

type transcodeRequest struct {
	Task  string `json:"task"`
	Extra string `json:"extra"`
}

type BaseResponse struct {
	Success bool        `json:"success"`
	Info    string      `json:"info"`
	Result  interface{} `json:"result"`
}

const (
	productionBaseURL = "https://talentmob.herokuapp.com"
	endPoint          = "/api/1/admin/system"
)

func getURL() string {
	return fmt.Sprintf("%s%s", productionBaseURL, endPoint)
}

func Transcode(videoID uint64) error {

	task := transcodeRequest{
		Task:  "transcode_video",
		Extra: fmt.Sprintf("%d", videoID),
	}

	return sendRequest(task)
}

func TranscodeWithWatermark(videoID uint64) error {
	task := transcodeRequest{
		Task:  "transcode_with_watermark_video",
		Extra: fmt.Sprintf("%d", videoID),
	}

	return sendRequest(task)
}

func sendRequest(task transcodeRequest) error {

	req, err := http.NewRequest(http.MethodPost, getURL(), NewReader(task))

	if AdminToken == "" {
		panic("missing ADMIN_TOKEN")
	}

	req.Header.Add("Authorization", AdminToken)

	if err != nil {
		return err
	}

	res, err := Client.Do(req)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {

		b, err := ioutil.ReadAll(res.Body)

		if err != nil {
			return err
		}

		return errors.New(fmt.Sprintf("request was not successful error: %s statusCode: %d", string(b), res.StatusCode))
	}

	var response BaseResponse

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return err
	}

	if !response.Success {
		return errors.New(response.Info)
	}

	return nil
}

func NewReader(data interface{}) io.Reader {
	var buf io.ReadWriter
	buf = new(bytes.Buffer)
	json.NewEncoder(buf).Encode(data)
	return buf
}

var Client = &http.Client{
	Timeout: time.Second * 10,
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	},
}
