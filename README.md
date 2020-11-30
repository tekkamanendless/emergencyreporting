# Emergency Reporting
Client for the Emergency Reporting API.

[![Go Report Card](https://goreportcard.com/badge/github.com/tekkamanendless/emergencyreporting)](https://goreportcard.com/report/github.com/tekkamanendless/emergencyreporting)
[![GoDoc](https://godoc.org/github.com/tekkamanendless/emergencyreporting?status.svg)](https://godoc.org/github.com/tekkamanendless/emergencyreporting)

This is basically an Emergency Reporting client package that also comes with a command line tool.

I have implemented the subsets of the API that I currently use, but there are many more endpoints that I haven't even looked at yet.

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

### Advanced Usage
The JSON configuration file supports the following additional fields:

* `tenant_host`; the TENANT_HOST value for authentication (default: `login.emergencyreporting.com`).
* `tenant_segment`; the TENANT_SEGMENT value for authentication (default: `login.emergencyreporting.com`).
* `host`; the host to use for API endpoints (default: `https://data.emergencyreporting.com`).
