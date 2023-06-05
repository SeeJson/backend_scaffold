package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/k0kubun/pp"
	"github.com/sirupsen/logrus"

	"github.com/SeeJson/backend_scaffold/docgen"
)

const apiUrl = "/server/index.php?s=/api/item/updateByApi"

func main() {
	var (
		inputFile string
		appKey    string
		appToken  string
		host      string
	)
	//flag.StringVar(&inputFile, "input", "../showdoc_data.json", "")
	flag.StringVar(&inputFile, "input", "../../cmd/backend/showdoc_data.json", "")
	flag.StringVar(&appKey, "app_id", "fc9edd3e9ea7e84f0d6f1f633628ff22381213627", "")
	flag.StringVar(&appToken, "app_token", "afe23adb764146d10eaacd4de642d93a2100043314", "")
	flag.StringVar(&host, "host", "http://159.75.48.131", "")
	flag.Parse()

	bb, err := ioutil.ReadFile(inputFile)
	Panic(err)

	var sdm []docgen.ShowDocModel
	Panic(json.Unmarshal(bb, &sdm))

	pp.Println(sdm)

	postUrl := host + apiUrl
	pp.Println(postUrl)
	for _, s := range sdm {
		buf := bytes.Buffer{}

		req := Request{
			ApiKey:       appKey,
			ApiToken:     appToken,
			ShowDocModel: s,
		}
		Panic(json.NewEncoder(&buf).Encode(req))
		fmt.Println(buf.String())

		rsp, err := postData(postUrl, &buf)
		if err != nil {
			panic(err)
		}
		pp.Println(rsp)
	}
}

type Request struct {
	ApiKey   string `json:"api_key"`
	ApiToken string `json:"api_token"`

	docgen.ShowDocModel `json:",inline"`
}

type Response struct {
	ErrorCode    int                    `json:"error_code"`
	ErrorMessage string                 `json:"error_message"`
	Data         map[string]interface{} `json:"data"`
}

func postData(url string, reqData io.Reader) (*Response, error) {
	rsp, err := http.Post(url, "application/json", reqData)
	if err != nil {
		logrus.Errorf("fail to POST, url: %v, err: %v", url, err)
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		logrus.Errorf("response error, url: %v, code: %v", url, rsp.StatusCode)
	}
	buf, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	var resp Response
	if err := json.Unmarshal(buf, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func Panic(err error) {
	if err != nil {
		panic(err)
	}
}
