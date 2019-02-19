package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/tekkamanendless/emergencyreporting"
)

func main() {
	configFile := flag.String("config", "config.json", "Path to the configuration file with the client credentials.")

	flag.Parse()

	if configFile == nil {
		fmt.Printf("Missing config file.\n")
		flag.Usage()
		os.Exit(1)
	}

	configBytes, err := ioutil.ReadFile(*configFile)
	if err != nil {
		fmt.Printf("Could not read '%s': %v\n", *configFile, err)
		os.Exit(1)
	}

	var client *emergencyreporting.Client
	err = json.Unmarshal(configBytes, &client)
	if err != nil {
		fmt.Printf("Could not parse '%s': %v\n", *configFile, err)
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) < 1 {
		panic("Missing argument: command")
	}
	command := args[0]
	args = args[1:]

	if len(client.Token) == 0 {
		tokenResponse, err := client.GenerateToken()
		if err != nil {
			fmt.Printf("Could not generate token: %v\n", err)
			os.Exit(1)
		}
		client.Token = tokenResponse.AccessToken
	}

	switch command {
	case "apparatus":
		doApparatus(client, args)
	case "incident":
		doIncident(client, args)
	case "user":
		doUser(client, args)
	default:
		panic("Unknown command: " + command)
	}
}

func doApparatus(client *emergencyreporting.Client, args []string) {
	if len(args) == 0 {
		panic("Missing argument: action")
	}
	action := args[0]
	args = args[1:]

	switch action {
	case "get":
		doApparatusGet(client, args)
	case "list":
		doApparatusList(client, args)
	default:
		panic("Bad action: " + action)
	}
}

func doApparatusGet(client *emergencyreporting.Client, args []string) {
	if len(args) == 0 {
		panic("Missing argument: filter (example: 'apparatusID eq 21-7'")
	}
	filter := args[0]
	args = args[1:]

	var currentApparatus *emergencyreporting.Apparatus
	{
		apparatusesResponse, err := client.GetApparatuses(map[string]string{"filter": filter})
		if err != nil {
			panic(err)
		}
		if len(apparatusesResponse.Apparatuses) > 0 {
			currentApparatus = apparatusesResponse.Apparatuses[0]
		}
	}
	if currentApparatus == nil {
		fmt.Printf("Apparatus not found.\n")
		return
	} else {
		spew.Dump(currentApparatus)
	}
}

func doApparatusList(client *emergencyreporting.Client, args []string) {
	if len(args) != 0 {
		panic("Too many arguments")
	}

	options := map[string]string{
		"limit": "100",
	}
	apparatusesResponse, err := client.GetApparatuses(options)
	if err != nil {
		panic(err)
	}
	spew.Dump(apparatusesResponse.Apparatuses)
}

func doIncident(client *emergencyreporting.Client, args []string) {
	if len(args) == 0 {
		panic("Missing argument: action")
	}
	action := args[0]
	args = args[1:]

	switch action {
	case "get":
		doIncidentGet(client, args)
	default:
		panic("Bad action: " + action)
	}
}

func doIncidentGet(client *emergencyreporting.Client, args []string) {
	if len(args) == 0 {
		panic("Missing argument: filter (example: 'dispatchRunNumber eq 1234'")
	}
	filter := args[0]
	args = args[1:]

	var currentIncident *emergencyreporting.Incident
	{
		incidentsResponse, err := client.GetIncidents(map[string]string{"filter": filter}, true)
		if err != nil {
			panic(err)
		}
		if len(incidentsResponse.Incidents) > 0 {
			currentIncident = incidentsResponse.Incidents[0]
		}
	}
	if currentIncident == nil {
		fmt.Printf("Incident not found.\n")
	} else {
		spew.Dump(currentIncident)
	}
}

func doUser(client *emergencyreporting.Client, args []string) {
	if len(args) == 0 {
		panic("Missing argument: action")
	}
	action := args[0]
	args = args[1:]

	switch action {
	case "get":
		doUserGet(client, args)
	case "list":
		doUserList(client, args)
	default:
		panic("Bad action: " + action)
	}
}

func doUserGet(client *emergencyreporting.Client, args []string) {
	if len(args) == 0 {
		panic("Missing argument: filter (example: 'primaryEmail eq bob@example.com'")
	}
	filter := args[0]
	args = args[1:]

	var currentUser *emergencyreporting.User
	{
		usersResponse, err := client.GetUsers(map[string]string{"filter": filter}, true)
		if err != nil {
			panic(err)
		}
		if len(usersResponse.Users) > 0 {
			currentUser = usersResponse.Users[0]
		}
	}
	if currentUser == nil {
		fmt.Printf("User not found.\n")
		return
	} else {
		spew.Dump(currentUser)
	}

	// ---

	if len(args) > 0 {
		action := args[0]
		args = args[1:]

		switch action {
		case "patch":
			if len(args) != 3 {
				panic("Patch needs three arguments")
			}
			operation := args[0]
			path := args[1]
			value := args[2]

			patchUserRequest := emergencyreporting.PatchUserRequest{
				{
					Operation: operation,
					Path:      path,
					Value:     value,
				},
			}
			patchResponse, err := client.PatchUser(currentUser.UserID, currentUser.RowVersion, patchUserRequest)
			if err != nil {
				panic(err)
			}
			spew.Dump(patchResponse)
		default:
			panic("Unsupported action: " + action)
		}
	}
}

