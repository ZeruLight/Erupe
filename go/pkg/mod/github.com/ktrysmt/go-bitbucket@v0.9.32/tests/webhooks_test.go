package tests

import (
	"os"
	"testing"

	"github.com/ktrysmt/go-bitbucket"
)

func TestWebhook(t *testing.T) {
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

	var webhookResourceUuid string

	t.Run("create", func(t *testing.T) {
		opt := &bitbucket.WebhooksOptions{
			Owner:       owner,
			RepoSlug:    repo,
			Description: "go-bb-test",
			Url:         "https://example.com",
			Active:      false,
			Events:      []string{"repo:push", "issue:created"},
		}

		webhook, err := c.Repositories.Webhooks.Create(opt)
		if err != nil {
			t.Error(err)
		}

		if webhook == nil {
			t.Error("The webhook could not be created.")
		}

		if webhook.Description != "go-bb-test" {
			t.Error("The webhook `description` attribute does not match the expected value.")
		}
		if webhook.Url != "https://example.com" {
			t.Error("The webhook `url` attribute does not match the expected value.")
		}
		if webhook.Active != false {
			t.Error("The webhook `active` attribute does not match the expected value.")
		}
		if len(webhook.Events) != 2 {
			t.Error("The webhook `events` attribute does not match the expected value.")
		}

		webhookResourceUuid = webhook.Uuid
	})

	t.Run("get", func(t *testing.T) {
		opt := &bitbucket.WebhooksOptions{
			Owner:    owner,
			RepoSlug: repo,
			Uuid:     webhookResourceUuid,
		}
		webhook, err := c.Repositories.Webhooks.Get(opt)
		if err != nil {
			t.Error(err)
		}

		if webhook == nil {
			t.Error("The webhook could not be retrieved.")
		}

		if webhook.Description != "go-bb-test" {
			t.Error("The webhook `description` attribute does not match the expected value.")
		}
		if webhook.Url != "https://example.com" {
			t.Error("The webhook `url` attribute does not match the expected value.")
		}
		if webhook.Active != false {
			t.Error("The webhook `active` attribute does not match the expected value.")
		}
		if len(webhook.Events) != 2 {
			t.Error("The webhook `events` attribute does not match the expected value.")
		}
	})

	t.Run("update", func(t *testing.T) {
		opt := &bitbucket.WebhooksOptions{
			Owner:       owner,
			RepoSlug:    repo,
			Uuid:        webhookResourceUuid,
			Description: "go-bb-test-new",
			Url:         "https://new-example.com",
			Events:      []string{"repo:push", "issue:created", "repo:fork"},
		}
		webhook, err := c.Repositories.Webhooks.Update(opt)
		if err != nil {
			t.Error(err)
		}

		if webhook == nil {
			t.Error("The webhook could not be retrieved.")
		}

		if webhook.Description != "go-bb-test-new" {
			t.Error("The webhook `description` attribute does not match the expected value.")
		}
		if webhook.Url != "https://new-example.com" {
			t.Error("The webhook `url` attribute does not match the expected value.")
		}
		if webhook.Active != false {
			t.Error("The webhook `active` attribute does not match the expected value.")
		}
		if len(webhook.Events) != 3 {
			t.Error("The webhook `events` attribute does not match the expected value.")
		}
	})

	t.Run("delete", func(t *testing.T) {
		opt := &bitbucket.WebhooksOptions{
			Owner:    owner,
			RepoSlug: repo,
			Uuid:     webhookResourceUuid,
		}
		_, err := c.Repositories.Webhooks.Delete(opt)
		if err != nil {
			t.Error(err)
		}
	})
}
