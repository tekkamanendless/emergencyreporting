package emergencyreporting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// ErrorDuplicate represents a duplicate key error.
var ErrorDuplicate = fmt.Errorf("Duplicate")

// ErrorNotFound represents a 404 "not found" error.
var ErrorNotFound = fmt.Errorf("NotFound")

// Logger is the basic logger for this package.
type Logger interface {
	Printf(format string, v ...interface{})
}

// DefaultLogger is the default logger that uses the "log" package.
type DefaultLogger struct{}

func (DefaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// NullLogger is a logger that doesn't log anything.
type NullLogger struct{}

func (NullLogger) Printf(format string, v ...interface{}) {}

// Client is the client interface for communicating with the Emergency Reporting API.
type Client struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ClientID        string `json:"client_id"`
	ClientSecret    string `json:"client_secret"`
	AccountID       string `json:"account_id"`       // (New password authentication) This is the user's account ID.
	UserID          string `json:"user_id"`          // (New password authentication) This is the user's ID.
	Token           string `json:"token"`            // Required, but can be generated using the username, etc.
	SubscriptionKey string `json:"subscription_key"` // Required no matter what.
	Logger          Logger // This is the Logger instance to use.  If empty, then the default one will be used.

	client http.Client
}

// init makes sure that everything is initialized.
func (c *Client) init() {
	if c.Logger == nil {
		c.Logger = &DefaultLogger{}
	}
}

func (c *Client) GenerateToken() (*GenerateTokenResponse, error) {
	c.init()

	targetURL := "https://auth.emergencyreporting.com/Token.php"
	values := url.Values{
		"grant_type":    {"password"},
		"username":      {c.Username},
		"password":      {c.Password},
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
	}

	if c.AccountID != "" && c.UserID != "" {
		tenantHost := "login.emergencyreporting.com"
		tenantSegment := "login.emergencyreporting.com"

		goLiveDate, _ := time.Parse("2006-01-02", "2020-12-06")
		if time.Now().Before(goLiveDate) {
			tenantHost = "loginrc.emergencyreporting.com"
			tenantSegment = "loginrc.emergencyreporting.com"
		}

		targetURL = "https://" + tenantHost + "/" + tenantSegment + "/B2C_1A_PasswordGrant/oauth2/v2.0/token"

		values.Set("scope", "https://"+tenantSegment+"/secure/full_access")
		values.Set("response_type", "token")
		values.Set("er_aid", c.AccountID)
		values.Set("er_uid", c.UserID)
	}

	c.Logger.Printf("POST %s\n", targetURL)
	response, err := c.client.PostForm(targetURL, values)
	if err != nil {
		return nil, fmt.Errorf("Could not post form: %v", err)
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Could not read body: %v", err)
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		c.Logger.Printf("Body: %v\n", string(contents))
		return nil, fmt.Errorf("Bad status code: %d", response.StatusCode)
	}

	var parsedResponse GenerateTokenResponse
	err = json.Unmarshal(contents, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not parse JSON: %v", err)
	}

	return &parsedResponse, nil
}

