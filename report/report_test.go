/*
   Copyright 2016 Vastech SA (PTY) LTD

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, convey.software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package report

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"github.com/mlesar/grafana-reporter/grafana"
	"github.com/smartystreets/goconvey/convey"
)

const dashJSON = `
{"Dashboard":
	{
		"Title":"My first dashboard",
		"Rows":
		[{"Panels":
			[{"Type":"singlestat", "Id":1},
			 {"Type":"graph", "Id":22}]
		},
		{"Panels":
			[
				{"Type":"singlestat", "Id":33},
				{"Type":"graph", "Id":44},
				{"Type":"graph", "Id":55},
				{"Type":"graph", "Id":66},
				{"Type":"graph", "Id":77},
				{"Type":"graph", "Id":88},
				{"Type":"graph", "Id":99}
			]
		}]
	},
"Meta":
	{"Slug":"testDash"}
}`

type mockGrafanaClient struct {
	getPanelCallCount int
	variables         url.Values
}

func (m *mockGrafanaClient) GetDashboard(dashName string) (grafana.Dashboard, error) {
	return grafana.NewDashboard([]byte(dashJSON), m.variables), nil
}

func (m *mockGrafanaClient) GetPanelPng(p grafana.Panel, dashName string, t grafana.TimeRange) (io.ReadCloser, error) {
	m.getPanelCallCount++
	return ioutil.NopCloser(bytes.NewBuffer([]byte("Not actually a png"))), nil
}

func TestReport(t *testing.T) {
	convey.Convey("When generating a report", t, func(c convey.C) {
		variables := url.Values{}
		variables.Add("var-test", "testvarvalue")
		gClient := &mockGrafanaClient{0, variables}
		rep := New(gClient, "testDash", grafana.TimeRange{From: "1453206447000", To: "1453213647000"}, "", false)
		defer func() {
			c.So(rep.Clean(), convey.ShouldBeNil)
		}()

		c.Convey("When rendering images", func(c convey.C) {
			dashboard, _ := gClient.GetDashboard("")
			rep.renderPNGsParallel(dashboard)

			c.Convey("It should create a temporary folder", func(c convey.C) {
				_, err := os.Stat(rep.tmpDir)
				c.So(err, convey.ShouldBeNil)
			})

			c.Convey("It should copy the file to the image folder", func(c convey.C) {
				_, err := os.Stat(rep.imgDirPath() + "/image1.png")
				c.So(err, convey.ShouldBeNil)
			})

			c.Convey("It shoud call getPanelPng once per panel", func(c convey.C) {
				c.So(gClient.getPanelCallCount, convey.ShouldEqual, 9)
			})

			c.Convey("It should create one file per panel", func(c convey.C) {
				f, err := os.Open(rep.imgDirPath())
				c.So(err, convey.ShouldBeNil)
				defer func() {
					err := f.Close()
					c.So(err, convey.ShouldBeNil)
				}()
				files, err := f.Readdir(0)
				c.So(files, convey.ShouldHaveLength, 9)
				c.So(err, convey.ShouldBeNil)
			})
		})

		c.Convey("When genereting the Tex file", func(c convey.C) {
			dashboard, _ := gClient.GetDashboard("")
			rep.generateTeXFile(dashboard)
			f, err := os.Open(rep.texPath())
			c.So(err, convey.ShouldBeNil)
			defer func() {
				err := f.Close()
				c.So(err, convey.ShouldBeNil)
			}()

			c.Convey("It should create a file in the temporary folder", func(c convey.C) {
				c.So(err, convey.ShouldBeNil)
			})

			c.Convey("The file should contain reference to the template data", func(c convey.C) {
				var buf bytes.Buffer
				io.Copy(&buf, f)
				s := buf.String()

				c.So(err, convey.ShouldBeNil)
				c.Convey("Including the Title", func(c convey.C) {
					c.So(s, convey.ShouldContainSubstring, "My first dashboard")

				})
				c.Convey("Including the varialbe values", func(c convey.C) {
					c.So(s, convey.ShouldContainSubstring, "testvarvalue")
				})
				c.Convey("and the images", func(c convey.C) {
					c.So(s, convey.ShouldContainSubstring, "image1")
					c.So(s, convey.ShouldContainSubstring, "image22")
					c.So(s, convey.ShouldContainSubstring, "image33")
					c.So(s, convey.ShouldContainSubstring, "image44")
					c.So(s, convey.ShouldContainSubstring, "image55")
					c.So(s, convey.ShouldContainSubstring, "image66")
					c.So(s, convey.ShouldContainSubstring, "image77")
					c.So(s, convey.ShouldContainSubstring, "image88")
					c.So(s, convey.ShouldContainSubstring, "image99")
				})
				c.Convey("and the time range", func(c convey.C) {
					//server time zone by shift hours timestamp
					//convey.so just test for day and year
					c.So(s, convey.ShouldContainSubstring, "Tue Jan 19")
					c.So(s, convey.ShouldContainSubstring, "2016")
				})
			})
		})

		c.Convey("Clean() should remove the temporary folder", func(c convey.C) {
			c.So(rep.Clean(), convey.ShouldBeNil)

			_, err := os.Stat(rep.tmpDir)
			c.So(os.IsNotExist(err), convey.ShouldBeTrue)
		})
	})

}

type errClient struct {
	getPanelCallCount int
	variables         url.Values
}

func (e *errClient) GetDashboard(dashName string) (grafana.Dashboard, error) {
	return grafana.NewDashboard([]byte(dashJSON), e.variables), nil
}

//Produce an error on the 2nd panel fetched
func (e *errClient) GetPanelPng(p grafana.Panel, dashName string, t grafana.TimeRange) (io.ReadCloser, error) {
	e.getPanelCallCount++
	if e.getPanelCallCount == 2 {
		return nil, errors.New("The second panel has convey.some problem")
	}
	return ioutil.NopCloser(bytes.NewBuffer([]byte("Not actually a png"))), nil
}

func TestReportErrorHandling(t *testing.T) {
	convey.Convey("When generating a report where one panels gives an error", t, func(c convey.C) {
		variables := url.Values{}
		gClient := &errClient{0, variables}
		rep := New(gClient, "testDash", grafana.TimeRange{From: "1453206447000", To: "1453213647000"}, "", false)
		defer func() {
			c.So(rep.Clean(), convey.ShouldBeNil)
		}()

		c.Convey("When rendering images", func(c convey.C) {
			dashboard, _ := gClient.GetDashboard("")
			err := rep.renderPNGsParallel(dashboard)

			c.Convey("It shoud call getPanelPng once per panel", func(c convey.C) {
				c.So(gClient.getPanelCallCount, convey.ShouldEqual, 9)
			})

			c.Convey("It should create one less image file than the total number of panels", func(c convey.C) {
				f, err := os.Open(rep.imgDirPath())
				c.So(err, convey.ShouldBeNil)
				defer func() {
					err := f.Close()
					c.So(err, convey.ShouldBeNil)
				}()
				files, err := f.Readdir(0)
				c.So(files, convey.ShouldHaveLength, 8) //one less than the total number of im
				c.So(err, convey.ShouldBeNil)
			})

			c.Convey("If any panels return errors, renderPNGsParralel should return the error message from one panel", func(c convey.C) {
				c.So(err, convey.ShouldNotBeNil)
				c.So(err.Error(), convey.ShouldContainSubstring, "The second panel has convey.some problem")
			})
		})

		c.Convey("Clean() should remove the temporary folder", func(c convey.C) {
			c.So(rep.Clean(), convey.ShouldBeNil)

			_, err := os.Stat(rep.tmpDir)
			c.So(os.IsNotExist(err), convey.ShouldBeTrue)
		})
	})

}
