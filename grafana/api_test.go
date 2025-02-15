/*
   Copyright 2016 Vastech SA (PTY) LTD

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package grafana

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestGrafanaClientFetchesDashboard(t *testing.T) {
	convey.Convey("When fetching a Dashboard", t, func(c convey.C) {
		requestURI := ""
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestURI = r.RequestURI
			fmt.Fprintln(w, `{"":""}`)
		}))
		defer ts.Close()

		c.Convey("When using the Grafana v4 client", func(c convey.C) {
			grf := NewV4Client(ts.URL, "", url.Values{}, true, false)
			grf.GetDashboard("testDash")

			c.Convey("It should use the v4 dashboards endpoint", func(c convey.C) {
				c.So(requestURI, convey.ShouldEqual, "/api/dashboards/db/testDash")
			})
		})

		c.Convey("When using the Grafana v5 client", func(c convey.C) {
			grf := NewV5Client(ts.URL, "", url.Values{}, true, false)
			grf.GetDashboard("rYy7Paekz")

			c.Convey("It should use the v5 dashboards endpoint", func(c convey.C) {
				c.So(requestURI, convey.ShouldEqual, "/api/dashboards/uid/rYy7Paekz")
			})
		})

	})
}

func TestGrafanaClientFetchesPanelPNG(t *testing.T) {
	convey.Convey("When fetching a panel PNG", t, func(c convey.C) {
		requestURI := ""
		requestHeaders := http.Header{}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestURI = r.RequestURI
			requestHeaders = r.Header
		}))
		defer ts.Close()

		apiToken := "1234"
		variables := url.Values{}
		variables.Add("var-host", "servername")
		variables.Add("var-port", "adapter")

		cases := map[string]struct {
			client      Client
			pngEndpoint string
		}{
			"v4": {NewV4Client(ts.URL, apiToken, variables, true, false), "/render/dashboard-solo/db/testDash"},
			"v5": {NewV5Client(ts.URL, apiToken, variables, true, false), "/render/d-solo/testDash/_"},
		}
		for clientDesc, cl := range cases {
			grf := cl.client
			grf.GetPanelPng(Panel{44, "singlestat", "title", GridPos{0, 0, 0, 0}}, "testDash", TimeRange{"now-1h", "now"})

			c.Convey(fmt.Sprintf("The %s client should use the render endpoint with the dashboard name", clientDesc), func(c convey.C) {
				c.So(requestURI, convey.ShouldStartWith, cl.pngEndpoint)
			})

			c.Convey(fmt.Sprintf("The %s client should request the panel ID", clientDesc), func(c convey.C) {
				c.So(requestURI, convey.ShouldContainSubstring, "panelId=44")
			})

			c.Convey(fmt.Sprintf("The %s client should request the time", clientDesc), func(c convey.C) {
				c.So(requestURI, convey.ShouldContainSubstring, "from=now-1h")
				c.So(requestURI, convey.ShouldContainSubstring, "to=now")
			})

			c.Convey(fmt.Sprintf("The %s client should insert auth token should in request header", clientDesc), func(c convey.C) {
				c.So(requestHeaders.Get("Authorization"), convey.ShouldContainSubstring, apiToken)
			})

			c.Convey(fmt.Sprintf("The %s client should pass variables in the request parameters", clientDesc), func(c convey.C) {
				c.So(requestURI, convey.ShouldContainSubstring, "var-host=servername")
				c.So(requestURI, convey.ShouldContainSubstring, "var-port=adapter")
			})

			c.Convey(fmt.Sprintf("The %s client should request singlestat panels at a smaller size", clientDesc), func(c convey.C) {
				c.So(requestURI, convey.ShouldContainSubstring, "width=300")
				c.So(requestURI, convey.ShouldContainSubstring, "height=150")
			})

			c.Convey(fmt.Sprintf("The %s client should request text panels with a small height", clientDesc), func(c convey.C) {
				grf.GetPanelPng(Panel{44, "text", "title", GridPos{0, 0, 0, 0}}, "testDash", TimeRange{"now", "now-1h"})
				c.So(requestURI, convey.ShouldContainSubstring, "width=1000")
				c.So(requestURI, convey.ShouldContainSubstring, "height=100")
			})

			c.Convey(fmt.Sprintf("The %s client should request other panels in a larger size", clientDesc), func(c convey.C) {
				grf.GetPanelPng(Panel{44, "graph", "title", GridPos{0, 0, 0, 0}}, "testDash", TimeRange{"now", "now-1h"})
				c.So(requestURI, convey.ShouldContainSubstring, "width=1000")
				c.So(requestURI, convey.ShouldContainSubstring, "height=500")
			})
		}

		casesGridLayout := map[string]struct {
			client      Client
			pngEndpoint string
		}{
			"v4": {NewV4Client(ts.URL, apiToken, variables, true, true), "/render/dashboard-solo/db/testDash"},
			"v5": {NewV5Client(ts.URL, apiToken, variables, true, true), "/render/d-solo/testDash/_"},
		}
		for clientDesc, cl := range casesGridLayout {
			grf := cl.client

			c.Convey(fmt.Sprintf("The %s client should request grid layout panels with width=1000 and height=240", clientDesc), func(c convey.C) {
				grf.GetPanelPng(Panel{44, "graph", "title", GridPos{6, 24, 0, 0}}, "testDash", TimeRange{"now", "now-1h"})
				c.So(requestURI, convey.ShouldContainSubstring, "width=960")
				c.So(requestURI, convey.ShouldContainSubstring, "height=240")
			})

			c.Convey(fmt.Sprintf("The %s client should request grid layout panels with width=480 and height=120", clientDesc), func(c convey.C) {
				grf.GetPanelPng(Panel{44, "graph", "title", GridPos{3, 12, 0, 0}}, "testDash", TimeRange{"now", "now-1h"})
				c.So(requestURI, convey.ShouldContainSubstring, "width=480")
				c.So(requestURI, convey.ShouldContainSubstring, "height=120")
			})
		}

	})
}

func init() {
	getPanelRetrySleepTime = time.Duration(1) * time.Millisecond //we want our tests to run fast
}

func TestGrafanaClientFetchPanelPNGErrorHandling(t *testing.T) {
	convey.Convey("When trying to fetching a panel from the server sometimes returns an error", t, func(c convey.C) {
		try := 0

		//create a server that will return error on the first call
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if try < 1 {
				w.WriteHeader(http.StatusInternalServerError)
				try++
			}
		}))
		defer ts.Close()

		grf := NewV4Client(ts.URL, "", url.Values{}, true, false)

		_, err := grf.GetPanelPng(Panel{44, "singlestat", "title", GridPos{0, 0, 0, 0}}, "testDash", TimeRange{"now-1h", "now"})

		c.Convey("It should retry a couple of times if it receives errors", func(c convey.C) {
			c.So(err, convey.ShouldBeNil)
		})
	})

	convey.Convey("When trying to fetching a panel from the server consistently returns an error", t, func(c convey.C) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		grf := NewV4Client(ts.URL, "", url.Values{}, true, false)

		_, err := grf.GetPanelPng(Panel{44, "singlestat", "title", GridPos{0, 0, 0, 0}}, "testDash", TimeRange{"now-1h", "now"})

		c.Convey("The Grafana API should return an error", func(c convey.C) {
			c.So(err, convey.ShouldNotBeNil)
		})
	})
}
