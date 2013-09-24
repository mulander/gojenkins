package gojenkins

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type Jenkins struct {
	Baseurl  string
	username string
	password string
}

type Job struct {
	Name  string
	URL   string
	Color string
}

func (j Jenkins) Get(url string) (map[string]interface{}, error) {
	client := &http.Client{}

	r, err := http.NewRequest("GET", j.Baseurl+url+"/api/json", nil)
	r.SetBasicAuth(j.username, j.password)

	resp, err := client.Do(r)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s", result)
	return result, nil
}

// List all jobs on the Baseurl Jenkins server
func (j Jenkins) Jobs() (map[string]Job, error) {
	resp, err := j.Get("")
	jobs := make(map[string]Job)
	for _, job := range resp["jobs"].([]interface{}) {
		entry := job.(map[string]interface{})
		j := Job{
			Name:  entry["name"].(string),
			URL:   entry["url"].(string),
			Color: entry["color"].(string),
		}
		jobs[entry["name"].(string)] = j
	}
	return jobs, err
}

// Sets the authentication for the Jenkins service
// Password can be an API token as described in:
// https://wiki.jenkins-ci.org/display/JENKINS/Authenticating+scripted+clients
func (j *Jenkins) SetAuth(username, password string) {
	j.username = username
	j.password = password
}
