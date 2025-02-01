package cmd

import (
	"fmt"
	"testing"

	"github.com/jmsarn/sdvc/utils"
	"github.com/stretchr/testify/assert"
)

func Test_ExecuteRemoteAddCommand(t *testing.T) {
	cleanUpConfig()
	defer cleanUpConfig()

	err := initializeProject("s3://test-bucket", false)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := utils.ReadConfig(false)
	if err != nil {
		t.Fatal(err)
	}

	remoteName := "test-remote"
	remoteUrl := "s3://test-remote-bucket/foo"
	cfg, err = addRemote(remoteName, remoteUrl, cfg)
	if err != nil {
		t.Fatal(err)
	}

	sec, err := cfg.File.GetSection(fmt.Sprintf(`remote "%s"`, remoteName))
	if err != nil {
		t.Fatal(err)
	}

	k, err := sec.GetKey("url")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, remoteUrl, k.String())
}
