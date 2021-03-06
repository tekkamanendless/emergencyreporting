package emergencyreporting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Error values.
//
// Remember to test for these using `errors.Is`, since they may be wrapped
// coming out of the client.
var (
	// ErrorDuplicate represents a duplicate key error.
	ErrorDuplicate = fmt.Errorf("Duplicate")
	// ErrorNotFound represents a 404 "not found" error.
	ErrorNotFound = fmt.Errorf("NotFound")
)

// Logger is the basic logger for this package.
type Logger interface {
	Printf(format string, v ...interface{})
}

// DefaultLogger is the default logger that uses the "log" package.
type DefaultLogger struct{}

// Printf is the print function for this logger.
func (DefaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// NullLogger is a logger that doesn't log anything.
type NullLogger struct{}

// Printf is the print function for this logger.
func (NullLogger) Printf(format string, v ...interface{}) {}

// Client is the client interface for communicating with the Emergency Reporting API.
type Client struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ClientID        string `json:"client_id"`
	ClientSecret    string `json:"client_secret"`
	AccountID       string `json:"account_id"`       // (New password authentication) This is the user's account ID.
	UserID          string `json:"user_id"`          // (New password authentication) This is the user's ID.
	TenantHost      string `json:"tenant_host"`      // (New password authentication) If present, use this instead of the default value.
	TenantSegment   string `json:"tenant_segment"`   // (New password authentication) If present, use this instead of the default value.
	Token           string `json:"token"`            // Required, but can be generated using the username, etc.
	Host            string `json:"host"`             // If set, this will be used instead of "https://data.emergencyreporting.com".  If the protocol is not specified, "https://" is assumed.
	SubscriptionKey string `json:"subscription_key"` // Required no matter what.

	Logger Logger `json:"-"` // This is the Logger instance to use.  If empty, then the default one will be used.

	client http.Client
}

// init makes sure that everything is initialized.
func (c *Client) init() {
	if c.Logger == nil {
		c.Logger = &DefaultLogger{}
	}
}

// SetTimeout sets the timeout for any given request.
// By default, this uses Go's default request timeout, which is fairly large.
//
// This may be more convenient than using `context.WithTimeout` because this will apply
// a *relative* timeout to all future requests as opposed to an *absolute* timeout on
// a shared context.
func (c *Client) SetTimeout(timeout time.Duration) {
	c.init()

	c.client.Timeout = timeout
}

// GenerateToken generates a new token.
func (c *Client) GenerateToken(ctx context.Context) (*GenerateTokenResponse, error) {
	c.init()

	goLiveDate, _ := time.Parse("2006-01-02", "2020-12-06")
	if !time.Now().After(goLiveDate) {
		response, err := c.GenerateTokenLegacy(ctx)
		if err == nil {
			return response, nil
		}
		if c.AccountID == "" || c.UserID == "" {
			return nil, err
		}
	}

	response, err := c.GenerateToken2020(ctx)
	if err != nil {
		return nil, err
	}

	var returnValue GenerateTokenResponse
	returnValue.AccessToken = response.AccessToken
	returnValue.TokenType = response.TokenType
	expiresIn, _ := strconv.ParseInt(response.ExpiresIn, 10, 64)
	returnValue.ExpiresIn = int(expiresIn)

	return &returnValue, nil
}

// GenerateTokenLegacy generates a token using the legacy workflow.
// Deprecated; this will no longer work.
func (c *Client) GenerateTokenLegacy(ctx context.Context) (*GenerateTokenResponse, error) {
	c.init()

	targetURL := "https://auth.emergencyreporting.com/Token.php"
	values := url.Values{
		"grant_type":    {"password"},
		"username":      {c.Username},
		"password":      {c.Password},
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
	}

	c.Logger.Printf("POST %s\n", targetURL)

	request, err := http.NewRequest(http.MethodPost, targetURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("could create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	request = request.WithContext(ctx)

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("could not post form: %w", err)
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		c.Logger.Printf("Body: %v\n", string(contents))
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode)
	}

	// DEBUG:
	//c.Logger.Printf("%s\n", contents)
	// :GUBED

	var parsedResponse GenerateTokenResponse
	err = json.Unmarshal(contents, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not parse JSON: %w", err)
	}

	return &parsedResponse, nil
}

