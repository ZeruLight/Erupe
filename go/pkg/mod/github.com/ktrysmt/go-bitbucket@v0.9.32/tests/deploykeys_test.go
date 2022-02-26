package tests

import (
	"fmt"
	"os"
	"testing"

	_ "github.com/k0kubun/pp"

	"github.com/ktrysmt/go-bitbucket"
)

func TestDeployKey(t *testing.T) {
	user := os.Getenv("BITBUCKET_TEST_USERNAME")
	pass := os.Getenv("BITBUCKET_TEST_PASSWORD")
	owner := os.Getenv("BITBUCKET_TEST_OWNER")
	repo := os.Getenv("BITBUCKET_TEST_REPOSLUG")

	if user == "" {
		t.Error("BITBUCKET_TEST_USERNAME is empty.")
	}
	if pass == "" {
		t.Error("BITBUCKET_TEST_PASSWORD is empty.")
	}
	if owner == "" {
		t.Error("BITBUCKET_TEST_OWNER is empty.")
	}
	if repo == "" {
		t.Error("BITBUCKET_TEST_REPOSLUG is empty.")
	}

	c := bitbucket.NewBasicAuth(user, pass)

	var deployKeyResourceId int

	label := "go-bb-test"
	key := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDAK/b1cHHDr/TEV1JGQl+WjCwStKG6Bhrv0rFpEsYlyTBm1fzN0VOJJYn4ZOPCPJwqse6fGbXntEs+BbXiptR+++HycVgl65TMR0b5ul5AgwrVdZdT7qjCOCgaSV74/9xlHDK8oqgGnfA7ZoBBU+qpVyaloSjBdJfLtPY/xqj4yHnXKYzrtn/uFc4Kp9Tb7PUg9Io3qohSTGJGVHnsVblq/rToJG7L5xIo0OxK0SJSQ5vuId93ZuFZrCNMXj8JDHZeSEtjJzpRCBEXHxpOPhAcbm4MzULgkFHhAVgp4JbkrT99/wpvZ7r9AdkTg7HGqL3rlaDrEcWfL7Lu6TnhBdq5"

	t.Run("create", func(t *testing.T) {
		opt := &bitbucket.DeployKeyOptions{
			Owner:    owner,
			RepoSlug: repo,
			Label:    label,
			Key:      key,
		}

		deployKey, err := c.Repositories.DeployKeys.Create(opt)
		if err != nil {
			t.Error(err)
		}

		if deployKey == nil {
			t.Error("The Deploy Key could not be created.")
		}

		if deployKey.Label != label {
			t.Error("The Deploy Key `label` attribute does not match the expected value.")
		}
		if deployKey.Key != key {
			t.Error("The Deploy Key `key` attribute does not match the expected value.")
		}

		deployKeyResourceId = deployKey.Id
	})

	t.Run("get", func(t *testing.T) {
		opt := &bitbucket.DeployKeyOptions{
			Owner:    owner,
			RepoSlug: repo,
			Id:       deployKeyResourceId,
		}
		deployKey, err := c.Repositories.DeployKeys.Get(opt)
		if err != nil {
			t.Error(err)
		}

		if deployKey == nil {
			t.Error("The Deploy Key could not be retrieved.")
		}

		if deployKey.Id != deployKeyResourceId {
			t.Error("The Deploy Key `label` attribute does not match the expected value.")
		}
		if deployKey.Label != label {
			t.Error("The Deploy Key `label` attribute does not match the expected value.")
		}
		if deployKey.Key != key {
			t.Error("The Deploy Key `key` attribute does not match the expected value.")
		}
	})

	t.Run("delete", func(t *testing.T) {
		opt := &bitbucket.DeployKeyOptions{
			Owner:    owner,
			RepoSlug: repo,
			Id:       deployKeyResourceId,
		}
		_, err := c.Repositories.DeployKeys.Delete(opt)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestDeployKeyWithComment(t *testing.T) {
	user := os.Getenv("BITBUCKET_TEST_USERNAME")
	pass := os.Getenv("BITBUCKET_TEST_PASSWORD")
	owner := os.Getenv("BITBUCKET_TEST_OWNER")
	repo := os.Getenv("BITBUCKET_TEST_REPOSLUG")

	if user == "" {
		t.Error("BITBUCKET_TEST_USERNAME is empty.")
	}
	if pass == "" {
		t.Error("BITBUCKET_TEST_PASSWORD is empty.")
	}
	if owner == "" {
		t.Error("BITBUCKET_TEST_OWNER is empty.")
	}
	if repo == "" {
		t.Error("BITBUCKET_TEST_REPOSLUG is empty.")
	}

	c := bitbucket.NewBasicAuth(user, pass)

	var deployKeyResourceId int

	label := "go-bb-test"
	key := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDAK/b1cHHDr/TEV1JGQl+WjCwStKG6Bhrv0rFpEsYlyTBm1fzN0VOJJYn4ZOPCPJwqse6fGbXntEs+BbXiptR+++HycVgl65TMR0b5ul5AgwrVdZdT7qjCOCgaSV74/9xlHDK8oqgGnfA7ZoBBU+qpVyaloSjBdJfLtPY/xqj4yHnXKYzrtn/uFc4Kp9Tb7PUg9Io3qohSTGJGVHnsVblq/rToJG7L5xIo0OxK0SJSQ5vuId93ZuFZrCNMXj8JDHZeSEtjJzpRCBEXHxpOPhAcbm4MzULgkFHhAVgp4JbkrT99/wpvZ7r9AdkTg7HGqL3rlaDrEcWfL7Lu6TnhBdq5"
	comment := "key@example.com"

	t.Run("create", func(t *testing.T) {
		opt := &bitbucket.DeployKeyOptions{
			Owner:    owner,
			RepoSlug: repo,
			Label:    label,
			Key:      fmt.Sprintf("%s %s", key, comment),
		}

		deployKey, err := c.Repositories.DeployKeys.Create(opt)
		if err != nil {
			t.Error(err)
		}

		if deployKey == nil {
			t.Error("The Deploy Key could not be created.")
		}

		if deployKey.Label != label {
			t.Error("The Deploy Key `label` attribute does not match the expected value.")
		}
		if deployKey.Key != key {
			t.Error("The Deploy Key `key` attribute does not match the expected value.")
		}
		if deployKey.Comment != comment {
			t.Error("The Deploy Key `comment` attribute does not match the expected value.")
		}

		deployKeyResourceId = deployKey.Id
	})

	t.Run("get", func(t *testing.T) {
		opt := &bitbucket.DeployKeyOptions{
			Owner:    owner,
			RepoSlug: repo,
			Id:       deployKeyResourceId,
		}
		deployKey, err := c.Repositories.DeployKeys.Get(opt)
		if err != nil {
			t.Error(err)
		}

		if deployKey == nil {
			t.Error("The Deploy Key could not be retrieved.")
		}

		if deployKey.Id != deployKeyResourceId {
			t.Error("The Deploy Key `label` attribute does not match the expected value.")
		}
		if deployKey.Label != label {
			t.Error("The Deploy Key `label` attribute does not match the expected value.")
		}
		if deployKey.Key != key {
			t.Error("The Deploy Key `key` attribute does not match the expected value.")
		}
		if deployKey.Comment != comment {
			t.Error("The Deploy Key `comment` attribute does not match the expected value.")
		}
	})

	t.Run("delete", func(t *testing.T) {
		opt := &bitbucket.DeployKeyOptions{
			Owner:    owner,
			RepoSlug: repo,
			Id:       deployKeyResourceId,
		}
		_, err := c.Repositories.DeployKeys.Delete(opt)
		if err != nil {
			t.Error(err)
		}
	})
}
