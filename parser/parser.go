package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/morpheusxaut/eveslackkills/database"
	"github.com/morpheusxaut/eveslackkills/misc"
	"github.com/morpheusxaut/eveslackkills/models"
)

// Parser represents the parser used for retrieving kills from zKillboard and posting them to Slack
type Parser struct {
	Corporations []*models.Corporation

	scheduler *time.Ticker
	config    *misc.Configuration
	database  database.Connection
}

// SetupParser sets up a new parser with the given information
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

// Start starts the parsing operations and retrieves kills and losses for every tracked corporation regularly
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

// Update retrieves the latest kills and losses and posts them to Slack if required
func (parser *Parser) Update(corporation *models.Corporation) error {
	misc.Logger.Debugf("Running update for corporation #%d", corporation.EVECorporationID)

	kills, err := parser.FetchKills(corporation)
	if err != nil {
		return err
	}

	misc.Logger.Tracef("Fetched %d kills for corporation #%d", len(kills), corporation.EVECorporationID)

	for _, kill := range kills {
		misc.Logger.Tracef("Processing kill #%d (victim %q)", kill.KillID, kill.Victim.CharacterName)

		regionID, err := parser.database.QueryRegionID(kill.SolarSystemID)
		if err != nil {
			misc.Logger.Warnf("Failed to query region ID for solar system #%d", kill.SolarSystemID)
			continue
		}

		skip := false
		for _, solarSystem := range corporation.IgnoredSolarSystems {
			if solarSystem == regionID || solarSystem == kill.SolarSystemID {
				skip = true
				break
			}
		}

		if skip {
			misc.Logger.Debugf("Found solar system ID %d for solar system #%d on ignore list, skipping kill", regionID, kill.SolarSystemID)
			continue
		}

		misc.Logger.Tracef("Solar system ID %d for solar system #%d not found on ignore list (%v), posting kill", regionID, kill.SolarSystemID, corporation.IgnoredSolarSystems)

		err = parser.SendMessage(corporation, kill, true)
		if err != nil {
			misc.Logger.Warnf("Failed to send kill message: [%v]", err)
			continue
		}

		if kill.KillID > corporation.LastKillID {
			corporation.LastKillID = kill.KillID
		}

		misc.Logger.Tracef("Finished processing kill #%d (victim %q)", kill.KillID, kill.Victim.CharacterName)

		// Wait in order to abide to Slack's message limit
		time.Sleep(time.Second * 1)
	}

	losses, err := parser.FetchLosses(corporation)
	if err != nil {
		return err
	}

	misc.Logger.Tracef("Fetched %d losses for corporation #%d", len(losses), corporation.EVECorporationID)

	for _, loss := range losses {
		misc.Logger.Tracef("Processing loss #%d (victim %q)", loss.KillID, loss.Victim.CharacterName)

		regionID, err := parser.database.QueryRegionID(loss.SolarSystemID)
		if err != nil {
			misc.Logger.Warnf("Failed to query region ID for solar system #%d", loss.SolarSystemID)
			continue
		}

		skip := false
		for _, solarSystem := range corporation.IgnoredSolarSystems {
			if solarSystem == regionID || solarSystem == loss.SolarSystemID {
				skip = true
				break
			}
		}

		if skip {
			misc.Logger.Debugf("Found solar system ID %d for solar system #%d on ignore list, skipping loss", regionID, loss.SolarSystemID)
			continue
		}

		misc.Logger.Tracef("Solar system ID %d for solar system #%d not found on ignore list (%v), posting loss", regionID, loss.SolarSystemID, corporation.IgnoredSolarSystems)

		err = parser.SendMessage(corporation, loss, false)
		if err != nil {
			misc.Logger.Warnf("Failed to send loss message: [%v]", err)
			continue
		}

		if loss.KillID > corporation.LastLossID {
			corporation.LastLossID = loss.KillID
		}

		misc.Logger.Tracef("Finished processing loss #%d (victim %q)", loss.KillID, loss.Victim.CharacterName)

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

// SendMessage prepares a payload and sends a formatted kill/loss message to the Slack webhook
func (parser *Parser) SendMessage(corporation *models.Corporation, entry models.ZKillboardEntry, killEntry bool) error {
	var payload models.SlackPayload
	var kill models.SlackAttachment

	var killer models.ZKillboardAttacker
	var killerName string
	var killerCorpName string
	var killerShipName string
	var victimName string
	var victimShipName string

	var highestDamageDealer models.ZKillboardAttacker
	var highestDamageValue int64

	var damageTakenTitle string

	for _, attacker := range entry.Attackers {
		if attacker.FinalBlow == 1 {
			killer = attacker
			killerCorpName = attacker.CorporationName
		}
		if attacker.CharacterID == 0 && attacker.FactionID != 0 {
			misc.Logger.Debugf("Found attacker with character ID 0 and faction ID #%d for kill #%d", attacker.FactionID, entry.KillID)
			continue
		}
		if attacker.CharacterID != 0 && attacker.DamageDone > highestDamageValue {
			highestDamageDealer = attacker
			highestDamageValue = attacker.DamageDone
		}
	}

	shipName, err := parser.database.QueryShipName(killer.ShipTypeID)
	if err != nil {
		misc.Logger.Warnf("Failed to query ship type ID #%d for killer of killboard entry #%d", killer.ShipTypeID, entry.KillID)
		return err
	}

	killerShipName = shipName

	if killer.CharacterID == 0 {
		killerName = shipName
	} else {
		killerName = killer.CharacterName
	}

	shipName, err = parser.database.QueryShipName(entry.Victim.ShipTypeID)
	if err != nil {
		misc.Logger.Warnf("Failed to query ship type ID #%d for victim of killboard entry #%d", entry.Victim.ShipTypeID, entry.KillID)
		return err
	}

	victimShipName = shipName

	if entry.Victim.CharacterID == 0 {
		victimName = shipName
	} else {
		victimName = entry.Victim.CharacterName
	}

	solarSystemName, err := parser.database.QuerySolarSystemName(entry.SolarSystemID)
	if err != nil {
		misc.Logger.Warnf("Failed to query solar system name for ID #%d of killboard entry #%d", entry.SolarSystemID, entry.KillID)
		return err
	}

	var comment string

	if killEntry {
		comment = corporation.KillComment
		kill.Color = "good"
		damageTakenTitle = "Damage dealt"
	} else {
		comment = corporation.LossComment
		kill.Color = "danger"
		damageTakenTitle = "Damage taken"
	}

	killLink := fmt.Sprintf("https://zkillboard.com/kill/%d/", entry.KillID)

	comment = strings.Replace(comment, "{victimshipname}", victimShipName, -1)
	comment = strings.Replace(comment, "{victimname}", victimName, -1)
	comment = strings.Replace(comment, "{victimcorpname}", entry.Victim.CorporationName, -1)
	comment = strings.Replace(comment, "{killershipname}", killerShipName, -1)
	comment = strings.Replace(comment, "{killername}", killerName, -1)
	comment = strings.Replace(comment, "{killercorpname}", killerCorpName, -1)
	comment = strings.Replace(comment, "{killid}", strconv.FormatInt(entry.KillID, 10), -1)
	comment = strings.Replace(comment, "{killlink}", killLink, -1)

	kill.Fallback = comment
	kill.Title = comment
	kill.TitleLink = killLink
	kill.ThumbURL = fmt.Sprintf("https://imageserver.eveonline.com/render/%d_64.png", entry.Victim.ShipTypeID)

	kill.Fields = append(kill.Fields, models.SlackField{
		Title: damageTakenTitle,
		Value: humanize.Comma(entry.Victim.DamageTaken),
		Short: true,
	})

	kill.Fields = append(kill.Fields, models.SlackField{
		Title: "Pilots involved",
		Value: humanize.Comma(int64(len(entry.Attackers))),
		Short: true,
	})

	kill.Fields = append(kill.Fields, models.SlackField{
		Title: "ISK value",
		Value: fmt.Sprintf("%s ISK", humanize.Commaf(entry.Misc.TotalValue)),
		Short: true,
	})

	kill.Fields = append(kill.Fields, models.SlackField{
		Title: "Highest damage",
		Value: fmt.Sprintf("<https://zkillboard.com/character/%d|%s> (%s damage)", highestDamageDealer.CharacterID, highestDamageDealer.CharacterName, humanize.Comma(highestDamageValue)),
		Short: true,
	})

	kill.Fields = append(kill.Fields, models.SlackField{
		Title: "Solar system",
		Value: fmt.Sprintf("<https://zkillboard.com/system/%d|%s>", entry.SolarSystemID, solarSystemName),
		Short: true,
	})

	kill.Fields = append(kill.Fields, models.SlackField{
		Title: "Ship",
		Value: victimShipName,
		Short: true,
	})

	kill.Fields = append(kill.Fields, models.SlackField{
		Title: "Timestamp",
		Value: entry.KillTime,
		Short: true,
	})

	kill.Fields = append(kill.Fields, models.SlackField{
		Title: "Kill ID",
		Value: fmt.Sprintf("%d", entry.KillID),
		Short: true,
	})

	payload.Attachments = append(payload.Attachments, kill)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", parser.config.SlackWebhookURL, bytes.NewBuffer(jsonPayload))
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

// FetchKills retrieves and parses the latest kills from the zKillboard API
func (parser *Parser) FetchKills(corporation *models.Corporation) ([]models.ZKillboardEntry, error) {
	resp, err := http.Get(fmt.Sprintf("https://zkillboard.com/api/kills/corporationID/%d/afterKillID/%d", corporation.EVECorporationID, corporation.LastKillID))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var kills []models.ZKillboardEntry

	err = json.Unmarshal(response, &kills)
	if err != nil {
		return nil, err
	}

	sort.Sort(models.ByKillID(kills))

	return kills, nil
}

// FetchLosses retrieves and parses the latest losses from the zKillboard API
func (parser *Parser) FetchLosses(corporation *models.Corporation) ([]models.ZKillboardEntry, error) {
	resp, err := http.Get(fmt.Sprintf("https://zkillboard.com/api/losses/corporationID/%d/afterKillID/%d", corporation.EVECorporationID, corporation.LastLossID))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var losses []models.ZKillboardEntry

	err = json.Unmarshal(response, &losses)
	if err != nil {
		return nil, err
	}

	sort.Sort(models.ByKillID(losses))

	return losses, nil
}
