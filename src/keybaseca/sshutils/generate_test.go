package sshutils

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/atvenu/bot-sshca/src/shared"
	"github.com/stretchr/testify/require"
)

// Test generating a new SSH key
func TestGenerateNewSSHKey(t *testing.T) {
	filename := "/tmp/bot-sshca-integration-test-generate-key"
	os.Remove(filename)

	err := GenerateNewSSHKey(filename, false, false)
	require.NoError(t, err)

	err = GenerateNewSSHKey(filename, false, false)
	require.Errorf(t, err, "Refusing to overwrite existing key (try with FORCE_WRITE=true if you're sure): "+filename)

	err = GenerateNewSSHKey(filename, true, false)
	require.NoError(t, err)

	bytes, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	require.True(t, strings.Contains(string(bytes), "PRIVATE"))

	bytes, err = ioutil.ReadFile(shared.KeyPathToPubKey(filename))
	require.NoError(t, err)
	require.False(t, strings.Contains(string(bytes), "PRIVATE"))
	require.True(t, strings.HasPrefix(string(bytes), "ssh-ed25519") || strings.HasPrefix(string(bytes), "ecdsa-sha2-nistp256"))
}
