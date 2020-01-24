# Emergency Reporting
Client for the Emergency Reporting API.

[![Go Report Card](https://goreportcard.com/badge/github.com/tekkamanendless/emergencyreporting)](https://goreportcard.com/report/github.com/tekkamanendless/emergencyreporting)
[![GoDoc](https://godoc.org/github.com/tekkamanendless/emergencyreporting?status.svg)](https://godoc.org/github.com/tekkamanendless/emergencyreporting)

TODO TODO TODO THIS IS SUPER NOT DONE YET

This is basically an Emergency Reporting client package that also comes with a command line tool.

## Using the CLI
Create a JSON file with your login and app information:

```
{
	"username": "YOUR USERNAME",
	"password": "YOUR PASSWORD",
	"client_id": "YOUR CLIENT ID/APP NAME",
	"client_secret": "YOUR CLIENT SECRET",
	"subscription_key": "YOUR SUBSCRIPTION KEY"
}
```

Then run:

```
emergencyreporting -config /path/to/config.json ...
```
