package models

type Corporation struct {
	ID               int64
	EVECorporationID int64
	LastKillID       int64
	LastLossID       int64
	Name             string
	KillComment      string
	LossComment      string
	IgnoredRegions   []int64
}
