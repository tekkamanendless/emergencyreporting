package emergencyreporting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// ErrorDuplicate represents a duplicate key error.
var ErrorDuplicate = fmt.Errorf("Duplicate")

type Client struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ClientID        string `json:"client_id"`
	ClientSecret    string `json:"client_secret"`
	Token           string `json:"token"`            // Required, but can be generated using the username, etc.
	SubscriptionKey string `json:"subscription_key"` // Required no matter what.
	client          http.Client
}

func (c *Client) GenerateToken() (*GenerateTokenResponse, error) {
	values := url.Values{
		"grant_type":    {"password"},
		"username":      {c.Username},
		"password":      {c.Password},
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
	}
	targetURL := "https://auth.emergencyreporting.com/Token.php"
	fmt.Printf("POST %s\n", targetURL)
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
		fmt.Printf("Body: %v\n", string(contents))
		return nil, fmt.Errorf("Bad status code: %d", response.StatusCode)
	}

	var parsedResponse GenerateTokenResponse
	err = json.Unmarshal(contents, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not parse JSON: %v", err)
	}

	return &parsedResponse, nil
}

func (c *Client) internalRequest(method string, targetURL string, options map[string]string, headers map[string]string, body []byte, targetPointer interface{}) error {
	queryParts := url.Values{}
	for key, value := range options {
		queryParts.Set(key, value)
	}
	if len(queryParts) > 0 {
		targetURL += "?" + queryParts.Encode()
	}

	fmt.Printf("%s %s\n", method, targetURL)
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
	fmt.Printf("%s %s %d %d\n", method, targetURL, response.StatusCode, len(contents))

	if response.StatusCode < 200 || response.StatusCode > 299 {
		var errorResponse ErrorResponse
		err = json.Unmarshal(contents, &errorResponse)
		if err != nil || errorResponse.Errors.Type == "" {
			fmt.Printf("Error: %s\n", string(contents))
			return fmt.Errorf("Bad status code: %d", response.StatusCode)
		}
		switch errorResponse.Errors.Type {
		case "Duplicate":
			return ErrorDuplicate
		default:
			return fmt.Errorf("Error type: %s (%s); status code: %d", errorResponse.Errors.Type, errorResponse.Errors.Message, response.StatusCode)
		}
	}

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
func (c *Client) GetIncidents(options map[string]string, deep bool) (*GetIncidentsResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/incidents[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/incidents"

	var parsedResponse GetIncidentsResponse

	err := c.internalRequest(http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the incidents: %v", err)
	}

	if deep {
		for _, incident := range parsedResponse.Incidents {
			exposuresResponse, err := c.GetExposures(incident.IncidentID, nil, deep)
			if err != nil {
				panic(err)
			}
			incident.Exposures = exposuresResponse.Exposures
		}
	}

	return &parsedResponse, nil
}

// PostIncident TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/postIncidents?
func (c *Client) PostIncident(incident Incident) (*PostIncidentResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/incidents[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/incidents"

	jsonInput, err := json.Marshal(incident)
	if err != nil {
		return nil, fmt.Errorf("Could not create JSON: %v", err)
	}
	fmt.Printf("JSON input: %s\n", string(jsonInput))

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

// GetExposures TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/IncidentsExposuresByIncidentIDGet?
func (c *Client) GetExposures(incidentID string, options map[string]string, deep bool) (*GetExposuresResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}/exposures[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/incidents/" + url.PathEscape(incidentID) + "/exposures"

	var parsedResponse GetExposuresResponse

	err := c.internalRequest(http.MethodGet, targetURL, options, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the exposures: %v", err)
	}

	if deep {
		for _, exposure := range parsedResponse.Exposures {
			exposureLocationResponse, err := c.GetExposureLocation(exposure.ExposureID)
			if err != nil {
				panic(err)
			}
			exposure.Location = exposureLocationResponse.Location

			getExposureApparatusesResponse, err := c.GetExposureApparatuses(exposure.ExposureID)
			if err != nil {
				return nil, fmt.Errorf("Could not get the exposure location: %v", err)
			}
			exposure.Apparatuses = getExposureApparatusesResponse.Apparatuses
		}
	}

	return &parsedResponse, nil
}

// GetExposure TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/IncidentsExposuresByIncidentIDAndExposureIDGet?
func (c *Client) GetExposure(incidentID string, exposureID string, deep bool) (*GetExposureResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}/exposures/{exposureID}

	targetURL := "https://data.emergencyreporting.com/agencyincidents/incidents/" + url.PathEscape(incidentID) + "/exposures/" + url.PathEscape(exposureID)

	var parsedResponse GetExposureResponse

	err := c.internalRequest(http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the exposure: %v", err)
	}

	if deep {
		getExposureLocationResponse, err := c.GetExposureLocation(exposureID)
		if err != nil {
			return nil, fmt.Errorf("Could not get the exposure location: %v", err)
		}
		parsedResponse.Exposure.Location = getExposureLocationResponse.Location

		getExposureApparatusesResponse, err := c.GetExposureApparatuses(exposureID)
		if err != nil {
			return nil, fmt.Errorf("Could not get the exposure location: %v", err)
		}
		parsedResponse.Exposure.Apparatuses = getExposureApparatusesResponse.Apparatuses
	}

	return &parsedResponse, nil
}

// PostExposure TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/IncidentsExposuresByIncidentIDPost?
func (c *Client) PostExposure(incidentID string, exposure Exposure) (*PostExposureResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/incidents/{incidentID}/exposures

	targetURL := "https://data.emergencyreporting.com/agencyincidents/incidents/" + url.PathEscape(incidentID) + "/exposures"

	jsonInput, err := json.Marshal(exposure)
	if err != nil {
		return nil, fmt.Errorf("Could not create JSON: %v", err)
	}
	fmt.Printf("JSON input: %s\n", string(jsonInput))

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var parsedResponse PostExposureResponse

	err = c.internalRequest(http.MethodPost, targetURL, nil, headers, jsonInput, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not create the incident: %v", err)
	}

	return &parsedResponse, nil
}

// GetExposureLocation TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresLocationByExposureIDGet?
func (c *Client) GetExposureLocation(exposureID string) (*GetExposureLocationResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/location[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/location"

	var parsedResponse GetExposureLocationResponse

	err := c.internalRequest(http.MethodGet, targetURL, nil, nil, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the exposure: %v", err)
	}

	return &parsedResponse, nil
}

// PutExposureLocation TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresLocationByExposureIDGet?
func (c *Client) PutExposureLocation(exposureID string, location ExposureLocation) (*PutExposureLocationResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/location[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/location"

	jsonInput, err := json.Marshal(location)
	if err != nil {
		return nil, fmt.Errorf("Could not create JSON: %v", err)
	}
	fmt.Printf("JSON input: %s\n", string(jsonInput))

	headers := map[string]string{
		"Content-Type": "application/json",
		"ETag":         location.RowVersion,
	}

	var parsedResponse PutExposureLocationResponse

	err = c.internalRequest(http.MethodPut, targetURL, nil, headers, jsonInput, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not put the exposure: %v", err)
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
		return nil, fmt.Errorf("Could not get the exposure: %v", err)
	}

	return &parsedResponse, nil
}

// PostExposureApparatus TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-incidents/operations/ExposuresApparatusesByExposureIDPost?
func (c *Client) PostExposureApparatus(exposureID string, apparatus ExposureApparatus) (*PostExposureApparatusResponse, error) {
	// https://data.emergencyreporting.com/agencyincidents/exposures/{exposureID}/apparatuses[?useAssociatedAgencyApparatusID]

	targetURL := "https://data.emergencyreporting.com/agencyincidents/exposures/" + url.PathEscape(exposureID) + "/apparatuses"
	options := map[string]string{
		"useAssociatedAgencyApparatusID": "1",
	}

	jsonInput, err := json.Marshal(apparatus)
	if err != nil {
		return nil, fmt.Errorf("Could not create JSON: %v", err)
	}
	fmt.Printf("JSON input: %s\n", string(jsonInput))

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

// GetUsers TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-users/operations/V1UsersGet?
func (c *Client) GetUsers(options map[string]string, deep bool) (*GetUsersResponse, error) {
	// https://data.emergencyreporting.com/agencyusers/users[?rowVersion][&limit][&offset][&filter][&orderby]

	targetURL := "https://data.emergencyreporting.com/agencyusers/users"

	headers := map[string]string{}

	var parsedResponse GetUsersResponse

	err := c.internalRequest(http.MethodGet, targetURL, options, headers, nil, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Could not get the users: %v", err)
	}

	if deep {
		for _, user := range parsedResponse.Users {
			getUserContactInfoResponse, err := c.GetUserContactInfo(user.UserID)
			if err != nil {
				return nil, fmt.Errorf("Could not get user contact info for user ID %s: %v", user.UserID, err)
			}
			user.ContactInfo = &getUserContactInfoResponse.ContactInfo
		}
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
		return nil, fmt.Errorf("Could not get the exposure: %v", err)
	}

	return &parsedResponse, nil
}

// PatchUser TODO
// See: https://developer.emergencyreporting.com/docs/services/agency-users/operations/V1UsersByUserIDPatch?
func (c *Client) PatchUser(userID string, rowVersion string, payload PatchUserRequest) (*PatchUserResponse, error) {
	// https://data.emergencyreporting.com/agencyusers/users/{userID}

	targetURL := "https://data.emergencyreporting.com/agencyusers/users/" + url.PathEscape(userID)

	jsonInput, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Could not create JSON: %v", err)
	}
	fmt.Printf("JSON input: %s\n", string(jsonInput))

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
