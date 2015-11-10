package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/morpheusxaut/eveslackkills/misc"
)

// CRESTRoot represents the root resource retrieved from the EVE CREST
type CRESTRoot struct {
	Endpoint      CRESTHref `json:"crestEndpoint"`
	ItemTypes     CRESTHref `json:"itemTypes"`
	ServerVersion string    `json:"serverVersion"`
	ServerName    string    `json:"serverName"`
}

// CRESTHref represents a link resource as provided by the EVE CREST
type CRESTHref struct {
	Href string `json:"href"`
}

// CRESTItemType represents information about an item type as provided by the EVE CREST
type CRESTItemType struct {
	Name   string  `json:"name"`
	Volume float64 `json:"volume"`
}

// CRESTSolarSystem represents information about a solar system as provided by the EVE CREST
type CRESTSolarSystem struct {
	Name           string    `json:"name"`
	SecurityStatus float64   `json:"securityStatus"`
	Constellation  CRESTHref `json:"constellation"`
}

// CRESTConstellation represents information about a constellation as provided by the EVE CREST
type CRESTConstellation struct {
	Name    string      `json:"name"`
	Region  CRESTHref   `json:"region"`
	Systems []CRESTHref `json:"systems"`
}

// CRESTRegion represents information about a region as provided by the EVE CREST
type CRESTRegion struct {
	Name           string      `json:"name"`
	Constellations []CRESTHref `json:"constellations"`
}

// CRESTLocationInfo stores information about a solarsystem, constellation and region
type CRESTLocationInfo struct {
	SolarSystemID       string
	SolarSystemName     string
	SolarSystemSecurity float64
	ConstellationID     string
	ConstellationName   string
	RegionID            string
	RegionName          string
}

// CRESTClient is used for retrieving data from the EVE CREST and caching it locally
type CRESTClient struct {
	crestRoot     string
	client        *http.Client
	serverVersion string
	itemTypes     map[int64]*CRESTItemType
	locationInfo  map[int64]*CRESTLocationInfo
}

// NewCRESTClient creates a new CRESTClient with the given root URL
func NewCRESTClient(root string) *CRESTClient {
	c := &CRESTClient{
		crestRoot:     root,
		client:        &http.Client{},
		serverVersion: "",
		itemTypes:     make(map[int64]*CRESTItemType),
		locationInfo:  make(map[int64]*CRESTLocationInfo),
	}

	return c
}

// FetchEndpoint retrieves the given CREST endpoint and returns the read data
func (c *CRESTClient) FetchEndpoint(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.ccp.eve.Api-v3+json")
	req.Header.Set("User-Agent", "eveslackkills github.com/morpheusxaut/eveslackkills")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received non-OK HTTP status code %s (%d)", resp.Status, resp.StatusCode)
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// CheckServerVersion compares the stored and provided server version, invalidating cached data if a change has been detected
func (c *CRESTClient) CheckServerVersion(version string) {
	if !strings.EqualFold(c.serverVersion, version) {
		misc.Logger.Tracef("CREST server version changed (was %q, is %q), deleting cached data", c.serverVersion, version)

		c.itemTypes = make(map[int64]*CRESTItemType)
		c.locationInfo = make(map[int64]*CRESTLocationInfo)
		c.serverVersion = version
	}
}

// FetchRoot retrieves the root CREST endpoint
func (c *CRESTClient) FetchRoot() (*CRESTRoot, error) {
	response, err := c.FetchEndpoint(c.crestRoot)
	if err != nil {
		return nil, err
	}

	var root *CRESTRoot

	err = json.Unmarshal(response, &root)
	if err != nil {
		return nil, err
	}

	return root, nil
}

// FetchItemType retrieves the item type information for the given ID
func (c *CRESTClient) FetchItemType(typeID int64) (*CRESTItemType, error) {
	item, ok := c.itemTypes[typeID]
	if ok {
		misc.Logger.Tracef("Found item type for type ID #%d in cache", typeID)
		return item, nil
	}

	misc.Logger.Tracef("Querying CREST for item type #%d", typeID)

	response, err := c.FetchEndpoint(fmt.Sprintf("%s/types/%d/", c.crestRoot, typeID))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, &item)
	if err != nil {
		return nil, err
	}

	c.itemTypes[typeID] = item

	return item, nil
}

// FetchLocationInfo retrieves all available location info for the given solar system ID
func (c *CRESTClient) FetchLocationInfo(systemID int64) (*CRESTLocationInfo, error) {
	info, ok := c.locationInfo[systemID]
	if ok {
		misc.Logger.Tracef("Found location info for solar system #%d in cache", systemID)
		return info, nil
	}

	misc.Logger.Tracef("Querying CREST for location info for solar system #%d", systemID)

	system, err := c.FetchSolarSystem(systemID)
	if err != nil {
		return nil, err
	}

	constellation, err := c.FetchConstellation(system.Constellation.Href)
	if err != nil {
		return nil, err
	}

	region, err := c.FetchRegion(constellation.Region.Href)
	if err != nil {
		return nil, err
	}

	idx := strings.LastIndex(system.Constellation.Href[:len(system.Constellation.Href)-1], "/")
	if idx < 0 {
		return nil, fmt.Errorf("Invalid constellation URL")
	}

	constellationID := system.Constellation.Href[idx+1 : len(system.Constellation.Href)-1]

	idx = strings.LastIndex(constellation.Region.Href[:len(constellation.Region.Href)-1], "/")
	if idx < 0 {
		return nil, fmt.Errorf("Invalid region URL")
	}

	regionID := constellation.Region.Href[idx+1 : len(constellation.Region.Href)-1]

	info = &CRESTLocationInfo{
		SolarSystemID:       fmt.Sprintf("%d", systemID),
		SolarSystemName:     system.Name,
		SolarSystemSecurity: system.SecurityStatus,
		ConstellationID:     constellationID,
		ConstellationName:   constellation.Name,
		RegionID:            regionID,
		RegionName:          region.Name,
	}

	c.locationInfo[systemID] = info

	return info, nil
}

// FetchSolarSystem retrieves the solar system information for the given ID
func (c *CRESTClient) FetchSolarSystem(systemID int64) (*CRESTSolarSystem, error) {
	response, err := c.FetchEndpoint(fmt.Sprintf("%s/solarsystems/%d/", c.crestRoot, systemID))
	if err != nil {
		return nil, err
	}

	var system *CRESTSolarSystem

	err = json.Unmarshal(response, &system)
	if err != nil {
		return nil, err
	}

	return system, nil
}

// FetchConstellation retrieves the constellation information for the given ID
func (c *CRESTClient) FetchConstellation(url string) (*CRESTConstellation, error) {
	response, err := c.FetchEndpoint(url)
	if err != nil {
		return nil, err
	}

	var constellation *CRESTConstellation

	err = json.Unmarshal(response, &constellation)
	if err != nil {
		return nil, err
	}

	return constellation, nil
}

// FetchRegion retrieves the region information for the given ID
func (c *CRESTClient) FetchRegion(url string) (*CRESTRegion, error) {
	response, err := c.FetchEndpoint(url)
	if err != nil {
		return nil, err
	}

	var region *CRESTRegion

	err = json.Unmarshal(response, &region)
	if err != nil {
		return nil, err
	}

	return region, nil
}
