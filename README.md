eveslackkills
=======

[![Travis CI](https://travis-ci.org/MorpheusXAUT/eveslackkills.svg?branch=master)](https://travis-ci.org/MorpheusXAUT/eveslackkills) [![GoDoc](https://godoc.org/github.com/MorpheusXAUT/eveslackkills?status.svg)](https://godoc.org/github.com/MorpheusXAUT/eveslackkills)

eveslackkills fetches kill- and loss-mails for a given corporation and posts them to a specified killboard-channel in Slack. eveslackkills uses the EVE CREST to retrieve information about ships and solar systems, so no manual SDE data updates have to be performed anymore.

Requirements
---------

- Slack
- MySQL server
- [Go 1.2 or newer](https://golang.org/dl/) to compile application (latest recommended)


Instructions
---------

- Create a new "Incoming WebHooks" integration in Slack (https://YOURNAME.slack.com/services/new/incoming-webhook)
  - Fill in channel/username as required
  - Copy generated web hook URL for later use config
- Clone the GitHub project and build the Go application
  - Use "go get -u ./..." to fetch all requirements
  - Build the app using "go build -v ./..."
  - You will only need the generated executable, no additional files as cloned from the repository
- Create the MySQL database required for the application
  - Use the script provided in database/mysql/eveslackkills_create.sql to create the database/tables required
  - Set up a username/password for the application to access the database
- Create a config file "config.cfg" using JSON format

```{
	"DatabaseType": 1,
	"DatabaseHost": "HOSTNAME:PORT",
	"DatabaseSchema": "MYSQLDATABASE",
	"DatabaseUser": "MYSQLUSER",
	"DatabasePassword": "MYSQLPASSWORD",
	"DebugLevel": 1,
	"SlackWebhookURL": "SLACKHOOKURL"
}```

- Run the application and use a monitoring service such as supervisord to restart it automatically if required

Copyright
---------

All information and data regarding [EVE Online](http://www.eveonline.com/) is provided by [CCP](http://www.ccpgames.com/en/home) according to this notice:

EVE Online and the EVE logo are the registered trademarks of CCP hf. All rights are reserved worldwide. All other trademarks are the property of their respective owners. EVE Online, the EVE logo, EVE and all associated logos and designs are the intellectual property of CCP hf. All artwork, screenshots, characters, vehicles, storylines, world facts or other recognizable features of the intellectual property relating to these trademarks are likewise the intellectual property of CCP hf. CCP hf. has granted permission to eveslackkills to use EVE Online and all associated logos and designs for promotional and information purposes on its website but does not endorse, and is not in any way affiliated with, eveslackkills. CCP is in no way responsible for the content on or functioning of this website, nor can it be liable for any damage arising from the use of this website.

License
-------

eveslackkills code and documentation are copyright 2015 by [MorpheusXAUT](https://github.com/MorpheusXAUT). The application is released under the GNU [GPLv3 license](https://www.gnu.org/licenses/gpl.html).
