package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"github.com/tekkamanendless/emergencyreporting"
)

func main() {
	var rootCommand = &cobra.Command{
		Use:   "emergencyreporting",
		Short: "EmergencyReporting command-line utility",
		Long: `
This tool talks to the EmergencyReporting REST API.
`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set up any global stuff here that needs to run for every command.
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}
	rootCommand.PersistentFlags().String("config", "config.json", "Path to the configuration file with the client credentials.")
	rootCommand.PersistentFlags().Int("limit", 100, "The page size for any queries.")

	{
		command := &cobra.Command{
			Use:   "apparatus",
			Short: "Apparatus sub-command",
			Long:  ``,
			Run:   nil,
		}
		rootCommand.AddCommand(command)

		subCommand := &cobra.Command{
			Use:   "get <filter>",
			Short: "Get an apparatus",
			Long: `
Example filter: 'apparatusID eq 21-7'
			`,
			Args: cobra.ExactArgs(1),
			Run:  doApparatusGet,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "list [,filter>]",
			Short: "List all apparatuses",
			Long:  ``,
			Run:   doApparatusList,
		}
		command.AddCommand(subCommand)
	}
	{
		command := &cobra.Command{
			Use:   "incident",
			Short: "Incident sub-command",
			Long:  ``,
			Run:   nil,
		}
		rootCommand.AddCommand(command)

		subCommand := &cobra.Command{
			Use:   "create <json>",
			Short: "Create an incident",
			Long:  ``,
			Args:  cobra.ExactArgs(1),
			Run:   doIncidentCreate,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "get <filter> [exposure ...]",
			Short: "Get an incident",
			Long: `
Additional sub-commands for a particular incident:
* exposure get <filter>
* exposure get <filter> delete
* exposure get <filter> exposure <filter> delete
* exposure get <filter> exposure <filter> members get <filter>
* exposure get <filter> exposure <filter> members list
`,
			Args: cobra.MinimumNArgs(1),
			Run:  doIncidentGet,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "list [<filter>]",
			Short: "List all incidents",
			Long: `
Example filter: 'dispatchRunNumber eq 1234'
			`,
			Run: doIncidentList,
		}
		command.AddCommand(subCommand)
	}
	{
		command := &cobra.Command{
			Use:   "station",
			Short: "Station sub-command",
			Long:  ``,
			Run:   nil,
		}
		rootCommand.AddCommand(command)

		subCommand := &cobra.Command{
			Use:   "get <filter>",
			Short: "Get a station",
			Long:  ``,
			Args:  cobra.ExactArgs(1),
			Run:   doStationGet,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "list [<filter>]",
			Short: "List all stations",
			Long: `
Example filter: 'stationNumber eq 2'
			`,
			Run: doStationList,
		}
		command.AddCommand(subCommand)
	}
	{
		command := &cobra.Command{
			Use:   "user",
			Short: "User sub-command",
			Long:  ``,
			Run:   nil,
		}
		rootCommand.AddCommand(command)

		subCommand := &cobra.Command{
			Use:   "get <filter> [patch <operation> <path> <value>]",
			Short: "Get a user",
			Long:  ``,
			Args:  cobra.ExactArgs(1),
			Run:   doUserGet,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "id <filter> [patch <operation> <path> <value>]",
			Short: "Get a user by ID",
			Long:  ``,
			Args:  cobra.ExactArgs(1),
			Run:   doUserID,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "list [<filter>]",
			Short: "List all users",
			Long: `
Example filter: 'stationNumber eq 2'
			`,
			Run: doUserList,
		}
		command.AddCommand(subCommand)
	}

	err := rootCommand.Execute()
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}

func makeClient(cmd *cobra.Command) *emergencyreporting.Client {
	flag := cmd.Flag("config")
	configFile := flag.Value.String()

	if configFile == "" {
		fmt.Printf("Missing config file.\n")
		cmd.Help()
		os.Exit(1)
	}

	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Could not read '%s': %v\n", configFile, err)
		os.Exit(1)
	}

	var client *emergencyreporting.Client
	err = json.Unmarshal(configBytes, &client)
	if err != nil {
		fmt.Printf("Could not parse '%s': %v\n", configFile, err)
		os.Exit(1)
	}

	if len(client.Token) == 0 {
		tokenResponse, err := client.GenerateToken()
		if err != nil {
			fmt.Printf("Could not generate token: %v\n", err)
			os.Exit(1)
		}
		client.Token = tokenResponse.AccessToken
	}

	return client
}

func doApparatusGet(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

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

func doApparatusList(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	filter := ""
	if len(args) > 0 {
		filter = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		panic("Too many arguments")
	}

	options := map[string]string{
		"filter": filter,
		"limit":  cmd.Flag("limit").Value.String(),
	}
	apparatusesResponse, err := client.GetApparatuses(options)
	if err != nil {
		panic(err)
	}
	spew.Dump(apparatusesResponse.Apparatuses)
}

func doIncidentCreate(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

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

func doIncidentGet(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

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
	case "delete":
		err := client.DeleteIncident(currentIncident.IncidentID)
		if err != nil {
			panic(err)
		}
	case "exposure":
		doIncidentExposure(client, currentIncident.IncidentID, args)
	default:
		panic("Bad action: " + action)
	}
}

func doIncidentList(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	filter := ""
	if len(args) > 0 {
		filter = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		panic("Too many arguments")
	}

	options := map[string]string{
		"filter": filter,
		"limit":  cmd.Flag("limit").Value.String(),
	}

	incidentsResponse, err := client.GetIncidents(options, true)
	if err != nil {
		panic(err)
	}
	spew.Dump(incidentsResponse.Incidents)
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

func doStationGet(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

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

func doStationList(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	filter := ""
	if len(args) > 0 {
		filter = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		panic("Too many arguments")
	}

	options := map[string]string{
		"filter": filter,
		"limit":  cmd.Flag("limit").Value.String(),
	}
	stationsResponse, err := client.GetStations(options)
	if err != nil {
		panic(err)
	}
	spew.Dump(stationsResponse.Stations)
}

func doUserGet(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

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
	}
	spew.Dump(currentUser)

	doUserActions(client, currentUser, args)
}

func doUserID(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	userID := args[0]
	args = args[1:]

	var currentUser *emergencyreporting.User
	{
		userResponse, err := client.GetUser(userID, true)
		if err != nil {
			panic(err)
		}
		currentUser = userResponse.User
	}
	if currentUser == nil {
		fmt.Printf("User not found.\n")
		return
	}
	spew.Dump(currentUser)

	doUserActions(client, currentUser, args)
}

func doUserActions(client *emergencyreporting.Client, currentUser *emergencyreporting.User, args []string) {
	if len(args) == 0 {
		return
	}

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

func doUserList(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	filter := ""
	if len(args) > 0 {
		filter = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		panic("Too many arguments")
	}

	options := map[string]string{
		"filter": filter,
		"limit":  cmd.Flag("limit").Value.String(),
	}
	usersResponse, err := client.GetUsers(options, true)
	if err != nil {
		panic(err)
	}
	spew.Dump(usersResponse.Users)
}
