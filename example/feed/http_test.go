package feed

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	kitlogger "github.com/go-kit/log"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type FeedHTTPTransportTestSuite struct {
	suite.Suite
}

func (s *FeedHTTPTransportTestSuite) TestHTTPTransport() {
	tests := []struct {
		name           string
		method         string
		url            string
		mocks          func(svc *MockFeedService)
		requestPayload []byte
		expectedCode   int
		expectedBody   []byte
	}{
		{
			name:   "ValidListFeedsRequest",
			method: http.MethodGet,
			url:    "/feeds",
			mocks: func(svc *MockFeedService) {
				svc.On("ListFeeds", mock.Anything).Return(
					[]Feed{
						{},
					},
					nil,
				)
			},
			requestPayload: []byte(``),
			expectedCode:   http.StatusOK,
			expectedBody: []byte(`{
				"feeds": [
					{
						"id": 0,
						"name": "",
						"posts": null,
						"link": "",
						"createdAt": "0001-01-01T00:00:00Z",
						"updatedAt": "0001-01-01T00:00:00Z",
						"deletedAt": "0001-01-01T00:00:00Z"
					}
				]
			}`),
		},
		{
			name:   "ValidCreateFeedRequest",
			method: http.MethodPost,
			url:    "/feeds",
			mocks: func(svc *MockFeedService) {
				svc.On("CreateFeed", mock.Anything, mock.AnythingOfType("*feed.Feed")).Return(nil)
			},
			requestPayload: []byte(`{
					"name": "my-test-feed",
					"link": "https://www.myrssfeed.com"
				}`),
			expectedCode: http.StatusOK,
			expectedBody: []byte(`{
				"feed": {
					"id": 0,
					"name": "my-test-feed",
					"posts": null,
					"link": "https://www.myrssfeed.com",
					"createdAt": "0001-01-01T00:00:00Z",
					"updatedAt": "0001-01-01T00:00:00Z",
					"deletedAt": "0001-01-01T00:00:00Z"
				}
			}`),
		},
		// Add more test cases as needed
	}

	for _, test := range tests {
		test := test
		s.T().Run(test.name, func(t *testing.T) {
			svc := NewMockFeedService(s.T())
			e := NewEndpointSet(svc)
			logger := kitlogger.NewJSONLogger(kitlogger.NewSyncWriter(os.Stderr))

			ts := httptest.NewServer(NewHTTPHandler(e, logger))
			defer ts.Close()

			test.mocks(svc)

			request, err := http.NewRequest(test.method, ts.URL+test.url, bytes.NewBuffer(test.requestPayload))
			if err != nil {
				log.Fatal(err)
			}

			client := &http.Client{}
			response, err := client.Do(request)
			if err != nil {
				log.Fatal(err)
			}
			defer response.Body.Close()

			var resBody interface{}
			json.NewDecoder(response.Body).Decode(&resBody)

			var expectedBody interface{}
			json.NewDecoder(bytes.NewBuffer(test.expectedBody)).Decode(&expectedBody)

			s.Assert().Equal(test.expectedCode, response.StatusCode)
			s.Assert().Equal(expectedBody, resBody)
		})
	}

}

func TestFeedHTTPTransportTestSuite(t *testing.T) {
	suite.Run(t, new(FeedHTTPTransportTestSuite))
}
