package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/jsonq"
	"github.com/tkanos/gonfig"
	"io/ioutil"
	"net/http"
	"strings"
)

type Tracker struct {
	Name     string   `json:"name"`
	TrackURL string   `json:"track_url"`
	Field    string   `json:"field"`
	MinValue float64  `json:"min_value"`
	MaxValue float64  `json:"max_value"`
	Alerts   []string `json:"alerts"`
}

type Alert struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Param string `json:"param"`
}


type Configuration struct {
	Trackers []Tracker `json:"trackers"`
	Alerts   []Alert `json:"alerts"`
}

var conf Configuration

func main() {
	conf = Configuration{}
	err := gonfig.GetConf("config.json", &conf)
	if err != nil {
		panic(err)
	}
	for _, t := range conf.Trackers {
		fmt.Printf("Processing tracker %s\n", t.Name)
		val, err := GetValue(t.TrackURL, t.Field);
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
		}
		fmt.Printf("Result %.1f\n", val)
		if val < t.MinValue {
			SendAlert(conf, t, val, fmt.Sprintf("Value %.1f under minimum limit %.1f", val, t.MinValue))
		}
		if val > t.MaxValue {
			SendAlert(conf, t, val, fmt.Sprintf("Value %.1f above maximum limit %.1f", val, t.MaxValue))
		}
	}
}

func GetValue(url, field string) (float64, error) {
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("http.client.Get error: %s\n", err)
		return 0, err
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("%s\n", err)
		return 0, err
	}

	data := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(string(contents)))
	dec.Decode(&data)
	jq := jsonq.NewQuery(data)

	i, _ := jq.Float(field)
	return i, nil
}

func SendAlert(conf Configuration, t Tracker, val float64, txt string) {
	fmt.Printf("%s\n", txt)
	for _, s := range t.Alerts {
		fmt.Printf("Alerting %s\n", s)
		for _, a := range conf.Alerts {
			if a.Name == s {
				SendJson(a.URL, []byte(`{"` + a.Param +  `":"` + txt + `"}`))
			}
		}
	}
}

func SendJson(url string, jsonStr []byte) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}




