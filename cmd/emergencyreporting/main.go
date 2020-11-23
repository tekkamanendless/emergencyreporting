package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

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
			Use:   "login",
			Short: "Log in and get a token",
			Long:  ``,
			Run:   doLogin,
		}
		rootCommand.AddCommand(command)
	}

	{
		command := &cobra.Command{
			Use:   "raw",
			Short: "raw sub-command",
			Long:  `Perform a raw API operation.`,
			Run:   nil,
		}
		rootCommand.AddCommand(command)

		for _, method := range []string{http.MethodDelete, http.MethodGet, http.MethodPost, http.MethodPatch} {
			func(method string) {
				var parameters []string
				var headers []string
				var contents string
				subCommand := &cobra.Command{
					Use:   fmt.Sprintf("%s <url>", strings.ToLower(method)),
					Short: fmt.Sprintf("Perform a %s operation", method),
					Long:  ``,
					Args:  cobra.ExactArgs(1),
					Run: func(cmd *cobra.Command, args []string) {
						doRaw(cmd, args, method, parameters, headers, contents)
					},
				}
				subCommand.Flags().StringArrayVar(&parameters, "parameter", nil, `URL parameter, such as "x=Hello There"`)
				subCommand.Flags().StringArrayVar(&headers, "header", nil, `Header parameter, such as "X-Custom-Key: Hello There"`)
				subCommand.Flags().StringVar(&contents, "contents", "", `Contents to send`)
				command.AddCommand(subCommand)
			}(method)
		}
	}

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
			Use:   "exposure",
			Short: "Exposure sub-command",
			Long:  ``,
			Run:   nil,
		}
		rootCommand.AddCommand(command)

		subCommand := &cobra.Command{
			Use:   "get <filter>",
			Short: "Get an exposure",
			Long:  ``,
			Args:  cobra.ExactArgs(1),
			Run:   doExposureGet,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "list [<filter>]",
			Short: "List all exposures",
			Long: `
Example filter: 'incidentID eq 1234'
			`,
			Run: doExposureList,
		}
		command.AddCommand(subCommand)
	}
	{
		command := &cobra.Command{
			Use:   "incident-exposure",
			Short: "Incident exposure sub-command",
			Long:  ``,
			Run:   nil,
		}
		rootCommand.AddCommand(command)

		subCommand := &cobra.Command{
			Use:   "create <incident-id> <json>",
			Short: "Create an exposure",
			Long:  ``,
			Args:  cobra.ExactArgs(2),
			Run:   doIncidentExposureCreate,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "delete <incident-id> <exposure-id>",
			Short: "Delete an exposure",
			Long:  ``,
			Args:  cobra.ExactArgs(2),
			Run:   doIncidentExposureDelete,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "get <incident-id> <filter>",
			Short: "Get an exposure",
			Long:  ``,
			Args:  cobra.ExactArgs(2),
			Run:   doIncidentExposureGet,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "list <incident-id> [<filter>]",
			Short: "List all exposures",
			Long: `
Example filter: 'incidentID eq 1234'
			`,
			Args: cobra.MinimumNArgs(1),
			Run:  doIncidentExposureList,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "patch <incident-id> <exposure-id> <json>",
			Short: "Get an exposure",
			Long:  ``,
			Args:  cobra.ExactArgs(3),
			Run:   doIncidentExposurePatch,
		}
		command.AddCommand(subCommand)
	}
	{
		command := &cobra.Command{
			Use:   "exposure-location",
			Short: "Exposure location sub-command",
			Long:  ``,
			Run:   nil,
		}
		rootCommand.AddCommand(command)

		subCommand := &cobra.Command{
			Use:   "get <incident-id>",
			Short: "Get an exposure location",
			Long:  ``,
			Args:  cobra.ExactArgs(1),
			Run:   doExposureLocationGet,
		}
		command.AddCommand(subCommand)
	}
	{
		command := &cobra.Command{
			Use:   "exposure-member",
			Short: "Exposure member sub-command",
			Long:  ``,
			Run:   nil,
		}
		rootCommand.AddCommand(command)

		subCommand := &cobra.Command{
			Use:   "list <exposure-id>",
			Short: "List the members",
			Long:  ``,
			Args:  cobra.ExactArgs(1),
			Run:   doExposureMemberList,
		}
		command.AddCommand(subCommand)
	}
	{
		command := &cobra.Command{
			Use:   "exposure-user-role",
			Short: "Exposure user-role sub-command",
			Long:  ``,
			Run:   nil,
		}
		rootCommand.AddCommand(command)

		subCommand := &cobra.Command{
			Use:   "list <exposure-user-id>",
			Short: "List the roles",
			Long:  ``,
			Args:  cobra.ExactArgs(1),
			Run:   doExposureUserRoleList,
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
			Use:   "delete <id> [...]",
			Short: "Delete an incident",
			Long:  ``,
			Args:  cobra.MinimumNArgs(1),
			Run:   doIncidentDelete,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "get <filter> [exposure ...]",
			Short: "Get an incident",
			Long:  ``,
			Args:  cobra.MinimumNArgs(1),
			Run:   doIncidentGet,
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
			Use:   "get <filter>",
			Short: "Get a user",
			Long:  ``,
			Args:  cobra.ExactArgs(1),
			Run:   doUserGet,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "id <user-id>",
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

		subCommand = &cobra.Command{
			Use:   "patch <user-id> <operation> <path> <value>",
			Short: "Patch a user",
			Long:  ``,
			Args:  cobra.ExactArgs(4),
			Run:   doUserPatch,
		}
		command.AddCommand(subCommand)
	}

	{
		command := &cobra.Command{
			Use:   "user-contact-info",
			Short: "User contact info sub-command",
			Long:  ``,
			Run:   nil,
		}
		rootCommand.AddCommand(command)

		subCommand := &cobra.Command{
			Use:   "id <user-id>",
			Short: "Get user contact info by user ID",
			Long:  ``,
			Args:  cobra.ExactArgs(1),
			Run:   doUserContactInfoID,
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
		_ = cmd.Help()
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

func doLogin(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)
	fmt.Printf("Token: %s", client.Token)
}

func doRaw(cmd *cobra.Command, args []string, method string, parameters []string, headers []string, contents string) {
	client := makeClient(cmd)

	targetURL := args[0]
	args = args[1:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	optionsMap := map[string]string{}
	for _, parameter := range parameters {
		parts := strings.SplitN(parameter, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		optionsMap[key] = value
	}

	headersMap := map[string]string{}
	for _, parameter := range headers {
		parts := strings.SplitN(parameter, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		headersMap[key] = value
	}

	response, err := client.RawOperation(method, targetURL, optionsMap, headersMap, []byte(contents))
	if err != nil {
		panic(err)
	}

	jsonBytes, err := json.MarshalIndent(response, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doApparatusGet(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing filter")
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
		jsonBytes, err := json.MarshalIndent(currentApparatus, "" /*prefix*/, "\t" /*indent*/)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonBytes))
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
	jsonBytes, err := json.MarshalIndent(apparatusesResponse.Apparatuses, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentCreate(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing JSON")
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
	jsonBytes, err := json.MarshalIndent(postIncidentResponse, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentDelete(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	for _, incidentID := range args {
		err := client.DeleteIncident(incidentID)
		if err != nil {
			panic(err)
		}
	}
}

func doIncidentGet(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing filter")
	}
	filter := args[0]
	args = args[1:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	var currentIncident *emergencyreporting.Incident
	{
		incidentsResponse, err := client.GetIncidents(map[string]string{"filter": filter})
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

	jsonBytes, err := json.MarshalIndent(currentIncident, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
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

	incidentsResponse, err := client.GetIncidents(options)
	if err != nil {
		panic(err)
	}
	jsonBytes, err := json.MarshalIndent(incidentsResponse.Incidents, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentExposureCreate(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing incident ID")
	}
	incidentID := args[0]
	if len(args) < 2 {
		panic("Missing JSON")
	}
	jsonString := args[1]
	args = args[2:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	var exposure emergencyreporting.Exposure
	err := json.Unmarshal([]byte(jsonString), &exposure)
	if err != nil {
		panic(fmt.Errorf("Could not parse JSON: %v", err))
	}
	postExposureResponse, err := client.PostIncidentExposure(incidentID, exposure)
	if err != nil {
		panic(fmt.Errorf("Could not create exposure: %v", err))
	}
	jsonBytes, err := json.MarshalIndent(postExposureResponse, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doExposureGet(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing argument: filter (example: 'incidentID eq 1234')")
	}
	filter := args[0]
	args = args[1:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	var currentExposure *emergencyreporting.Exposure
	{
		exposuresResponse, err := client.GetExposures(map[string]string{"filter": filter})
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

	jsonBytes, err := json.MarshalIndent(currentExposure, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentExposureDelete(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing incident ID")
	}
	incidentID := args[0]
	if len(args) < 2 {
		panic("Missing exposure ID")
	}
	exposureID := args[1]
	args = args[2:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	err := client.DeleteIncidentExposure(incidentID, exposureID)
	if err != nil {
		panic(err)
	}
}

func doExposureList(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing argument: filter (example: 'exposureID eq 1234')")
	}
	filter := args[0]
	args = args[1:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	options := map[string]string{
		"filter": filter,
		"limit":  cmd.Flag("limit").Value.String(),
	}

	exposuresResponse, err := client.GetExposures(options)
	if err != nil {
		panic(err)
	}

	jsonBytes, err := json.MarshalIndent(exposuresResponse.Exposures, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentExposureGet(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing incident ID")
	}
	incidentID := args[0]
	if len(args) < 2 {
		panic("Missing argument: filter (example: 'exposureID eq 1234')")
	}
	filter := args[1]
	args = args[2:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	var currentExposure *emergencyreporting.Exposure
	{
		exposuresResponse, err := client.GetIncidentExposures(incidentID, map[string]string{"filter": filter})
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

	jsonBytes, err := json.MarshalIndent(currentExposure, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentExposureList(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) == 0 {
		panic("Missing incident ID")
	}
	incidentID := args[0]
	args = args[1:]

	var filter string
	if len(args) >= 2 {
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

	exposuresResponse, err := client.GetIncidentExposures(incidentID, options)
	if err != nil {
		panic(err)
	}

	jsonBytes, err := json.MarshalIndent(exposuresResponse.Exposures, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentExposurePatch(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing incident ID")
	}
	incidentID := args[0]
	if len(args) < 2 {
		panic("Missing exposure ID")
	}
	exposureID := args[1]
	if len(args) < 3 {
		panic("Missing incident ID")
	}
	contents := args[2]
	args = args[3:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	var currentExposure *emergencyreporting.Exposure
	{
		exposureResponse, err := client.GetIncidentExposure(incidentID, exposureID)
		if err != nil {
			panic(err)
		}
		currentExposure = exposureResponse.Exposure
	}
	if currentExposure == nil {
		fmt.Printf("Exposure not found.\n")
		return
	}

	var patchExposureRequest emergencyreporting.PatchExposureRequest
	err := json.Unmarshal([]byte(contents), &patchExposureRequest)
	if err != nil {
		panic(err)
	}

	patchExposureResponse, err := client.PatchIncidentExposure(incidentID, exposureID, currentExposure.RowVersion, patchExposureRequest)
	if err != nil {
		panic(err)
	}

	jsonBytes, err := json.MarshalIndent(patchExposureResponse, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doExposureLocationGet(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing argument: exposure ID")
	}
	exposureID := args[0]
	args = args[1:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	response, err := client.GetExposureLocation(exposureID)
	if err != nil {
		panic(err)
	}

	jsonBytes, err := json.MarshalIndent(response.Location, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doExposureMemberGet(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing argument: exposure ID")
	}
	exposureID := args[0]
	if len(args) < 2 {
		panic("Missing argument: filter (example: 'exposureUserID eq 1234')")
	}
	filter := args[1]
	args = args[2:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

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

	jsonBytes, err := json.MarshalIndent(currentMember, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doExposureMemberList(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing exposure ID")
	}
	exposureID := args[0]
	args = args[1:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	options := map[string]string{
		"limit": cmd.Flag("limit").Value.String(),
	}

	membersResponse, err := client.GetExposureMembers(exposureID, options)
	if err != nil {
		panic(err)
	}
	jsonBytes, err := json.MarshalIndent(membersResponse.CrewMembers, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doExposureUserRoleList(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing exposure user ID")
	}
	exposureUserID := args[0]
	args = args[1:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	options := map[string]string{
		"limit": cmd.Flag("limit").Value.String(),
	}

	rolesResponse, err := client.GetExposureMemberRoles(exposureUserID, options)
	if err != nil {
		panic(err)
	}
	jsonBytes, err := json.MarshalIndent(rolesResponse.Roles, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doStationGet(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing filter")
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
		jsonBytes, err := json.MarshalIndent(currentStation, "" /*prefix*/, "\t" /*indent*/)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonBytes))
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
	jsonBytes, err := json.MarshalIndent(stationsResponse.Stations, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doUserGet(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing filter")
	}
	filter := args[0]
	args = args[1:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	var currentUser *emergencyreporting.User
	{
		usersResponse, err := client.GetUsers(map[string]string{"filter": filter})
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
	jsonBytes, err := json.MarshalIndent(currentUser, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doUserID(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing user ID")
	}
	userID := args[0]
	args = args[1:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	var currentUser *emergencyreporting.User
	{
		userResponse, err := client.GetUser(userID)
		if err != nil {
			panic(err)
		}
		currentUser = userResponse.User
	}
	if currentUser == nil {
		fmt.Printf("User not found.\n")
		return
	}
	jsonBytes, err := json.MarshalIndent(currentUser, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doUserPatch(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missing user ID")
	}
	userID := args[0]

	if len(args) < 2 {
		panic("Missing operation")
	}
	operation := args[1]
	if len(args) < 3 {
		panic("Missing path")
	}
	path := args[2]
	if len(args) < 4 {
		panic("Missing value")
	}
	value := args[3]
	args = args[4:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	var currentUser *emergencyreporting.User
	{
		userResponse, err := client.GetUser(userID)
		if err != nil {
			panic(err)
		}
		currentUser = userResponse.User
	}
	if currentUser == nil {
		fmt.Printf("User not found.\n")
		return
	}

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
	jsonBytes, err := json.MarshalIndent(patchResponse, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
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
	usersResponse, err := client.GetUsers(options)
	if err != nil {
		panic(err)
	}
	jsonBytes, err := json.MarshalIndent(usersResponse.Users, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func doUserContactInfoID(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)

	if len(args) < 1 {
		panic("Missin guser ID")
	}
	userID := args[0]
	args = args[1:]
	if len(args) > 0 {
		panic("Too many arguments")
	}

	getUserContactInfoResponse, err := client.GetUserContactInfo(userID)
	if err != nil {
		panic(fmt.Errorf("Could not get user contact info for user ID %s: %v", userID, err))
	}
	jsonBytes, err := json.MarshalIndent(getUserContactInfoResponse.ContactInfo, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}
