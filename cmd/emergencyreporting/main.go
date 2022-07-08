package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
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
			_ = cmd.Help()
			os.Exit(1)
		},
	}
	rootCommand.PersistentFlags().String("config", "config.json", "Path to the configuration file with the client credentials.")
	rootCommand.PersistentFlags().Int("limit", 100, "The page size for any queries.")
	rootCommand.PersistentFlags().String("token", "", "The Emergency Reporting token to use.  If this is not set, then this will attempt to log in and get a token.")

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
			Use:   "get <apparatus-id>",
			Short: "Get an apparatus",
			Long:  ``,
			Args:  cobra.ExactArgs(1),
			Run:   doApparatusGet,
		}
		command.AddCommand(subCommand)

		subCommand = &cobra.Command{
			Use:   "list [<filter>]",
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
			Use:   "get <incident-id> <exposure-id>",
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

		subCommand = &cobra.Command{
			Use:   "get <exposure-id> <exposure-user-id>",
			Short: "Get the member",
			Long:  ``,
			Args:  cobra.ExactArgs(2),
			Run:   doExposureMemberGet,
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
			Use:   "get <incident-id>",
			Short: "Get an incident",
			Long:  ``,
			Args:  cobra.ExactArgs(1),
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
			Use:   "get <station-id>",
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
			Use:   "get <user-id>",
			Short: "Get a user",
			Long:  ``,
			Args:  cobra.ExactArgs(1),
			Run:   doUserGet,
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
		logrus.Errorf("Could not execute root comamand: [%T] %v", err, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func makeClient(cmd *cobra.Command) *emergencyreporting.Client {
	ctx := context.Background()

	var configFile string
	{
		flag := cmd.Flag("config")
		configFile = flag.Value.String()
	}

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

	var token string
	{
		flag := cmd.Flag("token")
		token = flag.Value.String()
	}

	if token == "" {
		if len(client.Token) == 0 {
			tokenResponse, err := client.GenerateToken(ctx)
			if err != nil {
				fmt.Printf("Could not generate token: %v\n", err)
				os.Exit(1)
			}
			client.Token = tokenResponse.AccessToken
		}
	} else {
		client.Token = token
	}

	return client
}

func doLogin(cmd *cobra.Command, args []string) {
	client := makeClient(cmd)
	fmt.Printf("%s\n", client.Token)
}

func doRaw(cmd *cobra.Command, args []string, method string, parameters []string, headers []string, contents string) {
	ctx := context.Background()
	client := makeClient(cmd)

	targetURL := args[0]
	args = args[1:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
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

	response, err := client.RawOperation(ctx, method, targetURL, optionsMap, headersMap, []byte(contents))
	if err != nil {
		logrus.Errorf("Could not perform raw operation: [%T] %v", err, err)
		os.Exit(1)
	}

	jsonBytes, err := json.MarshalIndent(response, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doApparatusGet(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing apparatus ID")
		os.Exit(1)
	}
	apparatusID := args[0]
	args = args[1:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	apparatusResponse, err := client.GetApparatus(ctx, apparatusID)
	if err != nil {
		logrus.Errorf("Could not get apparatus: [%T] %v", err, err)
		os.Exit(1)
	}

	jsonBytes, err := json.MarshalIndent(apparatusResponse.Apparatus, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doApparatusList(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	filter := ""
	if len(args) > 0 {
		filter = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	options := map[string]string{
		"filter": filter,
		"limit":  cmd.Flag("limit").Value.String(),
	}
	apparatusesResponse, err := client.GetApparatuses(ctx, options)
	if err != nil {
		logrus.Errorf("Could not get apparatuses: [%T] %v", err, err)
		os.Exit(1)
	}
	jsonBytes, err := json.MarshalIndent(apparatusesResponse.Apparatuses, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentCreate(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing JSON")
		os.Exit(1)
	}
	jsonString := args[0]
	args = args[1:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	var incident emergencyreporting.Incident
	err := json.Unmarshal([]byte(jsonString), &incident)
	if err != nil {
		logrus.Errorf("Could not parse JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	postIncidentResponse, err := client.PostIncident(ctx, incident)
	if err != nil {
		logrus.Errorf("Could not create incident: [%T] %v", err, err)
		os.Exit(1)
	}
	jsonBytes, err := json.MarshalIndent(postIncidentResponse, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentDelete(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	for _, incidentID := range args {
		err := client.DeleteIncident(ctx, incidentID)
		if err != nil {
			logrus.Errorf("Could not delete incident '%s': [%T] %v", incidentID, err, err)
			os.Exit(1)
		}
	}
}

func doIncidentGet(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing incident ID")
		os.Exit(1)
	}
	incidentID := args[0]
	args = args[1:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	incidentResponse, err := client.GetIncident(ctx, incidentID)
	if err != nil {
		logrus.Errorf("Could not get incident '%s': [%T] %v", incidentID, err, err)
		os.Exit(1)
	}

	jsonBytes, err := json.MarshalIndent(incidentResponse.Incident, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentList(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	filter := ""
	if len(args) > 0 {
		filter = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	options := map[string]string{
		"filter": filter,
		"limit":  cmd.Flag("limit").Value.String(),
	}

	incidentsResponse, err := client.GetIncidents(ctx, options)
	if err != nil {
		logrus.Errorf("Could not get incidents: [%T] %v", err, err)
		os.Exit(1)
	}
	jsonBytes, err := json.MarshalIndent(incidentsResponse.Incidents, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentExposureCreate(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing incident ID")
		os.Exit(1)
	}
	incidentID := args[0]
	if len(args) < 2 {
		logrus.Errorf("Missing JSON")
		os.Exit(1)
	}
	jsonString := args[1]
	args = args[2:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	var exposure emergencyreporting.Exposure
	err := json.Unmarshal([]byte(jsonString), &exposure)
	if err != nil {
		logrus.Errorf("Could not parse JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	postExposureResponse, err := client.PostIncidentExposure(ctx, incidentID, exposure)
	if err != nil {
		logrus.Errorf("Could not create exposure: [%T] %v", err, err)
		os.Exit(1)
	}
	jsonBytes, err := json.MarshalIndent(postExposureResponse, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentExposureDelete(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing incident ID")
		os.Exit(1)
	}
	incidentID := args[0]
	if len(args) < 2 {
		logrus.Errorf("Missing exposure ID")
		os.Exit(1)
	}
	exposureID := args[1]
	args = args[2:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	err := client.DeleteIncidentExposure(ctx, incidentID, exposureID)
	if err != nil {
		logrus.Errorf("Could not delete exposure: [%T] %v", err, err)
		os.Exit(1)
	}
}

func doExposureList(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	var filter string
	if len(args) > 0 {
		filter = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	options := map[string]string{
		"filter": filter,
		"limit":  cmd.Flag("limit").Value.String(),
	}

	exposuresResponse, err := client.GetExposures(ctx, options)
	if err != nil {
		logrus.Errorf("Could not get exposures: [%T] %v", err, err)
		os.Exit(1)
	}

	jsonBytes, err := json.MarshalIndent(exposuresResponse.Exposures, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentExposureGet(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing incident ID")
		os.Exit(1)
	}
	incidentID := args[0]
	if len(args) < 2 {
		logrus.Errorf("Missing exposure ID")
		os.Exit(1)
	}
	exposureID := args[1]
	args = args[2:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	exposureResponse, err := client.GetIncidentExposure(ctx, incidentID, exposureID)
	if err != nil {
		logrus.Errorf("Could not get exposure: [%T] %v", err, err)
		os.Exit(1)
	}

	jsonBytes, err := json.MarshalIndent(exposureResponse.Exposure, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentExposureList(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) == 0 {
		logrus.Errorf("Missing incident ID")
		os.Exit(1)
	}
	incidentID := args[0]
	args = args[1:]

	var filter string
	if len(args) >= 2 {
		filter = args[0]
		args = args[1:]
	}

	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	options := map[string]string{
		"filter": filter,
		"limit":  cmd.Flag("limit").Value.String(),
	}

	exposuresResponse, err := client.GetIncidentExposures(ctx, incidentID, options)
	if err != nil {
		logrus.Errorf("Could not get exposures: [%T] %v", err, err)
		os.Exit(1)
	}

	jsonBytes, err := json.MarshalIndent(exposuresResponse.Exposures, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doIncidentExposurePatch(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing incident ID")
		os.Exit(1)
	}
	incidentID := args[0]
	if len(args) < 2 {
		logrus.Errorf("Missing exposure ID")
		os.Exit(1)
	}
	exposureID := args[1]
	if len(args) < 3 {
		logrus.Errorf("Missing incident ID")
		os.Exit(1)
	}
	contents := args[2]
	args = args[3:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	var currentExposure *emergencyreporting.Exposure
	{
		exposureResponse, err := client.GetIncidentExposure(ctx, incidentID, exposureID)
		if err != nil {
			logrus.Errorf("Could not get exposure: [%T] %v", err, err)
			os.Exit(1)
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
		logrus.Errorf("Could not parse JSON: [%T] %v", err, err)
		os.Exit(1)
	}

	patchExposureResponse, err := client.PatchIncidentExposure(ctx, incidentID, exposureID, currentExposure.RowVersion, patchExposureRequest)
	if err != nil {
		logrus.Errorf("Error patching exposure: [%T] %v", err, err)
		os.Exit(1)
	}

	jsonBytes, err := json.MarshalIndent(patchExposureResponse, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doExposureLocationGet(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing argument: exposure ID")
		os.Exit(1)
	}
	exposureID := args[0]
	args = args[1:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	response, err := client.GetExposureLocation(ctx, exposureID)
	if err != nil {
		logrus.Errorf("Could not get exposure location: [%T] %v", err, err)
		os.Exit(1)
	}

	jsonBytes, err := json.MarshalIndent(response.Location, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doExposureMemberGet(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing exposure ID")
		os.Exit(1)
	}
	exposureID := args[0]
	if len(args) < 2 {
		logrus.Errorf("Missing exposure user ID")
		os.Exit(1)
	}
	exposureUserID := args[1]
	args = args[2:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	memberResponse, err := client.GetExposureMember(ctx, exposureID, exposureUserID)
	if err != nil {
		logrus.Errorf("Could not get exposure member: [%T] %v", err, err)
		os.Exit(1)
	}

	jsonBytes, err := json.MarshalIndent(memberResponse.CrewMember, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doExposureMemberList(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing exposure ID")
		os.Exit(1)
	}
	exposureID := args[0]
	args = args[1:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	options := map[string]string{
		"limit": cmd.Flag("limit").Value.String(),
	}

	membersResponse, err := client.GetExposureMembers(ctx, exposureID, options)
	if err != nil {
		logrus.Errorf("Could not get exposure members: [%T] %v", err, err)
		os.Exit(1)
	}
	jsonBytes, err := json.MarshalIndent(membersResponse.CrewMembers, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doExposureUserRoleList(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing exposure user ID")
		os.Exit(1)
	}
	exposureUserID := args[0]
	args = args[1:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	options := map[string]string{
		"limit": cmd.Flag("limit").Value.String(),
	}

	rolesResponse, err := client.GetExposureMemberRoles(ctx, exposureUserID, options)
	if err != nil {
		logrus.Errorf("Could not get exposure member roles: [%T] %v", err, err)
		os.Exit(1)
	}
	jsonBytes, err := json.MarshalIndent(rolesResponse.Roles, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doStationGet(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing filter")
		os.Exit(1)
	}
	filter := args[0]
	args = args[1:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	var currentStation *emergencyreporting.Station
	{
		apparatusesResponse, err := client.GetStations(ctx, map[string]string{"filter": filter})
		if err != nil {
			logrus.Errorf("Could not get stations: [%T] %v", err, err)
			os.Exit(1)
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
			logrus.Errorf("Error writing JSON: [%T] %v", err, err)
			os.Exit(1)
		}
		fmt.Println(string(jsonBytes))
	}
}

func doStationList(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	filter := ""
	if len(args) > 0 {
		filter = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	options := map[string]string{
		"filter": filter,
		"limit":  cmd.Flag("limit").Value.String(),
	}
	stationsResponse, err := client.GetStations(ctx, options)
	if err != nil {
		logrus.Errorf("Could not get stations: [%T] %v", err, err)
		os.Exit(1)
	}
	jsonBytes, err := json.MarshalIndent(stationsResponse.Stations, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doUserGet(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing user ID")
		os.Exit(1)
	}
	userID := args[0]
	args = args[1:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	var currentUser *emergencyreporting.User
	{
		userResponse, err := client.GetUser(ctx, userID)
		if err != nil {
			logrus.Errorf("Could not get user: [%T] %v", err, err)
			os.Exit(1)
		}
		currentUser = userResponse.User
	}
	if currentUser == nil {
		fmt.Printf("User not found.\n")
		return
	}
	jsonBytes, err := json.MarshalIndent(currentUser, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doUserPatch(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missing user ID")
		os.Exit(1)
	}
	userID := args[0]

	if len(args) < 2 {
		logrus.Errorf("Missing operation")
		os.Exit(1)
	}
	operation := args[1]
	if len(args) < 3 {
		logrus.Errorf("Missing path")
		os.Exit(1)
	}
	path := args[2]
	if len(args) < 4 {
		logrus.Errorf("Missing value")
		os.Exit(1)
	}
	value := args[3]
	args = args[4:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	var currentUser *emergencyreporting.User
	{
		userResponse, err := client.GetUser(ctx, userID)
		if err != nil {
			logrus.Errorf("Could not get user: [%T] %v", err, err)
			os.Exit(1)
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
	patchResponse, err := client.PatchUser(ctx, currentUser.UserID, currentUser.RowVersion, patchUserRequest)
	if err != nil {
		logrus.Errorf("Could not patch user: [%T] %v", err, err)
		os.Exit(1)
	}
	jsonBytes, err := json.MarshalIndent(patchResponse, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doUserList(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	filter := ""
	if len(args) > 0 {
		filter = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	options := map[string]string{
		"filter": filter,
		"limit":  cmd.Flag("limit").Value.String(),
	}
	usersResponse, err := client.GetUsers(ctx, options)
	if err != nil {
		logrus.Errorf("Could not get users: [%T] %v", err, err)
		os.Exit(1)
	}
	jsonBytes, err := json.MarshalIndent(usersResponse.Users, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func doUserContactInfoID(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client := makeClient(cmd)

	if len(args) < 1 {
		logrus.Errorf("Missin guser ID")
		os.Exit(1)
	}
	userID := args[0]
	args = args[1:]
	if len(args) > 0 {
		logrus.Errorf("Too many arguments")
		os.Exit(1)
	}

	getUserContactInfoResponse, err := client.GetUserContactInfo(ctx, userID)
	if err != nil {
		logrus.Errorf("Could not get user contact info for user ID %s: [%T] %v", userID, err, err)
		os.Exit(1)
	}
	jsonBytes, err := json.MarshalIndent(getUserContactInfoResponse.ContactInfo, "" /*prefix*/, "\t" /*indent*/)
	if err != nil {
		logrus.Errorf("Error writing JSON: [%T] %v", err, err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}
