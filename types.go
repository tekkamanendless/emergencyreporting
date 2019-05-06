package emergencyreporting

type ErrorResponse struct {
	Errors []struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"errors"`
}
type ErrorResponseBuggy struct {
	Errors struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"errors"`
}

type GenerateTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

type Station struct {
	RowNum                  string  `json:"rowNum"`
	StationID               string  `json:"stationID"`
	StationNumber           string  `json:"stationNumber"`
	StationName             string  `json:"stationName"`
	CreateDate              string  `json:"createDate"`
	StreetNumber            string  `json:"streetNumber"`
	StreetPrefix            *string `json:"streetPrefix"`
	Address                 string  `json:"address"`
	StreetType              string  `json:"streetType"`
	StreetSuffix            *string `json:"streetSuffix"`
	City                    string  `json:"city"`
	State                   string  `json:"state"`
	ZipCode                 string  `json:"zip"`
	Latitude                string  `json:"latitude"`
	Longitude               string  `json:"longitude"`
	Manned                  string  `json:"manned"`
	Phone                   string  `json:"phone"`
	PhoneType               *string `json:"phoneType"`
	SecondaryPhone          *string `json:"secondaryPhone"`
	SecondaryPhoneType      *string `json:"secondaryPhoneType"`
	ZoneID                  *string `json:"zoneID"`
	ZoneCode                *string `json:"zoneCode"`
	RowVersion              string  `json:"rowVersion"`
	Nemsis3LocationType     *string `json:"nemsis3LocationType"`
	NationalGridCoordinates *string `json:"nationalGridCoordinates"`
	Country                 string  `json:"country"`
	FreeFormAddress         *string `json:"freeFormAddress"`
	AddressEntryFormat      string  `json:"addressEntryFormat"`
}

type GetStationsResponse struct {
	TotalRows string     `json:"totalRows"`
	Stations  []*Station `json:"stations"`
}

type Incident struct {
	StationID             string      `json:"stationID"`
	State                 string      `json:"state"`
	IncidentDateTime      string      `json:"incidentDateTime"`
	FDID                  string      `json:"fdid"`
	IncidentNumber        string      `json:"incidentNumber"`
	PartnerIncidentNumber string      `json:"partnerIncidentNumber"`
	DispatchRunNumber     string      `json:"dispatchRunNumber"`
	IsComplete            string      `json:"isComplete"`
	IsReviewed            string      `json:"isReviewed"`
	IncidentID            string      `json:"incidentID,omitempty"` // Not used for creating incidents.
	RowVersion            string      `json:"rowVersion,omitempty"` // Not used for creating incidents.
	Exposures             []*Exposure `json:"-"`
}

type GetIncidentsResponse struct {
	Incidents []*Incident `json:"incidents"`
}

type PostIncidentResponse struct {
	IncidentID string `json:"incidentID"`
}

type Exposure struct {
	ShiftsOrPlatoon                string               `json:"shiftsOrPlatoon"`
	IncidentType                   string               `json:"incidentType"`
	AssignedToUserID               string               `json:"assignedToUserID"`
	AidGivenOrReceived             string               `json:"aidGivenOrReceived"`
	HazmatReleased                 string               `json:"hazmatReleased"`
	PrimaryActionTaken             string               `json:"primaryActionTaken"`
	SecondaryActionTaken           string               `json:"secondaryActionTaken"`
	ThirdActionTaken               string               `json:"thirdActionTaken"`
	CompletedByUserID              string               `json:"completedByUserID"`
	ReviewedByUserID               string               `json:"reviewedByUserID"`
	CompletedDateTime              string               `json:"completedDateTime"`
	ReviewedDateTime               string               `json:"reviewedDateTime"`
	PSAPDateTime                   string               `json:"psapDateTime"`
	DispatchNotifiedDateTime       string               `json:"dispatchNotifiedDateTime"`
	InitialResponderDateTime       string               `json:"initialResponderDateTime"`
	HasPropertyLoss                string               `json:"hasPropertyLoss"`
	PropertyLossAmount             string               `json:"propertyLossAmount"`
	HasContentLoss                 string               `json:"hasContentLoss"`
	ContentLossAmount              string               `json:"contentLossAmount"`
	HasPreIncidentPropertyValue    string               `json:"hasPreIncidentPropertyValue"`
	PreIncidentPropertyValueAmount string               `json:"preIncidentPropertyValueAmount"`
	HasPreIncidentContentsValue    string               `json:"hasPreIncidentContentsValue"`
	PreIncidentContentsValueAmount string               `json:"preIncidentContentsValueAmount"`
	CompaintReportedByDispatch     string               `json:"complaintReportedByDispatch"`
	ExposureID                     string               `json:"exposureID,omitempty"`
	RowVersion                     string               `json:"rowVersion,omitempty"`
	Location                       *ExposureLocation    `json:"-"`
	Apparatuses                    []*ExposureApparatus `json:"-"`
	Narratives                     []*ExposureNarrative `json:"-"`
}

