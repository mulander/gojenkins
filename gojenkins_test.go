package gojenkins

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var jenkins *Jenkins
var testJob string

type Config struct {
	Baseurl  string
	Username string
	Password string
	TestJob  string
}

func Init() {
	jsonConfig, err := ioutil.ReadFile("test.json")
	if err != nil {
		log.Fatalf("Error reading test config. Provide one based on test.json.example: %s", err)
	}

	var config Config
	err = json.Unmarshal(jsonConfig, &config)
	if err != nil {
		log.Fatal(err)
	}

	jenkins = &Jenkins{
		Baseurl: config.Baseurl,
	}
	jenkins.SetAuth(config.Username, config.Password)

	testJob = config.TestJob
}

func TestAuth(t *testing.T) {
	Init()
	jenkins.Get("")
}

func TestJobs(t *testing.T) {
	Init()
	jobs, _ := jenkins.Jobs()
	for _, job := range jobs {
		t.Log(job.URL)
	}
}

func TestArtifacts(t *testing.T) {
	Init()
	artifacts, _ := jenkins.Artifacts(testJob, "lastSuccessfulBuild")
	for _, artifact := range artifacts {
		t.Log(artifact.FileName)
	}
}

func TestDownloadArtifactsLatest(t *testing.T) {
	Init()

	jobs, err := jenkins.Jobs()
	if err != nil {
		t.Error(err)
	}

	err = os.Mkdir("downloads", 0700)
	if err != nil {
		t.Error(err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	err = os.Chdir("downloads")
	if err != nil {
		t.Error(err)
	}

	artifacts, _ := jenkins.Artifacts(testJob, "lastSuccessfulBuild")
	for _, artifact := range artifacts {
		t.Logf("Downloading file: %s\n", artifact.FileName)
		out, err := os.Create(artifact.FileName)
		if err != nil {
			t.Error(err)
		}
		defer out.Close()

		a := jenkins.Download(jobs[testJob], "lastSuccessfulBuild", artifact)
		defer a.Close()

		_, err = io.Copy(out, a)
		if err != nil {
			t.Error(err)
		}
	}

	err = os.Chdir(pwd)
	if err != nil {
		t.Error(err)
	}

	err = os.RemoveAll("downloads")
	if err != nil {
		t.Error(err)
	}
}
