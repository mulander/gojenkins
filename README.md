gojenkins
=========

Go library for working with the Jenkins Remote access API (https://wiki.jenkins-ci.org/display/JENKINS/Remote+access+API)

installation
------------

`go get github.com/mulander/gojenkins` or just `import "github.com/mulander/gojenkins"` in your code.

status
------

In active development and maintained. Bugs/feature request reported on the github issue tracker will be actively fixed.

features
--------

* Authentication with the Jenkins json api
* List all jobs on the Baseurl Jenkins server
* List artifacts from the provided build of the job
* Download the artifact from the specified build of the provided job (returns a Reader)

contributing
------------

Pull requests and tickets welcome.

Good places for initial tasks:
* https://wiki.jenkins-ci.org/display/JENKINS/Remote+access+API take a look at available functionality
* Install a jenkins server and append /api/json to urls, look at the output and implement start hacking on an API
* Write more unit tests :)

usage
-----

The following example downloads all artifacts from a job named 'gojenkins'
and saves it locally to disk.

```go
jenkins = &Jenkins{
  Baseurl: "http://some.jenkins.build.server.example.com",
}

jenkins.SetAuth("user", "secret-api-token")

// List all Jenkins jobs
jobs, err := jenkins.Jobs()
if err != nil {
	log.Fatal(err)
}

// List all artifacts of the gojenkins job
artifacts, err := jenkins.Artifacts(jobs["gojenkins"], "lastSuccessfulBuild")
if err != nil {
	log.Fatal(err)
}

// For every job artifact
for _, artifact := range artifacts {
	out, err := os.Create(artifat.FileName)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// perform a download receiving a reader
	a, err := jenkins.Download(jobs[testJob], "lastSuccessfulBuild", artifact)
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	// store it to disk using io/ioutil
	_, err := io.Copy(out, a)
	if err != nil {
		log.Fatal(err)
	}
}
```