type GetExposuresResponse struct {
	Exposures []*Exposure `json:"exposures"`
}

type GetExposureResponse struct {
	Exposure *Exposure `json:"exposure"`
}

type PostExposureResponse struct {
	ExposureID string `json:"exposureID"`
}

type ExposureLocation struct {
	LocationType                 string  `json:"locationType"` // 1: Street Address
	MilePostNumber               string  `json:"milePostNumber"`
	StreetPrefix                 string  `json:"streetPrefix"`
	StreetName                   string  `json:"streetName"`
	StreetType                   string  `json:"streetType"`
	StreetSuffix                 string  `json:"streetSuffix"`
	AptOrSuiteNumber             string  `json:"aptOrSuiteNumber"`
	City                         string  `json:"city"`
	CityCode                     string  `json:"cityCode"`
	State                        string  `json:"state"`
	ZipCode                      string  `json:"zipCode"`
	CountyCode                   string  `json:"countyCode"`
	Latitude                     string  `json:"latitude"`
	Longitude                    string  `json:"longitude"`
	CrossStreetOrDirections      string  `json:"crossStreetOrDirections"`
	ZoneID                       *string `json:"zoneID"`
	PopulationDensity            string  `json:"populationDensity"`
	PropertyUse                  string  `json:"propertyUse"` // 3-digit code
	NemsisPropertyClassification string  `json:"nemsisPropertyClassification"`
	ExposureID                   string  `json:"exposureID,omitempty"`
	RowVersion                   string  `json:"rowVersion,omitempty"`
}

type GetExposureLocationResponse struct {
	Location *ExposureLocation `json:"exposureLocation"`
}

type PutExposureLocationResponse struct {
	RowVersion string `json:"rowVersion"`
}

type ExposureApparatus struct {
	ApparatusID                     string  `json:"apparatusID"`
	AlarmDateTime                   string  `json:"alarmDateTime"`
	EnrouteDateTime                 *string `json:"enrouteDateTime"`
	ArrivedDateTime                 *string `json:"arrivedDateTime"`
	InjuryOrOnsetDateTime           *string `json:"injuryOrOnsetDateTime"`
	InQuartersDateTime              *string `json:"inQuartersDateTime"`
	CallCompletedDateTime           *string `json:"callCompletedDateTime"`
	DispatchToSceneMileage          *string `json:"dispatchToSceneMileage"`
	ResponseModeToScene             string  `json:"responseModeToScene"`
	DispatchDepartmentLocationID    *string `json:"dispatchDepartmentLocationID"`
	IncidentID                      string  `json:"incidentID"`
	ExposureID                      string  `json:"exposureID"`
	TransferOfPatientCareDateTime   *string `json:"transferOfPatientCareDateTime"`
	DispatchNationalGridCoordinates string  `json:"dispatchNationalGridCoordinates"`
	WasCancelled                    string  `json:"wasCancelled"`
	ResponseModeNemsis3             string  `json:"responseModeNemsis3"`
	DispatchAcknowledgedDateTime    *string `json:"dispatchAcknowledgedDateTime"`
	AtDestinationDateTime           *string `json:"atDestinationDateTime"`
	CancelledDateTime               *string `json:"cancelledDateTime"`
	ClearedSceneDateTime            *string `json:"clearedSceneDateTime"`
	ArrivedAtLandingZoneDateTime    *string `json:"arrivedAtLandingZoneDateTime"`
	ClearedDestinationDateTime      *string `json:"clearedDestinationDateTime"`
	AgencyApparatusID               string  `json:"agencyApparatusID"`
	DepartmentApparatusID           string  `json:"departmentApparatusID"`
	DispatchDateTime                string  `json:"dispatchDateTime"`
	ArrivedAtPatientDateTime        *string `json:"arrivedAtPatientDateTime"`
	DispatchLatitude                *string `json:"dispatchLatitude"`
	RowVersion                      string  `json:"rowVersion"`
	ApparatusTypeID                 string  `json:"apparatusTypeID"`
	ApparatusUseID                  string  `json:"apparatusUseID"`
	InServiceDateTime               *string `json:"inServiceDateTime"`
	DispatchLongitude               *string `json:"dispatchLongitude"`
	DispatchZoneID                  *string `json:"dispatchZoneID"`
}

type GetExposureApparatusesResponse struct {
	Apparatuses []*ExposureApparatus `json:"exposureApparatuses"`
}

type PostExposureApparatusResponse map[string]interface{}

type ExposureNarrative map[string]interface{}

type GetExposureNarrativesResponse struct {
	Narratives []*ExposureNarrative `json:"exposureNarrative"`
}

type CrewMember struct {
	UserID         string `json:"userID"`
	ApparatusID    string `json:"apparatusID"`
	ExposureID     string `json:"exposureID"`
	ExposureUserID string `json:"exposureUserID"`
	RowVersion     string `json:"rowVersion"`
}

type GetExposureMembersResponse struct {
	CrewMembers []*CrewMember `json:"crewMembers"`
}

