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
	"account_id": "ER ACCOUNT ID",
	"user_id": "ER USER ID",
	"client_id": "YOUR CLIENT ID/APP NAME",
	"client_secret": "YOUR CLIENT SECRET",
	"subscription_key": "YOUR SUBSCRIPTION KEY"
}
```

Then run:

```
emergencyreporting -config /path/to/config.json ...
```

### Examples
Get a token:

```
emergencyreporting -config /path/to/config.json login
```

Raw operation to get the current user:

```
emergencyreporting -config /path/to/config.json raw get https://data.emergencyreporting.com/agencyusers/v2/users/me
```
