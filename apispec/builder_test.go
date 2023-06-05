package apispec

import (
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"testing"

	"github.com/k0kubun/pp"
)

type TestRequest1 struct {
	Common

	F1 string `json:"f1" comment:"oneday"`
	I2 int64  `json:"i2" comment:"someday"`

	E1 string `json:"e1" comment:"oneday" enum:"EN1"`

	BB struct {
		T1  string `json:"t1" comment:"vivian"`
		Arr []int  `json:"arr"`
	} `json:"bb"`

	DD struct {
		T1  string `json:"t1" comment:"vivian"`
		Brr []int  `json:"arr"`
	} `json:"dd"`

	Items []Item `json:"items"`

	File *multipart.FileHeader `form:"file"`

	Data KV `json:"data"`
}

type KV struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type Pager struct {
	Offset int64 `json:"offset" form:"offset" comment:"偏移量"`
	Limit  int64 `json:"limit" form:"limit" comment:"个数"`
}

type Common struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type Item struct {
	VV string `json:"vv"`
}

type TestResponse struct {
	Common

	Tids []string `json:"tids" comment:"oneday"`

	Total  int64          `json:"total"`
	Items  []Item         `json:"items"`
	Items2 []*Item        `json:"items2"`
	MM     map[int64]Item `json:"mm"`

	Ok Item `json:"ok"`

	NN map[string]bool `json:"nn"`

	ArrayArr [][]int `json:"array_arr"`
}

type EC struct {
	ErrorCode
}

func (e *EC) Code() string {
	return e.ErrorCode.Code
}
func (e *EC) Msg() string {
	return e.ErrorCode.Msg
}
func (e *EC) Desc() string {
	return e.ErrorCode.Desc
}

func TestBuilder(t *testing.T) {
	ab := NewApiSpecBuilder("test", "desc", "v1")

	ab.AddEnum("EN1", NewEnum("e1o1", 1, "e1o1_desc"))
	ab.AddEnum("EN1", NewEnum("e1o2", 2, "e1o2_desc"))
	ab.AddEnum("EN1", NewEnum("e1o3", 3, "e1o3_desc"))

	ab.AddEnum("T2", NewEnum("t2o1", 1, "t2o1_desc"))
	ab.AddEnum("T2", NewEnum("t2o2", 2, "t2o2_desc"))

	ab.AddErrorCode(&EC{ErrorCode{"401", "msg", "401"}})
	ab.AddErrorCode(&EC{ErrorCode{"402", "msg", "402"}})
	ab.AddErrorCode(&EC{ErrorCode{"403", "msg", "403"}})

	ab.AddRoute("default", RouteInfo{
		ID:              "HandleTest1",
		Name:            "test_api_1",
		Desc:            "desc",
		Method:          "POST",
		Uri:             "/test",
		RequestFormat:   ReqFormatJson,
		Request:         TestRequest1{},
		SuccessHttpCode: 200,
		SuccessResponse: TestResponse{},
		ErrorResponses: map[int]interface{}{
			500: Common{},
			400: Common{},
		},
	})

	ab.AddRoute("default", RouteInfo{
		//ID:              "HandleTest2",
		Name:            "test_api_2",
		Desc:            "desc",
		Method:          "GET",
		Uri:             "/test",
		RequestFormat:   ReqFormatQuery,
		Request:         Pager{},
		SuccessHttpCode: 200,
		SuccessResponse: TestResponse{},
		ErrorResponses: map[int]interface{}{
			500: Common{},
			400: Common{},
		},
	})

	ab.AddRoute("default", RouteInfo{
		//ID:              "HandleApiParam",
		Name:            "test_api_param",
		Desc:            "desc",
		Method:          "GET",
		Uri:             "/magazine/:id/:page",
		RequestFormat:   ReqFormatQuery,
		Request:         Pager{},
		SuccessHttpCode: 200,
		SuccessResponse: Common{},
	})

	//ab.AddRoute("static", RouteInfo{
	//	Name:          "download",
	//	Desc:          "desc",
	//	Method:        "GET",
	//	Uri:           "/download/:id",
	//	RequestFormat: ReqFormatQuery,
	//	//Request:         TestRequest1{},
	//	SuccessHttpCode: 200,
	//	//SuccessResponse: TestResponse{},
	//	ErrorResponses: map[string]interface{}{
	//		"4XX": Common{},
	//	},
	//})

	if err := ab.WriteFile("api_spec.yaml"); err != nil {
		t.Fatal(err)
	}
	spec := ab.Build()

	oa3, err := spec.ToOpenApiV3()
	if err != nil {
		t.Fatal(err)
	}

	bb, err := oa3.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	v := make(map[string]interface{})
	json.Unmarshal(bb, &v)

	ybb, err := yaml.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	ioutil.WriteFile("openapi3.yaml", ybb, 0644)

	bb, err = json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	ioutil.WriteFile("openapi3.json", bb, 0644)

	swg, err := spec.ToSwagger()
	if err != nil {
		t.Fatal(err)
	}

	bb, err = swg.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	v = make(map[string]interface{})
	json.Unmarshal(bb, &v)

	ybb, err = yaml.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	ioutil.WriteFile("swagger.yaml", ybb, 0644)

	bb, err = json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	ioutil.WriteFile("swagger.json", bb, 0644)
}

func TestSchemaNameMapping(t *testing.T) {
	scms := map[string]Schema{
		"github.com/SeeJson/backend_scaffold/apispec.Item":                         {},
		"github.com/SeeJson/cccccc/apispec.Item":                                   {},
		"github.com/SeeJson/dddd/apispec.Item":                                     {},
		`struct { T1 string "json:"t1" comment:"vivian"; Arr []int "json:"arr"" }`: {},
		`struct { T1 string "json:"t1" comment:"vivian"; Brr []int "json:"arr"" }`: {},
		"github.com/SeeJson/backend_scaffold/apispec.TestRequest":                  {},
		"github.com/SeeJson/backend_scaffold/apispec.KV":                           {},
	}

	mp := schemaNameMapping(scms)

	pp.Println(mp)
}

func TestGetDefaultRouteID(t *testing.T) {
	pp.Println(getDefaultRouteID("GET", "/test/cc"))
	pp.Println(getDefaultRouteID("GET", "/test/:cc"))
	pp.Println(getDefaultRouteID("GET", "/test/name_list"))
	pp.Println(getDefaultRouteID("POST", "/test/*action"))
}
