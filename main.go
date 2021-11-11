package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type ConfigYaml struct {
	Signsec    string `yaml:"signsec"`
	Oauthtoken string `yaml:"oauthtoken"`
}

type FormValue struct {
	Tid string
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "")
}

func InteractiveEndpoint(w http.ResponseWriter, r *http.Request) {
	var c ConfigYaml

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		panic(err)
	}

	tmpl, err := ioutil.ReadFile("form.template") // just pass the file name
	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		spew.Dump(err)
	} else {
		VERIFY_SLACKREQUEST := fmt.Sprintf(`v0:%s:%s`, r.Header.Get("X-Slack-Request-Timestamp"), b)
		SSS_TOKEN := []byte(c.Signsec)

		sig := hmac.New(sha256.New, SSS_TOKEN)
		sig.Write([]byte(VERIFY_SLACKREQUEST))

		if fmt.Sprintf("v0=%v", hex.EncodeToString(sig.Sum(nil))) == r.Header.Get("X-Slack-Signature") {
			decodedValue, err := url.QueryUnescape(fmt.Sprintf("%s", b)[8:])
			if err != nil {
				spew.Dump(err)
				fmt.Fprintf(w, "not ok")
			}
			fmt.Println(decodedValue)

			var result map[string]interface{}
			json.Unmarshal([]byte(decodedValue), &result)

			if result["type"] == "shortcut" {
				req, err := http.NewRequest("POST", "https://slack.com/api/views.open", bytes.NewBuffer([]byte(fmt.Sprintf(string(tmpl), result["trigger_id"]))))
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Oauthtoken))
				req.Header.Set("Content-Type", "application/json; charset=utf-8")

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					spew.Dump(err)
					fmt.Fprintf(w, "not ok")
				}
				defer resp.Body.Close()

				if resp.Status == "200 OK" {
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						spew.Dump(err)
						fmt.Fprintf(w, "not ok")
					} else {
						fmt.Printf("%s\n", body)
					}
					fmt.Fprintf(w, "")
				} else {
					fmt.Fprintf(w, "not ok")
				}

			} else if result["type"] == "view_submission" {
				jsonStr := []byte(fmt.Sprintf(`
		{
			"channel": "%s",
			"text": "message from %s: %s-%s-%s"
		}
				`,
					"C50EYHHNG",
					result["user"].(map[string]interface{})["username"],
					result["view"].(map[string]interface{})["state"].(map[string]interface{})["values"].(map[string]interface{})["block1"].(map[string]interface{})["a1"].(map[string]interface{})["value"],
					result["view"].(map[string]interface{})["state"].(map[string]interface{})["values"].(map[string]interface{})["block2"].(map[string]interface{})["a2"].(map[string]interface{})["value"],
					result["view"].(map[string]interface{})["state"].(map[string]interface{})["values"].(map[string]interface{})["block3"].(map[string]interface{})["a3"].(map[string]interface{})["value"],
				))

				req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(jsonStr))
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Oauthtoken))
				req.Header.Set("Content-Type", "application/json; charset=utf-8")

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Fprintf(w, "")
				}
				defer resp.Body.Close()

				if resp.Status == "200 OK" {
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						spew.Dump(err)
						fmt.Fprintf(w, "not ok")
					} else {
						fmt.Printf("%s\n", body)
					}
					fmt.Fprintf(w, "")
				} else {
					fmt.Fprintf(w, "not ok")
				}
			} else {
				spew.Dump(result)
				fmt.Fprintf(w, "type not recognized")
			}
		} else {
			fmt.Fprintf(w, "Only accepts request from slack")
		}
	}

}

func OptionsLoadEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok3")
}

func main() {
	http.HandleFunc("/", hello)
	http.HandleFunc("/interactive-endpoint", InteractiveEndpoint)
	http.HandleFunc("/options-load-endpoint", OptionsLoadEndpoint)

	fmt.Printf("Starting server for testing HTTP POST...\n")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}
