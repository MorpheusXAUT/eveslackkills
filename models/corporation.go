package models

// Corporation represents an EVE corporation to be tracked by the application
type Corporation struct {
	ID                  int64
	EVECorporationID    int64
	LastKillID          int64
	LastLossID          int64
	Name                string
	KillComment         string
	LossComment         string
	IgnoredSolarSystems []int64
}
