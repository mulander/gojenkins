package gojenkins

import (
	"encoding/json"
	"fmt"
	"io"
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

type Artifact struct {
	DisplayPath  string
	FileName     string
	RelativePath string
}

func (j Jenkins) Get(url string) (map[string]interface{}, error) {
	client := &http.Client{}
	target := j.Baseurl + url + "/api/json"

	//log.Println(target)

	r, err := http.NewRequest("GET", j.Baseurl+url+"/api/json", nil)
	r.SetBasicAuth(j.username, j.password)

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("gojenkins: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("gojenkins: Error parsing response from %s - %s", target, err.Error())
	}
	//log.Printf("%s", result)
	return result, nil
}

// List all jobs on the Baseurl Jenkins server
func (j Jenkins) Jobs() (map[string]Job, error) {
	resp, err := j.Get("")
	if err != nil {
		return nil, err
	}
	jobs := make(map[string]Job)
	for _, job := range resp["jobs"].([]interface{}) {
		entry := job.(map[string]interface{})
		j := Job{
			Name:  fmt.Sprint(entry["name"]),
			URL:   fmt.Sprint(entry["url"]),
			Color: fmt.Sprint(entry["color"]),
		}
		jobs[entry["name"].(string)] = j
	}
	return jobs, nil
}

// List artifacts from the provided build of the job
func (j Jenkins) Artifacts(job Job, build string) ([]Artifact, error) {
	resp, err := j.Get("/job/" + job.Name + "/" + build)
	if err != nil {
		return nil, err
	}
	artifacts := make([]Artifact, len(resp["artifacts"].([]interface{})))
	for idx, artifact := range resp["artifacts"].([]interface{}) {
		entry := artifact.(map[string]interface{})
		artifacts[idx] = Artifact{
			DisplayPath:  fmt.Sprint(entry["displayPath"]),
			FileName:     fmt.Sprint(entry["fileName"]),
			RelativePath: fmt.Sprint(entry["relativePath"]),
		}
	}
	return artifacts, err
}

// Download the artifact from the specified build of the provided job
// returns a Reader of the artifact
func (j Jenkins) Download(job Job, build string, a Artifact) (io.ReadCloser, error) {
	client := &http.Client{}

	url := j.Baseurl + "/job/" + job.Name + "/" + build + "/artifact/" + a.RelativePath
	//log.Println(url)

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	r.SetBasicAuth(j.username, j.password)

	resp, err := client.Do(r)
	if err != nil {
		log.Fatal(err)
	}

	switch resp.StatusCode {
	case 401:
		fallthrough
	case 404:
		return nil, fmt.Errorf("gojenkins: %s", resp.Status)
	}

	//log.Println(resp.StatusCode)

	return resp.Body, nil
}

// Sets the authentication for the Jenkins service
// Password can be an API token as described in:
// https://wiki.jenkins-ci.org/display/JENKINS/Authenticating+scripted+clients
func (j *Jenkins) SetAuth(username, password string) {
	j.username = username
	j.password = password
}
