package database

import (
	"fmt"

	"github.com/morpheusxaut/eveslackkills/database/mysql"
	"github.com/morpheusxaut/eveslackkills/misc"
	"github.com/morpheusxaut/eveslackkills/models"
)

// Connection provides an interface for communicating with a database backend in order to retrieve and persist the needed information
type Connection interface {
	// Connect tries to establish a connection to the database backend, returning an error if the attempt failed
	Connect() error

	// RawQuery performs a raw database query and returns a map of interfaces containing the retrieve data. An error is returned if the query failed
	RawQuery(query string, v ...interface{}) ([]map[string]interface{}, error)

	// LoadAllCorporations retrieves all corporations from the database, returning an error if the query failed
	LoadAllCorporations() ([]*models.Corporation, error)

	// LoadCorporation retrieves the corporation with the given ID from the database, returning an error if the query failed
	LoadCorporation(corporationID int64) (*models.Corporation, error)

	// LoadAllIgnoredRegionsForCorporation retrieves all ignored regions associated with the given corporation from the database, returning an error if the query failed
	LoadAllIgnoredRegionsForCorporation(corporationID int64) ([]int64, error)

	// QueryShipName looks up the given ship type ID and returns the ship's name, returning an error if the query failed
	QueryShipName(shipTypeID int64) (string, error)
	// QueryShipName looks up the given solar system ID and returns the associated region ID, returning an error if the query failed
	QueryRegionID(solarSystemID int64) (int64, error)

	// SaveCorporation saves a corporation to the database, returning the updated model or an error if the query failed
	SaveCorporation(corporation *models.Corporation) (*models.Corporation, error)
}

// SetupDatabase parses the database type set in the configuration and returns an appropriate database implementation or an error if the type is unknown
func SetupDatabase(conf *misc.Configuration) (Connection, error) {
	var database Connection

	switch Type(conf.DatabaseType) {
	case TypeMySQL:
		database = &mysql.DatabaseConnection{
			Config: conf,
		}
		break
	default:
		return nil, fmt.Errorf("Unknown type #%d", conf.DatabaseType)
	}

	return database, nil
}