// GenerateToken2020 generates a token using the late-2020 method.
func (c *Client) GenerateToken2020(ctx context.Context) (*GenerateTokenResponseV2, error) {
	c.init()

	tenantHost := "login.emergencyreporting.com"
	tenantSegment := "login.emergencyreporting.com"

	if c.TenantHost != "" {
		tenantHost = c.TenantHost
	}
	if c.TenantSegment != "" {
		tenantSegment = c.TenantSegment
	}

	targetURL := "https://" + tenantHost + "/" + tenantSegment + "/B2C_1A_PasswordGrant/oauth2/v2.0/token"

	values := url.Values{
		"grant_type":    {"password"},
		"username":      {c.Username},
		"password":      {c.Password},
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
		"scope":         {"https://" + tenantSegment + "/secure/full_access"},
		"response_type": {"token"},
		"er_aid":        {c.AccountID},
		"er_uid":        {c.UserID},
	}

	c.Logger.Printf("POST %s\n", targetURL)

	request, err := http.NewRequest(http.MethodPost, targetURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("could create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	request = request.WithContext(ctx)

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("could not post form: %w", err)
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		c.Logger.Printf("Body: %v\n", string(contents))
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode)
	}

	// DEBUG:
	//c.Logger.Printf("%s\n", contents)
	// :GUBED

	var parsedResponse GenerateTokenResponseV2
	err = json.Unmarshal(contents, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not parse JSON: %w", err)
	}

	return &parsedResponse, nil
}

// RawOperation performs a raw HTTP request.
func (c *Client) RawOperation(ctx context.Context, method string, targetURL string, options map[string]string, headers map[string]string, body []byte) (json.RawMessage, error) {
	var response json.RawMessage
	err := c.internalRequest(ctx, method, targetURL, options, headers, body, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// internalRequest makes an request to the Emergency Reporting API.
//
// To set a timeout, pass in a context created with `context.WithTimeout`.
func (c *Client) internalRequest(ctx context.Context, method string, targetURL string, options map[string]string, headers map[string]string, body []byte, targetPointer interface{}) error {
	c.init()

	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		host := "data.emergencyreporting.com"
		if c.Host != "" {
			host = c.Host
		}
		if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
			host = "https://" + host
		}
		targetURL = strings.TrimRight(host, "/") + "/" + strings.TrimLeft(targetURL, "/")
	}

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
		return fmt.Errorf("could not make request: %w", err)
	}
	request.Header.Set("Authorization", c.Token)
	request.Header.Set("Ocp-Apim-Subscription-Key", c.SubscriptionKey)
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	// Set the context for the request.
	request = request.WithContext(ctx)

	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("could not perform operation: %w", err)
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("could not read body: %w", err)
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
			return fmt.Errorf("bad status code: %d", response.StatusCode)
		}
		switch errorType {
		case "Duplicate":
			return ErrorDuplicate
		default:
			return fmt.Errorf("error type: %s (%s); status code: %d", errorType, errorMessage, response.StatusCode)
		}
	}
	// DEBUG:
	//c.Logger.Printf("%s\n", contents)
	// :GUBED

	if targetPointer != nil {
		err = json.Unmarshal(contents, targetPointer)
		if err != nil {
			return fmt.Errorf("could not parse JSON: %w", err)
		}
	}

	return nil
}

// GetStations TODO
// See: https://developer.emergencyreporting.com/docs/services/stations/operations/get-stations?
func (c *Client) GetStations(ctx context.Context, options map[string]string) (*GetStationsResponse, error) {
	// https://data.emergencyreporting.com/agencystations/stations[?rowVersion][&changesSince][&limit][&offset][&showArchived][&filter]

	targetURL := "/agencystations/stations"

	var parsedResponse GetStationsResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the stations: %w", err)
	}

	return &parsedResponse, nil
}

// GetIncident TODO
// See: https://developer.emergencyreporting.com/api-details#api=agency-incidents&operation=getIncident
func (c *Client) GetIncident(ctx context.Context, incidentID string) (*GetIncidentResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}

	targetURL := "/agencyincidents/incidents/" + url.PathEscape(incidentID)

	var parsedResponse GetIncidentResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the incident: %w", err)
	}

	return &parsedResponse, nil
}

