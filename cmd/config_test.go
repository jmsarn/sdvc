package cmd

import (
	"testing"

	"github.com/jmsarn/sdvc/utils"
	"github.com/stretchr/testify/assert"
)

func Test_ExecuteConfigCommand(t *testing.T) {
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
	cfg, err = updateConfig("core.remote", remoteName, cfg, false)
	if err != nil {
		t.Fatal(err)
	}

	sec, err := cfg.File.GetSection("core")
	if err != nil {
		t.Fatal(err)
	}

	k, err := sec.GetKey("remote")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, remoteName, k.String())
}