func doUserList(client *emergencyreporting.Client, args []string) {
	if len(args) != 0 {
		panic("Too many arguments")
	}

	usersResponse, err := client.GetUsers(nil, true)
	if err != nil {
		panic(err)
	}
	spew.Dump(usersResponse.Users)
}

func listStations(client *emergencyreporting.Client) map[string]*emergencyreporting.Station {
	stationsResponse, err := client.GetStations(nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Stations: %d\n", len(stationsResponse.Stations))
	for stationIndex, station := range stationsResponse.Stations {
		fmt.Printf("Station %d / %d:\n", stationIndex+1, len(stationsResponse.Stations))
		spew.Dump(station)
	}

	result := map[string]*emergencyreporting.Station{}
	for _, station := range stationsResponse.Stations {
		result[station.StationNumber] = station
	}
	return result
}

func listApparatuses(client *emergencyreporting.Client) map[string]*emergencyreporting.Apparatus {
	options := map[string]string{
		"limit": "100",
	}
	apparatusesResponse, err := client.GetApparatuses(options)
	if err != nil {
		panic(err)
	}

	result := map[string]*emergencyreporting.Apparatus{}
	for _, apparatus := range apparatusesResponse.Apparatuses {
		result[apparatus.ApparatusID] = apparatus
	}
	return result
}

func listSomeIncidents(client *emergencyreporting.Client) {
	incidents, err := client.GetIncidents(map[string]string{"limit": "4"}, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Incidents: %d\n", len(incidents.Incidents))
	for incidentIndex, incident := range incidents.Incidents {
		fmt.Printf("Incident %d / %d:\n", incidentIndex+1, len(incidents.Incidents))
		spew.Dump(incident)
	}
}

func makeAnIncident(client *emergencyreporting.Client, desiredIncident emergencyreporting.Incident) {
	var currentIncident *emergencyreporting.Incident
	{
		incidentsResponse, err := client.GetIncidents(map[string]string{"filter": "dispatchRunNumber eq " + desiredIncident.DispatchRunNumber}, true)
		if err != nil {
			panic(err)
		}
		if len(incidentsResponse.Incidents) > 0 {
			currentIncident = incidentsResponse.Incidents[0]
		}
	}
	if currentIncident != nil {
		fmt.Printf("We already have incident %s in the system.\n", desiredIncident.DispatchRunNumber)
	} else {
		fmt.Printf("We do not have incident %s in the system.\n", desiredIncident.DispatchRunNumber)
		newIncident, err := client.PostIncident(desiredIncident)
		if err != nil {
			panic(err)
		}
		spew.Dump(newIncident)

		incidentsResponse, err := client.GetIncidents(map[string]string{"filter": "dispatchRunNumber eq " + desiredIncident.DispatchRunNumber}, false)
		if err != nil {
			panic(err)
		}
		if len(incidentsResponse.Incidents) > 0 {
			currentIncident = incidentsResponse.Incidents[0]
		}
	}
	spew.Dump(currentIncident)

	for desiredExposureIndex, desiredExposure := range desiredIncident.Exposures {
		var currentExposure *emergencyreporting.Exposure
		newExposure := false
		if len(currentIncident.Exposures) >= desiredExposureIndex+1 {
			fmt.Printf("We already have exposure #%d.\n", desiredExposureIndex+1)
			currentExposure = currentIncident.Exposures[desiredExposureIndex]
		} else {
			fmt.Printf("We need to make exposure #%d.\n", desiredExposureIndex+1)
			postExposureResponse, err := client.PostExposure(currentIncident.IncidentID, *desiredExposure)
			if err != nil {
				panic(err)
			}
			getExposureResponse, err := client.GetExposure(currentIncident.IncidentID, postExposureResponse.ExposureID, true)
			if err != nil {
				panic(err)
			}
			currentExposure = getExposureResponse.Exposure

			newExposure = true
			desiredExposure.Location.RowVersion = currentExposure.Location.RowVersion
		}
		spew.Dump(currentExposure)

		if newExposure {
			_, err := client.PutExposureLocation(currentExposure.ExposureID, *desiredExposure.Location)
			if err != nil {
				panic(err)
			}
		}

		for _, desiredApparatus := range desiredExposure.Apparatuses {
			_, err := client.PostExposureApparatus(currentExposure.ExposureID, *desiredApparatus)
			if err != nil {
				panic(err)
			}
		}
	}
}
