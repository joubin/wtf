package travisci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/senorprogrammer/wtf/wtf"
)

const APIEnvToken = "WTF_TRAVIS_API_TOKEN"

func BuildsFor() (*Builds, error) {
	builds := &Builds{}

	pro := wtf.Config.UBool("wtf.mods.travisci.pro", false)
	travisAPIURL.Host = hosts[pro]

	resp, err := travisRequest("builds")
	if err != nil {
		return builds, err
	}

	parseJson(&builds, resp.Body)

	return builds, nil
}

/* -------------------- Unexported Functions -------------------- */

var (
	travisAPIURL = &url.URL{Scheme: "https", Path: "/"}
	hosts        = map[bool]string{
		false: "api.travis-ci.org",
		true:  "api.travis-ci.com",
	}
)

func travisRequest(path string) (*http.Response, error) {
	params := url.Values{}
	params.Add("limit", "10")

	url := travisAPIURL.ResolveReference(&url.URL{Path: path, RawQuery: params.Encode()})

	req, err := http.NewRequest("GET", url.String(), nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Travis-API-Version", "3")

	bearer := fmt.Sprintf("token %s", apiToken())
	req.Header.Add("Authorization", bearer)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf(resp.Status)
	}

	return resp, nil
}

func apiToken() string {
	return wtf.Config.UString(
		"wtf.mods.travisci.apiKey",
		os.Getenv(APIEnvToken),
	)
}

func parseJson(obj interface{}, text io.Reader) {
	jsonStream, err := ioutil.ReadAll(text)
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonStream))

	for {
		if err := decoder.Decode(obj); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}
}
