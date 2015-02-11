package models

import (
	"encoding/xml"
)

type ZKillboardEntry struct {
	XMLName     xml.Name             `xml:"eveapi"`
	CurrentTime string               `xml:"currentTime"`
	Rows        []ZKillboardEntryRow `xml:"result>rowset>row"`
	CachedUntil string               `xml:"cachedUntil"`
}

type ZKillboardEntryRow struct {
	KillID        int64            `xml:"killID,attr"`
	SolarSystemID int64            `xml:"solarSystemID,attr"`
	KillTime      string           `xml:"killTime,attr"`
	MoonID        int64            `xml:"moonID,attr"`
	Victim        ZKillboardVictim `xml:"victim"`
}

type ZKillboardVictim struct {
	CharacterID     int64  `xml:"characterID,attr"`
	CharacterName   string `xml:"characterName,attr"`
	CorporationID   int64  `xml:"corporationID,attr"`
	CorporationName string `xml:"corporationName,attr"`
	AllianceID      int64  `xml:"allianceID,attr"`
	AllianceName    string `xml:"allianceName,attr"`
	FactionID       int64  `xml:"factionID,attr"`
	FactionName     string `xml:"factionName,attr"`
	DamageTaken     int64  `xml:"damageTaken,attr"`
	ShipTypeID      int64  `xml:"shipTypeID,attr"`
}

type ByKillID []ZKillboardEntryRow

func (k ByKillID) Len() int {
	return len(k)
}

func (k ByKillID) Swap(i, j int) {
	k[i], k[j] = k[j], k[i]
}

func (k ByKillID) Less(i, j int) bool {
	return k[i].KillID < k[j].KillID
}