// RawOperation performs a raw HTTP request.
func (c *Client) RawOperation(method string, targetURL string, options map[string]string, headers map[string]string, body []byte) (json.RawMessage, error) {
	var response json.RawMessage
	err := c.internalRequest(method, targetURL, options, headers, body, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) internalRequest(method string, targetURL string, options map[string]string, headers map[string]string, body []byte, targetPointer interface{}) error {
	c.init()

	queryParts := url.Values{}
	for key, value := range options {
		queryParts.Set(key, value)
	}
	if len(queryParts) > 0 {
		targetURL += "?" + queryParts.Encode()
	}

	c.Logger.Printf("%s %s\n", method, targetURL)
	request, err := http.NewRequest(method, targetURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("Could not make request: %v", err)
	}
	request.Header.Set("Authorization", c.Token)
	request.Header.Set("Ocp-Apim-Subscription-Key", c.SubscriptionKey)
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("Could not perform operation: %v", err)
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Could not read body: %v", err)
	}
	c.Logger.Printf("%s %s %d %d\n", method, targetURL, response.StatusCode, len(contents))

	if response.StatusCode < 200 || response.StatusCode > 299 {
		var errorType string
		var errorMessage string
		{
			var errorResponse ErrorResponse
			err = json.Unmarshal(contents, &errorResponse)
			if err != nil {
				// Oh well.
			} else {
				if len(errorResponse.Errors) > 0 {
					errorType = errorResponse.Errors[0].Type
					errorMessage = errorResponse.Errors[0].Message
				}
			}
		}
		if errorType == "" {
			var errorResponse ErrorResponseBuggy
			err = json.Unmarshal(contents, &errorResponse)
			if err != nil {
				// Oh well.
			} else {
				errorType = errorResponse.Errors.Type
				errorMessage = errorResponse.Errors.Message
			}
		}
		if err != nil || errorType == "" {
			if response.StatusCode == http.StatusNotFound {
				return ErrorNotFound
			}

			c.Logger.Printf("Error: %s\n", string(contents))
			return fmt.Errorf("Bad status code: %d", response.StatusCode)
		}
		switch errorType {
		case "Duplicate":
			return ErrorDuplicate
		default:
			return fmt.Errorf("Error type: %s (%s); status code: %d", errorType, errorMessage, response.StatusCode)
		}
	}
	// DEBUG:
	//c.Logger.Printf("%s\n", contents)
	// :GUBED

	if targetPointer != nil {
		err = json.Unmarshal(contents, targetPointer)
		if err != nil {
			return fmt.Errorf("Could not parse JSON: %v", err)
		}
	}

	return nil
}

// GetStations TODO
// See: https://developer.emergencyreporting.com/docs/services/stations/operations/get-stations?
func (c *Client) GetStations(options map[string]string) (*GetStationsResponse, error) {
	// https://data.emergencyreporting.com/agencystations/stations[?rowVersion][&changesSince][&limit][&offset][&showArchived][&filter]

	targetURL := "https://data.emergencyreporting.com/agencystations/stations"

	var parsedResponse GetStationsResponse

	err := c.internalRequest(http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the stations: %v", err)
	}

	return &parsedResponse, nil
}

// GetIncidents TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/getIncidents?
func (c *Client) GetIncidents(options map[string]string) (*GetIncidentsResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/incidents[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/incidents"

	var parsedResponse GetIncidentsResponse

	err := c.internalRequest(http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the incidents: %v", err)
	}

	return &parsedResponse, nil
}

// PostIncident TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/postIncidents?
func (c *Client) PostIncident(incident Incident) (*PostIncidentResponse, error) {
	c.init()

	// https://data.emergencyreporting.com/agencyincidents/incidents[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/incidents"

	jsonInput, err := json.Marshal(incident)
	if err != nil {
		return nil, fmt.Errorf("Could not create JSON: %v", err)
	}
	c.Logger.Printf("JSON input: %s\n", string(jsonInput))

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var parsedResponse PostIncidentResponse

	err = c.internalRequest(http.MethodPost, targetURL, nil, headers, jsonInput, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not create the incident: %v", err)
	}

	return &parsedResponse, nil
}

// DeleteIncident TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/deleteIncident?
func (c *Client) DeleteIncident(incidentID string) error {
	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}

	targetURL := "https://data.emergencyreporting.com/agencyincidents/incidents/" + url.PathEscape(incidentID)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	err := c.internalRequest(http.MethodDelete, targetURL, nil, headers, nil, nil)
	if err != nil {
		return fmt.Errorf("Could not delete the incident: %v", err)
	}

	return nil
}

// GetExposures TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/IncidentsExposuresByIncidentIDGet?
func (c *Client) GetExposures(incidentID string, options map[string]string) (*GetExposuresResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}/exposures[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/incidents/" + url.PathEscape(incidentID) + "/exposures"

	var parsedResponse GetExposuresResponse

	err := c.internalRequest(http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the exposures: %v", err)
	}

	return &parsedResponse, nil
}

// GetExposure TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/IncidentsExposuresByIncidentIDAndExposureIDGet?
func (c *Client) GetExposure(incidentID string, exposureID string) (*GetExposureResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}/exposures/{exposureID}

	targetURL := "https://data.emergencyreporting.com/agencyincidents/incidents/" + url.PathEscape(incidentID) + "/exposures/" + url.PathEscape(exposureID)

	var parsedResponse GetExposureResponse

	err := c.internalRequest(http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the exposure: %v", err)
	}

	return &parsedResponse, nil
}

