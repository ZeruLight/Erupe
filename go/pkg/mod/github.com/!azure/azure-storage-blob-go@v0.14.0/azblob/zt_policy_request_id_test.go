package azblob

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/Azure/azure-pipeline-go/pipeline"
	chk "gopkg.in/check.v1"
)

type requestIDTestScenario int

const (
	// Testing scenarios for echoing Client Request ID
	clientRequestIDMissing             requestIDTestScenario = 1
	errorFromNextPolicy                requestIDTestScenario = 2
	clientRequestIDMatch               requestIDTestScenario = 3
	clientRequestIDNoMatch             requestIDTestScenario = 4
	errorMessageClientRequestIDNoMatch                       = "client Request ID from request and response does not match"
	errorMessageFromNextPolicy                               = "error is not nil"
)

type clientRequestIDPolicy struct {
	matchID  string
	scenario requestIDTestScenario
}

func (p clientRequestIDPolicy) Do(ctx context.Context, request pipeline.Request) (pipeline.Response, error) {
	var header http.Header = make(map[string][]string)
	var err error

	// Set headers and errors according to each scenario
	switch p.scenario {
	case clientRequestIDMissing:
	case errorFromNextPolicy:
		err = errors.New(errorMessageFromNextPolicy)
	case clientRequestIDMatch:
		header.Add(xMsClientRequestID, request.Header.Get(xMsClientRequestID))
	case clientRequestIDNoMatch:
		header.Add(xMsClientRequestID, "fake-client-request-id")
	default:
		header.Add(xMsClientRequestID, newUUID().String())
	}

	response := http.Response{Header: header}

	return pipeline.NewHTTPResponse(&response), err
}

func (s *aztestsSuite) TestEchoClientRequestIDMissing(c *chk.C) {
	factory := NewUniqueRequestIDPolicyFactory()

	// Scenario 1: Client Request ID is missing
	policy := factory.New(clientRequestIDPolicy{scenario: clientRequestIDMissing}, nil)
	request, _ := pipeline.NewRequest("GET", url.URL{}, nil)
	resp, err := policy.Do(context.Background(), request)

	c.Assert(err, chk.IsNil)
	c.Assert(resp, chk.NotNil)
	c.Assert(resp.Response().Header.Get(xMsClientRequestID), chk.Equals, "")
}

func (s *aztestsSuite) TestEchoClientRequestIDErrorFromNextPolicy(c *chk.C) {
	factory := NewUniqueRequestIDPolicyFactory()

	// Scenario 2: Do method returns an error
	policy := factory.New(clientRequestIDPolicy{scenario: errorFromNextPolicy}, nil)
	request, _ := pipeline.NewRequest("GET", url.URL{}, nil)
	resp, err := policy.Do(context.Background(), request)

	c.Assert(err, chk.NotNil)
	c.Assert(err.Error(), chk.Equals, errorMessageFromNextPolicy)
	c.Assert(resp, chk.NotNil)
}

func (s *aztestsSuite) TestEchoClientRequestIDMatch(c *chk.C) {
	factory := NewUniqueRequestIDPolicyFactory()

	// Scenario 3: Client Request ID matches
	matchRequestID := newUUID().String()
	policy := factory.New(clientRequestIDPolicy{matchID: matchRequestID, scenario: clientRequestIDMatch}, nil)
	request, _ := pipeline.NewRequest("GET", url.URL{}, nil)
	request.Header.Set(xMsClientRequestID, matchRequestID)
	resp, err := policy.Do(context.Background(), request)

	c.Assert(err, chk.IsNil)
	c.Assert(resp, chk.NotNil)
	c.Assert(resp.Response().Header.Get(xMsClientRequestID), chk.Equals, request.Header.Get(xMsClientRequestID))
}

func (s *aztestsSuite) TestEchoClientRequestIDNoMatch(c *chk.C) {
	factory := NewUniqueRequestIDPolicyFactory()

	// Scenario 4: Client Request ID does not match
	matchRequestID := newUUID().String()
	policy := factory.New(clientRequestIDPolicy{matchID: matchRequestID, scenario: clientRequestIDNoMatch}, nil)
	request, _ := pipeline.NewRequest("GET", url.URL{}, nil)
	request.Header.Set(xMsClientRequestID, matchRequestID)
	resp, err := policy.Do(context.Background(), request)

	c.Assert(err, chk.NotNil)
	c.Assert(err.Error(), chk.Equals, errorMessageClientRequestIDNoMatch)
	c.Assert(resp, chk.NotNil)
}
