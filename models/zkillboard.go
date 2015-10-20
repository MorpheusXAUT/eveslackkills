package models

// ZKillboardEntry represents a kill or loss entry received via the zKillboard API
type ZKillboardEntry struct {
	KillID        int64                   `json:"killID"`
	SolarSystemID int64                   `json:"solarSystemID"`
	MoonID        int64                   `json:"moonID"`
	KillTime      string                  `json:"killTime"`
	Victim        ZKillboardVictim        `json:"victim"`
	Attackers     []ZKillboardAttacker    `json:"attackers"`
	Items         []ZKillboardItem        `json:"items"`
	Misc          ZKillboardMiscellaneous `json:"zkb"`
}

// ZKillboardVictim represents the victim of a kill as received via the zKillboard API
type ZKillboardVictim struct {
	CharacterID     int64  `json:"characterID"`
	CharacterName   string `json:"characterName"`
	CorporationID   int64  `json:"corporationID"`
	CorporationName string `json:"corporationName"`
	AllianceID      int64  `json:"allianceID"`
	AllianceName    string `json:"allianceName"`
	FactionID       int64  `json:"factionID"`
	FactionName     string `json:"factionName"`
	ShipTypeID      int64  `json:"shipTypeID"`
	DamageTaken     int64  `json:"damageTaken"`
}

// ZKillboardAttacker represents an attacker of a kill as received via the zKillboard API
type ZKillboardAttacker struct {
	CharacterID     int64   `json:"characterID"`
	CharacterName   string  `json:"characterName"`
	CorporationID   int64   `json:"corporationID"`
	CorporationName string  `json:"corporationName"`
	AllianceID      int64   `json:"allianceID"`
	AllianceName    string  `json:"allianceName"`
	FactionID       int64   `json:"factionID"`
	FactionName     string  `json:"factionName"`
	ShipTypeID      int64   `json:"shipTypeID"`
	WeaponTypeID    int64   `json:"weaponTypeID"`
	DamageDone      int64   `json:"damageDone"`
	FinalBlow       int64   `json:"finalBlow"`
	SecurityStatus  float64 `json:"securityStatus"`
}

// ZKillboardItem represents an item of a kill as received via the zKillboard API
type ZKillboardItem struct {
	TypeID            int64 `json:"typeID"`
	Flag              int64 `json:"flag"`
	QuantityDropped   int64 `json:"qntDropped"`
	QuantityDestroyed int64 `json:"qntDestroyed"`
	Singleton         int64 `json:"singleton"`
}

// ZKillboardMiscellaneous represents miscellaneous information about a kill as received via the zKillboard API
type ZKillboardMiscellaneous struct {
	Hash       string  `json:"hash"`
	TotalValue float64 `json:"totalValue"`
	Points     int64   `json:"points"`
}

// ByKillID represents an array of kills or losses, used for sorting by kill ID
type ByKillID []ZKillboardEntry

// Len returns the length of the array of kills or losses to sort
func (k ByKillID) Len() int {
	return len(k)
}

// Swap swaps two entries in the array of kills or losses to sort
func (k ByKillID) Swap(i, j int) {
	k[i], k[j] = k[j], k[i]
}

// Less is used for sorting the array of kills or losses by comparing kill IDs
func (k ByKillID) Less(i, j int) bool {
	return k[i].KillID < k[j].KillID
}
