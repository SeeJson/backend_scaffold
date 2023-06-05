package apispec

import (
	"encoding/json"
	"errors"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"path/filepath"
)

type ApiSpec struct {
	Info

	RouteGroups []RouteGroup `json:"route_groups" yaml:"route_groups"`

	Schemas map[string]Schema `json:"schemas" yaml:"schemas"`

	Enums      map[string][]Enum `json:"enums" yaml:"enums"`
	ErrorCodes []ErrorCode       `json:"error_codes" yaml:"error_codes"`
}

type Info struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	Version     string `json:"version" yaml:"version"`
}

type RouteGroup struct {
	Name string `json:"name" yaml:"name"`

	Routes []Route `json:"routes" yaml:"routes"`
}

type Route struct {
	ID   string `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
	Desc string `json:"desc" yaml:"desc"`

	Method      string `json:"method" yaml:"method"`
	Uri         string `json:"uri" yaml:"uri"`
	Permissions string `json:"permissions" yaml:"permissions"`

	Request         *Request   `yaml:"request" json:"request"`
	SuccessResponse Response   `yaml:"success_response"`
	ErrorResponses  []Response `yaml:"error_responses"`
}

type Request struct {
	Format      string `yaml:"format" json:"format"`
	Schema      string `yaml:"schema" json:"schema"`
	ParamSchema string `yaml:"param_schema" json:"param_schema"`

	JsonExample string `yaml:"json_example" json:"json_example"`
}

type Response struct {
	HttpCode int    `yaml:"http_code" json:"http_code"`
	Format   string `yaml:"format" json:"format"`

	Schema string `yaml:"schema" json:"schema"`

	JsonExample string `yaml:"json_example" json:"json_example"`
}

type Enum struct {
	Name string `json:"name" yaml:"name"`
	Int  int8   `json:"int" yaml:"int"`
	Desc string `json:"desc" yaml:"desc"`
}

type ErrorCode struct {
	Code string `json:"code" yaml:"code"`
	Msg  string `json:"msg" yaml:"msg"`
	Desc string `json:"desc" yaml:"desc"`
}

type Schema struct {
	ID     string  `json:"id" yaml:"id"`
	Name   string  `json:"name" yaml:"name"`
	Fields []Field `json:"fields" yaml:"fields"`
}

type Field struct {
	Name string `json:"name" yaml:"name"`
	//Type    string `json:"type" yaml:"type"`
	Type    FieldType `json:"type" yaml:"type"`
	Comment string    `json:"comment" yaml:"comment"`
	Enum    string    `json:"enum" yaml:"enum"`
	//Required bool
}

type FieldType struct {
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type" json:"type"`
	Key  string `yaml:"key" json:"key"`
	Ref  string `yaml:"ref" json:"ref"`
}

func LoadFile(fname string) (ApiSpec, error) {
	var spec ApiSpec

	bb, err := ioutil.ReadFile(fname)
	if err != nil {
		return spec, err
	}

	ext := filepath.Ext(fname)
	switch ext {
	case ".json":
		err = json.Unmarshal(bb, &spec)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(bb, &spec)
	default:
		return spec, errors.New("unsupported format, only support json or yaml")
	}

	return spec, err
}