// PostExposure TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/IncidentsExposuresByIncidentIDPost?
func (c *Client) PostExposure(incidentID string, exposure Exposure) (*PostExposureResponse, error) {
	c.init()

	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}/exposures

	targetURL := "https://data.emergencyreporting.com/agencyincidents/incidents/" + url.PathEscape(incidentID) + "/exposures"

	jsonInput, err := json.Marshal(exposure)
	if err != nil {
		return nil, fmt.Errorf("Could not create JSON: %v", err)
	}
	c.Logger.Printf("JSON input: %s\n", string(jsonInput))

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var parsedResponse PostExposureResponse

	err = c.internalRequest(http.MethodPost, targetURL, nil, headers, jsonInput, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not create the exposure: %v", err)
	}

	return &parsedResponse, nil
}

// DeleteExposure TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/IncidentsExposuresByIncidentIDAndExposureIDDelete?
func (c *Client) DeleteExposure(incidentID string, exposureID string) error {
	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}/exposures/{exposureID}

	targetURL := "https://data.emergencyreporting.com/agencyincidents/incidents/" + url.PathEscape(incidentID) + "/exposures/" + url.PathEscape(exposureID)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	err := c.internalRequest(http.MethodDelete, targetURL, nil, headers, nil, nil)
	if err != nil {
		return fmt.Errorf("Could not delete the exposure: %v", err)
	}

	return nil
}

