package parser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/morpheusxaut/eveslackkills/database"
	"github.com/morpheusxaut/eveslackkills/misc"
	"github.com/morpheusxaut/eveslackkills/models"
)

type Parser struct {
	Corporations []*models.Corporation

	scheduler *time.Ticker
	config    *misc.Configuration
	database  database.Connection
}

func SetupParser(conf *misc.Configuration, db database.Connection, interval time.Duration) (*Parser, error) {
	parser := &Parser{
		Corporations: make([]*models.Corporation, 0),
		scheduler:    time.NewTicker(interval),
		config:       conf,
		database:     db,
	}

	corporations, err := db.LoadAllCorporations()
	if err != nil {
		return nil, err
	}

	parser.Corporations = corporations

	return parser, nil
}

func (parser *Parser) Start() {
	for _, corporation := range parser.Corporations {
		err := parser.Update(corporation)
		if err != nil {
			misc.Logger.Errorf("Received error while updating corporation #%d: [%v]", corporation.EVECorporationID, err)
		}
	}

	for {
		select {
		case <-parser.scheduler.C:
			for _, corporation := range parser.Corporations {
				err := parser.Update(corporation)
				if err != nil {
					misc.Logger.Errorf("Received error while updating corporation #%d: [%v]", corporation.EVECorporationID, err)
				}
			}
		}
	}
}

func (parser *Parser) Update(corporation *models.Corporation) error {
	misc.Logger.Debugf("Running update for corporation #%d", corporation.EVECorporationID)

	kills, err := parser.FetchKills(corporation)
	if err != nil {
		return err
	}

	misc.Logger.Tracef("Fetched %d kills for corporation #%d", len(kills.Rows), corporation.EVECorporationID)

	for _, kill := range kills.Rows {
		regionID, err := parser.database.QueryRegionID(kill.SolarSystemID)
		if err != nil {
			misc.Logger.Warnf("Failed to query region ID for solar system #%d", kill.SolarSystemID)
			continue
		}

		skip := false
		for _, region := range corporation.IgnoredRegions {
			if region == regionID {
				skip = true
				break
			}
		}

		if skip {
			misc.Logger.Debugf("Found region ID %d for solar system #%d on ignore list, skipping kill", regionID, kill.SolarSystemID)
			continue
		}

		misc.Logger.Tracef("Region ID %d for solar system #%d not found on ignore list, posting kill", regionID, kill.SolarSystemID)

		shipName, err := parser.database.QueryShipName(kill.Victim.ShipTypeID)
		if err != nil {
			misc.Logger.Warnf("Failed to query ship name for ship type #%d", kill.Victim.ShipTypeID)
			continue
		}

		err = parser.SendMessage(corporation, kill.KillID, kill.Victim.CharacterName, shipName, true)
		if err != nil {
			misc.Logger.Warnf("Failed to send kill message: [%v]", err)
			continue
		}

		if kill.KillID > corporation.LastKillID {
			corporation.LastKillID = kill.KillID
		}

		// Wait in order to abide to Slack's message limit
		time.Sleep(time.Second * 1)
	}

	losses, err := parser.FetchLosses(corporation)
	if err != nil {
		return err
	}

	misc.Logger.Tracef("Fetched %d losses for corporation #%d", len(losses.Rows), corporation.EVECorporationID)

	for _, loss := range losses.Rows {
		regionID, err := parser.database.QueryRegionID(loss.SolarSystemID)
		if err != nil {
			misc.Logger.Warnf("Failed to query region ID for solar system #%d", loss.SolarSystemID)
			continue
		}

		skip := false
		for _, region := range corporation.IgnoredRegions {
			if region == regionID {
				skip = true
				break
			}
		}

		if skip {
			misc.Logger.Debugf("Found region ID %d for solar system #%d on ignore list, skipping loss", regionID, loss.SolarSystemID)
			continue
		}

		misc.Logger.Tracef("Region ID %d for solar system #%d not found on ignore list, posting loss", regionID, loss.SolarSystemID)

		shipName, err := parser.database.QueryShipName(loss.Victim.ShipTypeID)
		if err != nil {
			misc.Logger.Warnf("Failed to query ship name for ship type #%d", loss.Victim.ShipTypeID)
			continue
		}

		err = parser.SendMessage(corporation, loss.KillID, loss.Victim.CharacterName, shipName, true)
		if err != nil {
			misc.Logger.Warnf("Failed to send loss message: [%v]", err)
			continue
		}

		if loss.KillID > corporation.LastLossID {
			corporation.LastLossID = loss.KillID
		}

		// Wait in order to abide to Slack's message limit
		time.Sleep(time.Second * 1)
	}

	_, err = parser.database.SaveCorporation(corporation)
	if err != nil {
		return err
	}

	misc.Logger.Debugf("Finished update for corporation #%d", corporation.EVECorporationID)

	return nil
}

func (parser *Parser) SendMessage(corporation *models.Corporation, killID int64, playerName string, shipName string, kill bool) error {
	var comment string

	if kill {
		comment = corporation.KillComment
	} else {
		comment = corporation.LossComment
	}

	killLink := fmt.Sprintf("https://zkillboard.com/kill/%d/", killID)

	comment = strings.Replace(comment, "{corpname}", corporation.Name, -1)
	comment = strings.Replace(comment, "{shipname}", shipName, -1)
	comment = strings.Replace(comment, "{playername}", playerName, -1)
	comment = strings.Replace(comment, "{killid}", strconv.FormatInt(killID, 10), -1)
	comment = strings.Replace(comment, "{killlink}", killLink, -1)

	req, err := http.NewRequest("POST", parser.config.SlackWebhookURL, bytes.NewBufferString(comment))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("Failed to send message: %s (status code: %d)", string(respBody), resp.StatusCode)
	}

	return nil
}

func (parser *Parser) FetchKills(corporation *models.Corporation) (models.ZKillboardEntry, error) {
	var zKillboardEntry models.ZKillboardEntry

	resp, err := http.Get(fmt.Sprintf("https://zkillboard.com/api/kills/corporationID/%d/afterKillID/%d/xml/", corporation.EVECorporationID, corporation.LastKillID))
	if err != nil {
		return zKillboardEntry, err
	}

	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&zKillboardEntry)
	if err != nil && !strings.EqualFold(err.Error(), "EOF") {
		return zKillboardEntry, err
	}

	sort.Sort(models.ByKillID(zKillboardEntry.Rows))

	return zKillboardEntry, nil
}

func (parser *Parser) FetchLosses(corporation *models.Corporation) (models.ZKillboardEntry, error) {
	var zKillboardEntry models.ZKillboardEntry

	resp, err := http.Get(fmt.Sprintf("https://zkillboard.com/api/losses/corporationID/%d/afterKillID/%d/xml/", corporation.EVECorporationID, corporation.LastLossID))
	if err != nil {
		return zKillboardEntry, err
	}

	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&zKillboardEntry)
	if err != nil && !strings.EqualFold(err.Error(), "EOF") {
		return zKillboardEntry, err
	}

	sort.Sort(models.ByKillID(zKillboardEntry.Rows))

	return zKillboardEntry, nil
}