type CrewMemberRole struct {
	ExposureUserRoleID string `json:"exposureUserRoleID"`
	ExposureID         string `json:"exposureID"`
	NFIRSCode          string `json:"nfirsCode"`
	RowVersion         string `json:"rowVersion"`
}

type GetExposureMemberRolesResponse struct {
	Roles []*CrewMemberRole `json:"roles"`
}

// Note: "rowNum" (string) is the 1-index of the entry; might be tacked on to all array responses?

type User struct {
	RowNum                   string           `json:"rowNum,omitempty"`
	AgencyPersonnelID        *string          `json:"agencyPersonnelID"`
	Email                    *string          `json:"email"`
	RoleName                 string           `json:"roleName"`
	UserID                   string           `json:"userID"`
	Title                    *string          `json:"title"`
	FullName                 string           `json:"fullName"`
	Login                    string           `json:"login"`
	Archive                  string           `json:"Archive"`
	PrimaryEmail             string           `json:"primaryEmail"`
	CertificationStatus      string           `json:"certificationStatus"`
	RoleID                   string           `json:"roleID"`
	DefaultEventPaygradeName *string          `json:"defaultEventPaygradeName"`
	DefaultEventPaygradeRate *string          `json:"defaultEventPaygradeRate"`
	Station                  *string          `json:"station"`
	Shift                    *string          `json:"shift"`
	RowVersion               string           `json:"rowVersion"`
	ContactInfo              *UserContactInfo `json:"-"`
}

type GetUsersResponse struct {
	Users []*User `json:"users"`
}

type GetUserResponse struct {
	User *User `json:"user"`
}

// TODO: MAKE THIS THE CORRECT TYPE
type UserContactInfo map[string]interface{}

type GetUserContactInfoResponse struct {
	ContactInfo UserContactInfo `json:"contactInfo"`
}

type PatchOperation struct {
	Operation string `json:"op"`
	Path      string `json:"path"`
	Value     string `json:"value"`
}

type PatchUserRequest []PatchOperation

type PatchUserResponse struct {
	RowVersion string `json:"rowVersion"`
}

type Apparatus struct {
	DepartmentApparatusID         string  `json:"departmentApparatusID"`
	ApparatusID                   string  `json:"apparatusID"`
	YearOfManufacture             string  `json:"yearOfManufacture"`
	Model                         string  `json:"model"`
	Engine                        string  `json:"engine"`
	TankVolume                    string  `json:"tankVolume"`
	PumpManufacturer              string  `json:"pumpManufacturer"`
	Notes                         string  `json:"notes"`
	ApparatusStationID            string  `json:"apparatusStationID"`
	DateInService                 string  `json:"dateInService"`
	ApparatusType                 string  `json:"apparatusType"`
	ReplaceDate                   string  `json:"replaceDate"`
	PrimaryUse                    string  `json:"primaryUse"`
	PrimaryUseName                string  `json:"primaryUseName"`
	StationNumber                 string  `json:"stationNumber"`
	StationName                   string  `json:"stationName"`
	VehicleNumber                 string  `json:"vehicleNumber"`
	VIN                           string  `json:"vinNumber"`
	LicencePlateNumber            string  `json:"licensePlateNumber"`
	DefaultPrimaryRoleOfUnit      string  `json:"defaultPrimaryRoleOfUnit"`
	DefaultPrimaryRoleOfUnitName  string  `json:"defaultPrimaryRoleOfUnitName"`
	DefaultServiceLevelOfUnit     string  `json:"defaultServiceLevelOfUnit"`
	DefaultServiceLevelOfUnitName string  `json:"defaultServiceLevelOfUnitName"`
	DepartmentApparatusName       string  `json:"departmentApparatusName"`
	VehicleInitialCost            string  `json:"vehicleInitialCost"`
	NemesisVehicleType            string  `json:"nemsisVehicleType"`
	NemesisVehicleTypeName        *string `json:"nemsisVehicleTypeName"`
	Archive                       string  `json:"archive"`
	EmsUnitCallSign               string  `json:"emsUnitCallSign"`
	Nemesis3VehicleType           string  `json:"nemsis3VehicleType"`
	Nemesis3VehicleTypeName       *string `json:"nemsis3VehicleTypeName"`
	ApparatusOwnership            string  `json:"apparatusOwnership"`
	Nemesis3TransportMethod       string  `json:"nemsis3TransportMethod"`
	Nemesis3TransportMethodName   string  `json:"nemsis3TransportMethodName"`
	InService                     string  `json:"inService"`
	NFPACompliance                string  `json:"nfpaCompliance"`
	RecurrenceTypeID              string  `json:"recurrenceTypeID"`
	RowVersion                    string  `json:"rowVersion"`
	ApparatusTypeName             string  `json:"apparatusTypeName"`
	Manufacturer                  string  `json:"manufacturer"`
}

type GetApparatusesResponse struct {
	Apparatuses []*Apparatus `json:"apparatus"`
}