// GetExposureLocation TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresLocationByExposureIDGet?
func (c *Client) GetExposureLocation(exposureID string) (*GetExposureLocationResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/location[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/location"

	var parsedResponse GetExposureLocationResponse

	err := c.internalRequest(http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the exposure location: %v", err)
	}

	return &parsedResponse, nil
}

// PutExposureLocation TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresLocationByExposureIDGet?
func (c *Client) PutExposureLocation(exposureID string, location ExposureLocation) (*PutExposureLocationResponse, error) {
	c.init()

	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/location[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/location"

	jsonInput, err := json.Marshal(location)
	if err != nil {
		return nil, fmt.Errorf("Could not create JSON: %v", err)
	}
	c.Logger.Printf("JSON input: %s\n", string(jsonInput))

	headers := map[string]string{
		"Content-Type": "application/json",
		"ETag":         location.RowVersion,
	}

	var parsedResponse PutExposureLocationResponse

	err = c.internalRequest(http.MethodPut, targetURL, nil, headers, jsonInput, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not put the exposure location: %v", err)
	}

	return &parsedResponse, nil
}

// GetExposureFire TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresFireByExposureIDGet?
func (c *Client) GetExposureFire(exposureID string) (*GetExposureFireResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/fire[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/fire"

	var parsedResponse GetExposureFireResponse

	err := c.internalRequest(http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		if err == ErrorNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("Could not get the exposure fire: %v", err)
	}

	return &parsedResponse, nil
}

// GetExposureApparatuses TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresApparatusesByExposureIDGet?
func (c *Client) GetExposureApparatuses(exposureID string) (*GetExposureApparatusesResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/apparatuses[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/apparatuses"

	var parsedResponse GetExposureApparatusesResponse

	err := c.internalRequest(http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the exposure apparatuses: %v", err)
	}

	return &parsedResponse, nil
}

// PostExposureApparatus TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresApparatusesByExposureIDPost?
func (c *Client) PostExposureApparatus(exposureID string, apparatus ExposureApparatus) (*PostExposureApparatusResponse, error) {
	c.init()

	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/apparatuses[?useAssociatedAgencyApparatusID]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/apparatuses"
	options := map[string]string{
		"useAssociatedAgencyApparatusID": "1",
	}

	jsonInput, err := json.Marshal(apparatus)
	if err != nil {
		return nil, fmt.Errorf("Could not create JSON: %v", err)
	}
	c.Logger.Printf("JSON input: %s\n", string(jsonInput))

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var parsedResponse PostExposureApparatusResponse

	err = c.internalRequest(http.MethodPost, targetURL, options, headers, jsonInput, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not create the apparatus: %v", err)
	}

	return &parsedResponse, nil
}

// GetExposureMembers TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresCrewmembersByExposureIDGet?
func (c *Client) GetExposureMembers(exposureID string, options map[string]string) (*GetExposureMembersResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/crewmembers[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/crewmembers"

	var parsedResponse GetExposureMembersResponse

	err := c.internalRequest(http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the exposure members: %v", err)
	}

	return &parsedResponse, nil
}

// GetExposureMemberRoles TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/CrewmembersRolesByExposureUserIDGet?
func (c *Client) GetExposureMemberRoles(exposureID string, exposureUserID string, options map[string]string) (*GetExposureMemberRolesResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/crewmembers/{exposureUserID}/roles[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/crewmembers/" + url.PathEscape(exposureUserID) + "/roles"

	var parsedResponse GetExposureMemberRolesResponse

	err := c.internalRequest(http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the exposure member roles: %v", err)
	}

	return &parsedResponse, nil
}

// GetUsers TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-users/operations/V1UsersGet?
func (c *Client) GetUsers(options map[string]string) (*GetUsersResponse, error) {
	// https://data.emergencyreporting.com/agencyusers/users[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyusers/users"

	headers := map[string]string{}

	var parsedResponse GetUsersResponse

	err := c.internalRequest(http.MethodGet, targetURL, options, headers, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the users: %v", err)
	}

	return &parsedResponse, nil
}

// GetUser TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-users/operations/V1UsersByUserIDGet?
func (c *Client) GetUser(userID string) (*GetUserResponse, error) {
	// https://data.emergencyreporting.com/agencyusers/users/{userID}

	targetURL := "https://data.emergencyreporting.com/agencyusers/users/" + url.PathEscape(userID)

	options := map[string]string{}
	headers := map[string]string{}

	var parsedResponse GetUserResponse

	err := c.internalRequest(http.MethodGet, targetURL, options, headers, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the user: %v", err)
	}

	return &parsedResponse, nil
}

// GetExposureApparatuses TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-users/operations/V1UsersContactinfoByUserIDGet?
func (c *Client) GetUserContactInfo(userID string) (*GetUserContactInfoResponse, error) {
	// https://data.emergencyreporting.com/agencyusers/users/{userID}/contactinfo

	targetURL := "https://data.emergencyreporting.com/agencyusers/users/" + url.PathEscape(userID) + "/contactinfo"

	var parsedResponse GetUserContactInfoResponse

	err := c.internalRequest(http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the user contact info: %v", err)
	}

	return &parsedResponse, nil
}

// PatchUser TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-users/operations/V1UsersByUserIDPatch?
func (c *Client) PatchUser(userID string, rowVersion string, payload PatchUserRequest) (*PatchUserResponse, error) {
	c.init()

	// https://data.emergencyreporting.com/agencyusers/users/{userID}

	targetURL := "https://data.emergencyreporting.com/agencyusers/users/" + url.PathEscape(userID)

	jsonInput, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Could not create JSON: %v", err)
	}
	c.Logger.Printf("JSON input: %s\n", string(jsonInput))

	headers := map[string]string{
		"Content-Type": "application/json",
		"ETag":         rowVersion,
	}

	var parsedResponse PatchUserResponse

	err = c.internalRequest(http.MethodPatch, targetURL, nil, headers, jsonInput, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not patch the user: %v", err)
	}

	return &parsedResponse, nil
}

// GetApparatuses TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-apparatus/operations/ApparatusGet?
func (c *Client) GetApparatuses(options map[string]string) (*GetApparatusesResponse, error) {
	// https://data.emergencyreporting.com/agencyapparatus/apparatus[?limit][&offset][&filter][&orderby][&rowVersion]

	targetURL := "https://data.emergencyreporting.com/agencyapparatus/apparatus"

	headers := map[string]string{}

	var parsedResponse GetApparatusesResponse

	err := c.internalRequest(http.MethodGet, targetURL, options, headers, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the apparatuses: %v", err)
	}

	return &parsedResponse, nil
}
