package kssh

import (
	"encoding/json"
	"fmt"
	"github.com/atvenu/bot-sshca/src/shared"
	"io/ioutil"
	"os"
	"strings"
)

// A ConfigFile that is provided by the keybaseca server process and lives in kbfs. It is used to share configuration
// information about how kssh should communicate with the keybaseca chatbot.
type ConfigFile struct {
	TeamName    string `json:"teamname"`
	ChannelName string `json:"channelname"`
	BotName     string `json:"botname"`
}

// LoadConfigs loads client configs from KBFS. Returns a (listOfConfigFiles, listOfBotNames, err)
// Both lists are deduplicated based on ConfigFile.BotName. Runs the KBFS operations in parallel
// to speed up loading configs.
func LoadConfigs(requestName string) ([]ConfigFile, []string, error) {
	config := ConfigFile{}
	filename := fmt.Sprintf("/keybase/team/%s/%s", "atvenu.ssh."+requestName, shared.ConfigFilename)
	exists, err := GetKBFSOperationsStruct().KBFSFileExists(filename)
	if err != nil {
		exists = false
	}
	if exists {
		conf, err := LoadConfig(filename)
		if err != nil {
			fmt.Errorf("Error loading config",err)
		} else {
			config = conf
		}
	}
	var configs []ConfigFile
	var botnames []string
	if (ConfigFile{}) != config {
		configs = append(configs, config)
		botnames = append(botnames, config.BotName)
	}

	return configs, botnames, nil
}

// Load a kssh client config file from the given filename
func LoadConfig(kbfsFilename string) (ConfigFile, error) {
	var cf ConfigFile
	if !strings.HasPrefix(kbfsFilename, "/keybase/") {
		return cf, fmt.Errorf("cannot load a kssh config from outside of KBFS")
	}
	bytes, err := GetKBFSOperationsStruct().KBFSRead(kbfsFilename)
	if err != nil {
		return cf, fmt.Errorf("found a config file at %s that could not be read: %v", kbfsFilename, err)
	}
	err = json.Unmarshal(bytes, &cf)
	if err != nil {
		return cf, fmt.Errorf("failed to parse config file at %s: %v", kbfsFilename, err)
	}
	if cf.TeamName == "" || cf.BotName == "" {
		return cf, fmt.Errorf("found a config file at %s that is missing data: %s", kbfsFilename, string(bytes))
	}
	return cf, err
}

// A LocalConfigFile is a file that lives on the FS of the computer running kssh. By default (and for most users), this
// file is not used.
//
// If a user of kssh is in in multiple teams that are running the CA bot they can configure a default bot to communicate
// with. Note that we store the team in here (even though it wasn't specified by the user) so that we can avoid doing
// a call to `LoadConfigs` if a default is set (since `LoadConfigs can be very slow if the user is in a large number of teams).
// This is controlled via `kssh --set-default-bot foo`.
//
// If a user of kssh wishes to configure a default ssh user to use (see README.md for a description of why this may
// be useful) this is also stored in the local config file. This is controlled via `kssh --set-default-user foo`.
type LocalConfigFile struct {
	DefaultBotName string `json:"default_bot"`
	DefaultBotTeam string `json:"default_team"`
	DefaultSSHUser string `json:"default_ssh_user"`
	KeybaseBinPath string `json:"keybase_binary"`
}

func GetKeybaseBinaryPath() string {
	lcf, err := getCurrentConfig()
	if err != nil {
		return "keybase"
	}

	if lcf.KeybaseBinPath != "" {
		return lcf.KeybaseBinPath
	}
	return "keybase"
}

func GetKBFSOperationsStruct() *shared.KBFSOperation {
	return &shared.KBFSOperation{KeybaseBinaryPath: GetKeybaseBinaryPath()}
}

// Where to store the local config file. Just stash it in ~/.ssh rather than making a ~/.kssh folder
var localConfigFileLocation = shared.ExpandPathWithTilde("~/.ssh/kssh-config.json")

// Get the default SSH user to use for kssh connections. Empty if no user is configured.
func GetDefaultSSHUser() (string, error) {
	lcf, err := getCurrentConfig()
	if err != nil {
		return "", err
	}

	return lcf.DefaultSSHUser, nil
}

// Set the default SSH user to use for kssh connections.
func SetKeybaseBinaryPath(path string) error {
	lcf, err := getCurrentConfig()
	if err != nil {
		return err
	}

	lcf.KeybaseBinPath = path
	return writeConfig(lcf)
}

// Set the default SSH user to use for kssh connections.
func SetDefaultSSHUser(username string) error {
	if strings.ContainsAny(username, " \t\n\r'\"") {
		return fmt.Errorf("invalid username: %s", username)
	}

	lcf, err := getCurrentConfig()
	if err != nil {
		return err
	}

	lcf.DefaultSSHUser = username
	return writeConfig(lcf)
}

// Write the given config file to disk
func writeConfig(lcf LocalConfigFile) error {
	bytes, err := json.Marshal(&lcf)
	if err != nil {
		return fmt.Errorf("failed to marshal json into config file: %v", err)
	}

	// Create ~/.ssh/ if it does not yet exist
	err = MakeDotSSH()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(localConfigFileLocation, bytes, 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	return nil
}

// Get the current kssh config file
func getCurrentConfig() (lcf LocalConfigFile, err error) {
	if _, err := os.Stat(localConfigFileLocation); os.IsNotExist(err) {
		return lcf, nil
	}
	bytes, err := ioutil.ReadFile(localConfigFileLocation)
	if err != nil {
		return lcf, fmt.Errorf("failed to read local config file: %v", err)
	}
	err = json.Unmarshal(bytes, &lcf)
	if err != nil {
		return lcf, fmt.Errorf("failed to parse local config file: %v", err)
	}
	return lcf, nil
}

// Set the default keybaseca bot to communicate with.
func SetDefaultBot(botname string) error {
	teamname := ""
	var err error
	if botname != "" {
		// Get the team associated with it and cache that too in order to avoid looking it up everytime
		teamname, err = GetTeamFromBot(botname)
		if err != nil {
			return err
		}
	}

	lcf, err := getCurrentConfig()
	if err != nil {
		return err
	}
	lcf.DefaultBotName = botname
	lcf.DefaultBotTeam = teamname

	return writeConfig(lcf)
}

// Get the default bot and team for kssh
func GetDefaultBotAndTeam() (string, string, error) {
	lcf, err := getCurrentConfig()
	if err != nil {
		return "", "", err
	}
	return lcf.DefaultBotName, lcf.DefaultBotTeam, nil
}

// Get the teamname associated with the given botname
func GetTeamFromBot(botname string) (string, error) {
	configs, _, err := LoadConfigs("nothing-to-match")
	if err != nil {
		return "", err
	}
	for _, config := range configs {
		if config.BotName == botname {
			return config.TeamName, nil
		}
	}
	return "", fmt.Errorf("did not find a client config file matching botname=%s (is the CA bot running and are you in the correct teams?)", botname)
}
