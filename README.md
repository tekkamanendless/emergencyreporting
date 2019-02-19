# Emergency Reporting
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