// GetIncidents TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/getIncidents?
func (c *Client) GetIncidents(ctx context.Context, options map[string]string) (*GetIncidentsResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/incidents[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "/agencyincidents/incidents"

	var parsedResponse GetIncidentsResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the incidents: %w", err)
	}

	return &parsedResponse, nil
}

// PostIncident TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/postIncidents?
func (c *Client) PostIncident(ctx context.Context, incident Incident) (*PostIncidentResponse, error) {
	c.init()

	// https://data.emergencyreporting.com/agencyincidents/incidents[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "/agencyincidents/incidents"

	jsonInput, err := json.Marshal(incident)
	if err != nil {
		return nil, fmt.Errorf("could not create JSON: %w", err)
	}
	c.Logger.Printf("JSON input: %s\n", string(jsonInput))

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var parsedResponse PostIncidentResponse

	err = c.internalRequest(ctx, http.MethodPost, targetURL, nil, headers, jsonInput, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not create the incident: %w", err)
	}

	return &parsedResponse, nil
}

// DeleteIncident TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/deleteIncident?
func (c *Client) DeleteIncident(ctx context.Context, incidentID string) error {
	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}

	targetURL := "/agencyincidents/incidents/" + url.PathEscape(incidentID)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	err := c.internalRequest(ctx, http.MethodDelete, targetURL, nil, headers, nil, nil)
	if err != nil {
		return fmt.Errorf("could not delete the incident: %w", err)
	}

	return nil
}

// GetIncidentExposures TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/IncidentsExposuresByIncidentIDGet?
func (c *Client) GetIncidentExposures(ctx context.Context, incidentID string, options map[string]string) (*GetExposuresResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}/exposures[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "/agencyincidents/incidents/" + url.PathEscape(incidentID) + "/exposures"

	var parsedResponse GetExposuresResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the exposures: %w", err)
	}

	return &parsedResponse, nil
}

// GetIncidentExposure TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/IncidentsExposuresByIncidentIDAndExposureIDGet?
func (c *Client) GetIncidentExposure(ctx context.Context, incidentID string, exposureID string) (*GetExposureResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}/exposures/{exposureID}

	targetURL := "/agencyincidents/incidents/" + url.PathEscape(incidentID) + "/exposures/" + url.PathEscape(exposureID)

	var parsedResponse GetExposureResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the exposure: %w", err)
	}

	return &parsedResponse, nil
}

// PostIncidentExposure TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/IncidentsExposuresByIncidentIDPost?
func (c *Client) PostIncidentExposure(ctx context.Context, incidentID string, exposure Exposure) (*PostExposureResponse, error) {
	c.init()

	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}/exposures

	targetURL := "/agencyincidents/incidents/" + url.PathEscape(incidentID) + "/exposures"

	jsonInput, err := json.Marshal(exposure)
	if err != nil {
		return nil, fmt.Errorf("could not create JSON: %w", err)
	}
	c.Logger.Printf("JSON input: %s\n", string(jsonInput))

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var parsedResponse PostExposureResponse

	err = c.internalRequest(ctx, http.MethodPost, targetURL, nil, headers, jsonInput, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not create the exposure: %w", err)
	}

	return &parsedResponse, nil
}

// DeleteIncidentExposure TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/IncidentsExposuresByIncidentIDAndExposureIDDelete?
func (c *Client) DeleteIncidentExposure(ctx context.Context, incidentID string, exposureID string) error {
	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}/exposures/{exposureID}

	targetURL := "/agencyincidents/incidents/" + url.PathEscape(incidentID) + "/exposures/" + url.PathEscape(exposureID)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	err := c.internalRequest(ctx, http.MethodDelete, targetURL, nil, headers, nil, nil)
	if err != nil {
		return fmt.Errorf("could not delete the exposure: %w", err)
	}

	return nil
}

// GetExposures TODO
// See: https://developer.emergencyreporting.com/api-details#api=agency-incidents&operation=IncidentsExposuresGet
func (c *Client) GetExposures(ctx context.Context, options map[string]string) (*GetExposuresResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/incidents/exposures[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "/agencyincidents/incidents/exposures"

	var parsedResponse GetExposuresResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the exposures: %w", err)
	}

	return &parsedResponse, nil
}

