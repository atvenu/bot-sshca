package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// Test generating a new SSH key
func TestGetTeams_basicParsing(t *testing.T) {
	var conf = EnvConfig{}
	var teams = conf.GetTeams()
	require.Empty(t, teams)
}
func TestGetTeams_oneTeam(t *testing.T) {
	os.Setenv("TEAMS", "team1")
	var conf = EnvConfig{}
	var teams = conf.GetTeams()
	require.NotEmpty(t, teams)
	require.Equal(t, teams, []string{"team1"})
}
func TestGetTeams_twoTeam(t *testing.T) {
	os.Setenv("TEAMS", "team2, team3")
	var conf = EnvConfig{}
	var teams = conf.GetTeams()
	require.NotEmpty(t, teams)
	require.Equal(t, teams, []string{"team2", "team3"})

}
