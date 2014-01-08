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
	_, err := jenkins.Get("")
	if err != nil {
		t.Errorf("Expected proper authentication but instead got:\n%s", err.Error())
	}
}

func TestBadAuth(t *testing.T) {
	Init()
	jenkins.SetAuth("bad", "auth")
	_, err := jenkins.Get("")
	if err == nil {
		t.Error("Incorrect authorization should return an error")
	}
}

func TestParseError(t *testing.T) {
	jenkins = &Jenkins{
		Baseurl: "http://example.com",
	}
	_, err := jenkins.Get("")
	if err == nil {
		t.Errorf("Expected a parsing error because the target is not a jenkins instance")
	}
}

func TestBadURL(t *testing.T) {
	jenkins = &Jenkins{
		Baseurl: "htt://example.com",
	}
	_, err := jenkins.Get("")
	if err == nil {
		t.Errorf("Expected a net/http error for incorrect URL")
	}
}

func TestJobsParseError(t *testing.T) {
	jenkins = &Jenkins{
		Baseurl: "http://example.com",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Error("jenkins.Jobs() on an incorrect url should not panic")
		}
	}()

	_, err := jenkins.Jobs()
	if err == nil {
		t.Error("Expected to receive an error from calling jenkins.Jobs() on an incorrect url")
	}
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

	jobs, err := jenkins.Jobs()
	if err != nil {
		t.Error(err)
	}

	artifacts, _ := jenkins.Artifacts(jobs[testJob], "lastSuccessfulBuild")
	for _, artifact := range artifacts {
		t.Log(artifact.FileName)
	}
}

func TestArtifactsParseError(t *testing.T) {
	jenkins = &Jenkins{
		Baseurl: "http://example.com",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Error("jenkins.Artifacts() on an incorrect url/data should not panic")
		}
	}()

	_, err := jenkins.Artifacts(Job{"bad-job", "bad-job", "bad-color"}, "bad-build")
	if err == nil {
		t.Error("Expected to receive an error from calling jenkins.Artifacts() on an incorrect url/data")
	}
	t.Log(err)
}

func TestArtifactsBadJobBuild(t *testing.T) {
	Init()
	defer func() {
		if r := recover(); r != nil {
			t.Error("jenkins.Artifacts() on an incorrect job should not panic")
		}
	}()

	_, err := jenkins.Artifacts(Job{"bad-job", "bad-job", "bad-color"}, "bad-build")
	if err == nil {
		t.Error("Expected to receive an error from calling jenkins.Artifacts() on an incorrect url/data")
	}
	t.Log(err)
}

func TestArtifactsBadBuild(t *testing.T) {
	Init()
	defer func() {
		if r := recover(); r != nil {
			t.Error("jenkins.Artifacts() on an incorrect job should not panic")
		}
	}()

	jobs, err := jenkins.Jobs()
	if err != nil {
		t.Errorf("Did not expect an error from listing jobs but got: ", err.Error())
	}
	for _, job := range jobs {
		_, err = jenkins.Artifacts(job, "bad-build")
		if err == nil {
			t.Error("Expected to receive an error from calling jenkins.Artifacts() on an incorrect url/data")
		}
		t.Log(err)
	}
}

func TestDownloadNoAuth(t *testing.T) {
	Init()
	jenkins.SetAuth("bad", "auth")
	_, err := jenkins.Download(Job{"bad-job", "bad-job", "bad-color"}, "lastSuccessfulBuild", Artifact{"bad-path", "bad-file", "bad-relative-path"})
	if err == nil {
		t.Error("test jenkins.Download: Incorrect authorization should return an error")
	}
}

func TestDownloadBadBuild(t *testing.T) {
	Init()

	jobs, err := jenkins.Jobs()
	if err != nil {
		t.Error(err)
	}

	_, err = jenkins.Download(jobs[testJob], "bad-build", Artifact{"bad-path", "bad-file", "bad-relative-path"})
	if err == nil {
		t.Error("Asking for a download from a bad build should be an error")
	}
}

func TestDownloadBadFile(t *testing.T) {
	Init()

	jobs, err := jenkins.Jobs()
	if err != nil {
		t.Error(err)
	}

	_, err = jenkins.Download(jobs[testJob], "lastSuccessfulBuild", Artifact{"bad-path", "bad-file", "bad-relative-path"})
	if err == nil {
		t.Error("Asking for a download for a bad file should be an error")
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

	artifacts, _ := jenkins.Artifacts(jobs[testJob], "lastSuccessfulBuild")
	for _, artifact := range artifacts {
		t.Logf("Downloading file: %s\n", artifact.FileName)
		out, err := os.Create(artifact.FileName)
		if err != nil {
			t.Error(err)
		}
		defer out.Close()

		a, err := jenkins.Download(jobs[testJob], "lastSuccessfulBuild", artifact)
		if err != nil {
			t.Error("Downloading a proper artifact should not result in an error")
		}
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