// PatchIncidentExposure TODO
// See: https://developer.emergencyreporting.com/api-details#api=agency-incidents&operation=IncidentsExposuresByIncidentIDAndExposureIDPatch
func (c *Client) PatchIncidentExposure(ctx context.Context, incidentID string, exposureID string, rowVersion string, payload PatchExposureRequest) (*PatchExposureResponse, error) {
	c.init()

	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}/exposures/{exposureID}

	targetURL := "/agencyincidents/incidents/" + url.PathEscape(incidentID) + "/exposures/" + url.PathEscape(exposureID)

	jsonInput, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("could not create JSON: %w", err)
	}
	c.Logger.Printf("JSON input: %s\n", string(jsonInput))

	headers := map[string]string{
		"Content-Type": "application/json",
		"ETag":         rowVersion,
	}

	var parsedResponse PatchExposureResponse

	err = c.internalRequest(ctx, http.MethodPatch, targetURL, nil, headers, jsonInput, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not patch the exposure: %w", err)
	}

	return &parsedResponse, nil
}

// GetExposureLocation TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresLocationByExposureIDGet?
func (c *Client) GetExposureLocation(ctx context.Context, exposureID string) (*GetExposureLocationResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/location[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/location"

	var parsedResponse GetExposureLocationResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the exposure location: %w", err)
	}

	return &parsedResponse, nil
}

// PutExposureLocation TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresLocationByExposureIDGet?
func (c *Client) PutExposureLocation(ctx context.Context, exposureID string, location ExposureLocation) (*PutExposureLocationResponse, error) {
	c.init()

	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/location[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/location"

	jsonInput, err := json.Marshal(location)
	if err != nil {
		return nil, fmt.Errorf("could not create JSON: %w", err)
	}
	c.Logger.Printf("JSON input: %s\n", string(jsonInput))

	headers := map[string]string{
		"Content-Type": "application/json",
		"ETag":         location.RowVersion,
	}

	var parsedResponse PutExposureLocationResponse

	err = c.internalRequest(ctx, http.MethodPut, targetURL, nil, headers, jsonInput, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not put the exposure location: %w", err)
	}

	return &parsedResponse, nil
}

// GetExposureFire TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresFireByExposureIDGet?
func (c *Client) GetExposureFire(ctx context.Context, exposureID string) (*GetExposureFireResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/fire[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/fire"

	var parsedResponse GetExposureFireResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		if err == ErrorNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("could not get the exposure fire: %w", err)
	}

	return &parsedResponse, nil
}

// GetExposureApparatuses TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresApparatusesByExposureIDGet?
func (c *Client) GetExposureApparatuses(ctx context.Context, exposureID string) (*GetExposureApparatusesResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/apparatuses[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/apparatuses"

	var parsedResponse GetExposureApparatusesResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the exposure apparatuses: %w", err)
	}

	return &parsedResponse, nil
}

// PostExposureApparatus TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresApparatusesByExposureIDPost?
func (c *Client) PostExposureApparatus(ctx context.Context, exposureID string, apparatus ExposureApparatus) (*PostExposureApparatusResponse, error) {
	c.init()

	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/apparatuses[?useAssociatedAgencyApparatusID]

	targetURL := "/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/apparatuses"
	options := map[string]string{
		"useAssociatedAgencyApparatusID": "1",
	}

	jsonInput, err := json.Marshal(apparatus)
	if err != nil {
		return nil, fmt.Errorf("could not create JSON: %w", err)
	}
	c.Logger.Printf("JSON input: %s\n", string(jsonInput))

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var parsedResponse PostExposureApparatusResponse

	err = c.internalRequest(ctx, http.MethodPost, targetURL, options, headers, jsonInput, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not create the apparatus: %w", err)
	}

	return &parsedResponse, nil
}

// GetExposureMember TODO
// See: https://developer.emergencyreporting.com/api-details#api=agency-incidents&operation=ExposuresCrewmembersByExposureIDAndExposureUserIDGet
func (c *Client) GetExposureMember(ctx context.Context, exposureID string, exposureUserID string) (*GetExposureMemberResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/crewmembers/{exposureUserID}

	targetURL := "/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/crewmembers/" + url.PathEscape(exposureUserID)

	var parsedResponse GetExposureMemberResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the exposure members: %w", err)
	}

	return &parsedResponse, nil
}

