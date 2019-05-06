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
	case "station":
		doStation(client, args)
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
	if len(args) > 0 {
		panic("Too many arguments")
	}

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
	case "create":
		doIncidentCreate(client, args)
	case "get":
		doIncidentGet(client, args)
	default:
		panic("Bad action: " + action)
	}
}

func doIncidentCreate(client *emergencyreporting.Client, args []string) {
	if len(args) == 0 {
		panic("Missing argument: incident JSON")
	}
	jsonString := args[0]
	args = args[1:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	var incident emergencyreporting.Incident
	err := json.Unmarshal([]byte(jsonString), &incident)
	if err != nil {
		panic(fmt.Errorf("Could not parse JSON: %v", err))
	}
	postIncidentResponse, err := client.PostIncident(incident)
	if err != nil {
		panic(fmt.Errorf("Could not create incident: %v", err))
	}
	spew.Dump(postIncidentResponse)
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
		return
	}

	spew.Dump(currentIncident)

	if len(args) == 0 {
		return
	}

	action := args[0]
	args = args[1:]

	switch action {
	case "exposure":
		doIncidentExposure(client, currentIncident.IncidentID, args)
	default:
		panic("Bad action: " + action)
	}
}

func doIncidentExposure(client *emergencyreporting.Client, incidentID string, args []string) {
	if len(args) == 0 {
		panic("Missing argument: action")
	}

	action := args[0]
	args = args[1:]

	switch action {
	case "get":
		doIncidentExposureGet(client, incidentID, args)
	default:
		panic("Bad action: " + action)
	}
}

func doIncidentExposureGet(client *emergencyreporting.Client, incidentID string, args []string) {
	if len(args) == 0 {
		panic("Missing argument: filter (example: 'exposureID eq 1234'")
	}
	filter := args[0]
	args = args[1:]

	var currentExposure *emergencyreporting.Exposure
	{
		exposuresResponse, err := client.GetExposures(incidentID, map[string]string{"filter": filter}, true)
		if err != nil {
			panic(err)
		}
		if len(exposuresResponse.Exposures) > 0 {
			currentExposure = exposuresResponse.Exposures[0]
		}
	}
	if currentExposure == nil {
		fmt.Printf("Exposure not found.\n")
		return
	}

	spew.Dump(currentExposure)

	if len(args) == 0 {
		return
	}

	action := args[0]
	args = args[1:]

	switch action {
	case "delete":
		err := client.DeleteExposure(incidentID, currentExposure.ExposureID)
		if err != nil {
			panic(err)
		}
	case "members":
		doExposureMembers(client, currentExposure.ExposureID, args)
	default:
		panic("Bad action: " + action)
	}
}

func doExposureMembers(client *emergencyreporting.Client, exposureID string, args []string) {
	if len(args) == 0 {
		panic("Missing argument: action")
	}

	action := args[0]
	args = args[1:]

	switch action {
	case "get":
		doExposureMembersGet(client, exposureID, args)
	case "list":
		doExposureMembersList(client, exposureID, args)
	default:
		panic("Bad action: " + action)
	}
}

func doExposureMembersGet(client *emergencyreporting.Client, exposureID string, args []string) {
	if len(args) == 0 {
		panic("Missing argument: filter (example: 'exposureID eq 1234'")
	}
	filter := args[0]
	args = args[1:]

	var currentMember *emergencyreporting.CrewMember
	{
		membersResponse, err := client.GetExposureMembers(exposureID, map[string]string{"filter": filter})
		if err != nil {
			panic(err)
		}
		if len(membersResponse.CrewMembers) > 0 {
			currentMember = membersResponse.CrewMembers[0]
		}
	}
	if currentMember == nil {
		fmt.Printf("Member not found.\n")
		return
	}

	spew.Dump(currentMember)

	if len(args) == 0 {
		return
	}

	action := args[0]
	args = args[1:]

	switch action {
	case "roles":
		rolesResponse, err := client.GetExposureMemberRoles(exposureID, currentMember.ExposureUserID, nil)
		if err != nil {
			panic(err)
		}
		spew.Dump(rolesResponse.Roles)
	default:
		panic("Bad action: " + action)
	}
}

func doExposureMembersList(client *emergencyreporting.Client, exposureID string, args []string) {
	if len(args) > 0 {
		panic("Unexpected arguments")
	}

	membersResponse, err := client.GetExposureMembers(exposureID, nil)
	if err != nil {
		panic(err)
	}
	spew.Dump(membersResponse.CrewMembers)
}

func doStation(client *emergencyreporting.Client, args []string) {
	if len(args) == 0 {
		panic("Missing argument: action")
	}
	action := args[0]
	args = args[1:]

	switch action {
	case "get":
		doStationGet(client, args)
	case "list":
		doStationList(client, args)
	default:
		panic("Bad action: " + action)
	}
}

func doStationGet(client *emergencyreporting.Client, args []string) {
	if len(args) == 0 {
		panic("Missing argument: filter (example: 'stationNumber eq 2'")
	}
	filter := args[0]
	args = args[1:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	var currentStation *emergencyreporting.Station
	{
		apparatusesResponse, err := client.GetStations(map[string]string{"filter": filter})
		if err != nil {
			panic(err)
		}
		if len(apparatusesResponse.Stations) > 0 {
			currentStation = apparatusesResponse.Stations[0]
		}
	}
	if currentStation == nil {
		fmt.Printf("Station not found.\n")
		return
	} else {
		spew.Dump(currentStation)
	}
}

func doStationList(client *emergencyreporting.Client, args []string) {
	if len(args) != 0 {
		panic("Too many arguments")
	}

	options := map[string]string{
		"limit": "100",
	}
	stationsResponse, err := client.GetStations(options)
	if err != nil {
		panic(err)
	}
	spew.Dump(stationsResponse.Stations)
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
