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
	"net/url"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestV4Dashboard(t *testing.T) {
	convey.Convey("When creating a new dashboard from Grafana v4 dashboard JSON", t, func(c convey.C) {
		const v4DashJSON = `
{"Dashboard":
	{
		"Rows":
			[{
				"Panels":
					[{"Type":"singlestat", "Id":1},
					{"Type":"graph", "Id":2}],
				"Title": "RowTitle #"
			},
			{"Panels":
				[{"Type":"singlestat", "Id":3, "Title": "Panel3Title #"}]
			}],
		"title":"DashTitle #"
	},
"Meta":
	{"Slug":"testDash"}
}`
		dash := NewDashboard([]byte(v4DashJSON), url.Values{})

		c.Convey("Panel Is(type) should work for all panels", func(c convey.C) {
			c.So(dash.Panels[0].Is(Graph), convey.ShouldBeFalse)
			c.So(dash.Panels[0].Is(Text), convey.ShouldBeFalse)
			c.So(dash.Panels[0].Is(Table), convey.ShouldBeFalse)
			c.So(dash.Panels[0].Is(SingleStat), convey.ShouldBeTrue)
			c.So(dash.Panels[1].Is(Graph), convey.ShouldBeTrue)
			c.So(dash.Panels[2].Is(SingleStat), convey.ShouldBeTrue)
		})

		c.Convey("Row title should be parsed and santised", func(c convey.C) {
			c.So(dash.Rows[0].Title, convey.ShouldEqual, "RowTitle \\#")
		})

		c.Convey("Panel titles should be parsed and sanitised", func(c convey.C) {
			c.So(dash.Panels[2].Title, convey.ShouldEqual, "Panel3Title \\#")
		})

		c.Convey("When accessing Panels from within Rows, titles should still be sanitised", func(c convey.C) {
			c.So(dash.Rows[1].Panels[0].Title, convey.ShouldEqual, "Panel3Title \\#")
		})

		c.Convey("Panels should contain all panels from all rows", func(c convey.C) {
			c.So(dash.Panels, convey.ShouldHaveLength, 3)
		})

		c.Convey("The Title should be parsed and sanitised", func(c convey.C) {
			c.So(dash.Title, convey.ShouldEqual, "DashTitle \\#")
		})
	})
}

func TestV5Dashboard(t *testing.T) {
	convey.Convey("When creating a new dashboard from Grafana v5 dashboard JSON", t, func(c convey.C) {
		const v5DashJSON = `
{"Dashboard":
	{
		"Panels":
			[{"Type":"singlestat", "Id":0},
			{"Type":"graph", "Id":1, "GridPos":{"H":6,"W":24,"X":0,"Y":0}},
			{"Type":"singlestat", "Id":2, "Title":"Panel3Title #"},
			{"Type":"text", "GridPos":{"H":6.5,"W":20.5,"X":0,"Y":0}, "Id":3},
			{"Type":"table", "Id":4},
			{"Type":"row", "Id":5}],
		"Title":"DashTitle #"
	},

"Meta":
	{"Slug":"testDash"}
}`
		dash := NewDashboard([]byte(v5DashJSON), url.Values{})

		c.Convey("Panel Is(type) should work for all panels", func(c convey.C) {
			c.So(dash.Panels[0].Is(SingleStat), convey.ShouldBeTrue)
			c.So(dash.Panels[1].Is(Graph), convey.ShouldBeTrue)
			c.So(dash.Panels[2].Is(SingleStat), convey.ShouldBeTrue)
			c.So(dash.Panels[3].Is(Text), convey.ShouldBeTrue)
			c.So(dash.Panels[4].Is(Table), convey.ShouldBeTrue)
		})

		c.Convey("Panel titles should be parsed and sanitised", func(c convey.C) {
			c.So(dash.Panels[2].Title, convey.ShouldEqual, "Panel3Title \\#")
		})

		c.Convey("Panels should contain all panels that have type != row", func(c convey.C) {
			c.So(dash.Panels, convey.ShouldHaveLength, 5)
			c.So(dash.Panels[0].Id, convey.ShouldEqual, 0)
			c.So(dash.Panels[1].Id, convey.ShouldEqual, 1)
			c.So(dash.Panels[2].Id, convey.ShouldEqual, 2)
		})

		c.Convey("The Title should be parsed", func(c convey.C) {
			c.So(dash.Title, convey.ShouldEqual, "DashTitle \\#")
		})

		c.Convey("Panels should contain GridPos H & W", func(c convey.C) {
			c.So(dash.Panels[1].GridPos.H, convey.ShouldEqual, 6)
			c.So(dash.Panels[1].GridPos.W, convey.ShouldEqual, 24)
		})

		c.Convey("Panels GridPos should allow floatt", func(c convey.C) {
			c.So(dash.Panels[3].GridPos.H, convey.ShouldEqual, 6.5)
			c.So(dash.Panels[3].GridPos.W, convey.ShouldEqual, 20.5)
		})

	})
}

func TestVariableValues(t *testing.T) {
	convey.Convey("When creating a dashboard and passing url varialbes in", t, func(c convey.C) {
		const v5DashJSON = `
{
	"Dashboard":
		{
		}
}`
		vars := url.Values{}
		vars.Add("var-one", "oneval")
		vars.Add("var-two", "twoval")
		dash := NewDashboard([]byte(v5DashJSON), vars)

		c.Convey("The dashboard should contain the variable values in a random order", func(c convey.C) {
			c.So(dash.VariableValues, convey.ShouldContainSubstring, "oneval")
			c.So(dash.VariableValues, convey.ShouldContainSubstring, "twoval")
		})
	})
}