// GetExposureMembers TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresCrewmembersByExposureIDGet?
func (c *Client) GetExposureMembers(ctx context.Context, exposureID string, options map[string]string) (*GetExposureMembersResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/crewmembers[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/crewmembers"

	var parsedResponse GetExposureMembersResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the exposure members: %w", err)
	}

	return &parsedResponse, nil
}

// GetExposureMemberRoles TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/CrewmembersRolesByExposureUserIDGet?
func (c *Client) GetExposureMemberRoles(ctx context.Context, exposureUserID string, options map[string]string) (*GetExposureMemberRolesResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/crewmembers/{exposureUserID}/roles[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "/agencyincidents/crewmembers/" + url.PathEscape(exposureUserID) + "/roles"

	var parsedResponse GetExposureMemberRolesResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the exposure member roles: %w", err)
	}

	return &parsedResponse, nil
}

// GetUsers TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-users/operations/V1UsersGet?
func (c *Client) GetUsers(ctx context.Context, options map[string]string) (*GetUsersResponse, error) {
	// https://data.emergencyreporting.com/agencyusers/users[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "/agencyusers/users"

	headers := map[string]string{}

	var parsedResponse GetUsersResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, options, headers, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the users: %w", err)
	}

	return &parsedResponse, nil
}

// GetUser TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-users/operations/V1UsersByUserIDGet?
func (c *Client) GetUser(ctx context.Context, userID string) (*GetUserResponse, error) {
	// https://data.emergencyreporting.com/agencyusers/users/{userID}

	targetURL := "/agencyusers/users/" + url.PathEscape(userID)

	options := map[string]string{}
	headers := map[string]string{}

	var parsedResponse GetUserResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, options, headers, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the user: %w", err)
	}

	return &parsedResponse, nil
}

// GetUserContactInfo TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-users/operations/V1UsersContactinfoByUserIDGet?
func (c *Client) GetUserContactInfo(ctx context.Context, userID string) (*GetUserContactInfoResponse, error) {
	// https://data.emergencyreporting.com/agencyusers/users/{userID}/contactinfo

	targetURL := "/agencyusers/users/" + url.PathEscape(userID) + "/contactinfo"

	var parsedResponse GetUserContactInfoResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the user contact info: %w", err)
	}

	return &parsedResponse, nil
}

// PatchUser TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-users/operations/V1UsersByUserIDPatch?
func (c *Client) PatchUser(ctx context.Context, userID string, rowVersion string, payload PatchUserRequest) (*PatchUserResponse, error) {
	c.init()

	// https://data.emergencyreporting.com/agencyusers/users/{userID}

	targetURL := "/agencyusers/users/" + url.PathEscape(userID)

	jsonInput, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("could not create JSON: %w", err)
	}
	c.Logger.Printf("JSON input: %s\n", string(jsonInput))

	headers := map[string]string{
		"Content-Type": "application/json",
		"ETag":         rowVersion,
	}

	var parsedResponse PatchUserResponse

	err = c.internalRequest(ctx, http.MethodPatch, targetURL, nil, headers, jsonInput, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not patch the user: %w", err)
	}

	return &parsedResponse, nil
}

// GetApparatus TODO
// See: https://developer.emergencyreporting.com/api-details#api=agency-apparatus&operation=ApparatusByDepartmentApparatusIDGet
func (c *Client) GetApparatus(ctx context.Context, apparatusID string) (*GetApparatusResponse, error) {
	// https://data.emergencyreporting.com/agencyapparatus/apparatus/{departmentApparatusID}

	targetURL := "/agencyapparatus/apparatus/" + url.PathEscape(apparatusID)

	var parsedResponse GetApparatusResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the apparatuses: %w", err)
	}

	return &parsedResponse, nil
}

// GetApparatuses TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-apparatus/operations/ApparatusGet?
func (c *Client) GetApparatuses(ctx context.Context, options map[string]string) (*GetApparatusesResponse, error) {
	// https://data.emergencyreporting.com/agencyapparatus/apparatus[?limit][&offset][&filter][&orderby][&rowVersion]

	targetURL := "/agencyapparatus/apparatus"

	var parsedResponse GetApparatusesResponse

	err := c.internalRequest(ctx, http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("could not get the apparatuses: %w", err)
	}

	return &parsedResponse, nil
}
