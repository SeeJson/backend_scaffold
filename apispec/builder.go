package apispec

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/k0kubun/pp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type EnumItem interface {
	Name() string
	Int() int8
	Desc() string
}

type ErrorCodeItem interface {
	Code() string
	Msg() string
	Desc() string
}

type ApiSpecBuilder struct {
	spec ApiSpec

	pkt2schema map[string]string
}

func NewApiSpecBuilder(title, desc, version string) *ApiSpecBuilder {
	return &ApiSpecBuilder{
		spec: ApiSpec{
			Info: Info{
				Title:       title,
				Description: desc,
				Version:     version,
			},
			Schemas: make(map[string]Schema),
			Enums:   make(map[string][]Enum),
		},
	}
}

func (ab *ApiSpecBuilder) AddEnum(typ string, en EnumItem) {
	ens, ok := ab.spec.Enums[typ]
	if ok {
		ens = append(ens, Enum{Name: en.Name(), Int: en.Int(), Desc: en.Desc()})
		sort.Slice(ens, func(i, j int) bool {
			return ens[i].Int < ens[j].Int
		})
		ab.spec.Enums[typ] = ens
	} else {
		ab.spec.Enums[typ] = []Enum{
			{Name: en.Name(), Int: en.Int(), Desc: en.Desc()},
		}
	}
}

func (ab *ApiSpecBuilder) AddErrorCode(ecs ...ErrorCodeItem) {
	for _, ec := range ecs {
		ab.spec.ErrorCodes = append(ab.spec.ErrorCodes, ErrorCode{
			Code: ec.Code(),
			Msg:  ec.Msg(),
			Desc: ec.Desc(),
		})
	}
}

func (ab *ApiSpecBuilder) WriteFile(fname string) error {
	spec := ab.Build()

	var (
		bb  []byte
		err error
	)
	ext := filepath.Ext(fname)
	switch ext {
	case ".json":
		bb, err = json.Marshal(spec)
	case ".yaml", ".yml":
		bb, err = yaml.Marshal(spec)
	default:
		return errors.New("unsupported format, only support json or yaml")
	}

	if err != nil {
		return err
	}
	return ioutil.WriteFile(fname, bb, 0644)
}

func (ab *ApiSpecBuilder) Write(format string) ([]byte, error) {
	spec := ab.Build()

	var (
		bb  []byte
		err error
	)
	switch format {
	case "json":
		bb, err = json.Marshal(spec)
	case "yaml":
		bb, err = yaml.Marshal(spec)
	default:
		return nil, errors.New("unsupported format, only support json or yaml")
	}
	return bb, err
}

type RouteInfo struct {
	ID   string
	Name string
	Desc string

	Method      string
	Uri         string
	Permissions string

	RequestFormat string
	Request       interface{}
	Param         interface{}

	SuccessHttpCode int
	SuccessResponse interface{}
	ErrorResponses  map[int]interface{}
}

func (ab *ApiSpecBuilder) AddRoute(section string, r RouteInfo) {
	route, scms := ab.parseRouteInfo(r)

	for k, v := range scms {
		ab.spec.Schemas[k] = v.Schema
	}

	idx := -1
	for i := range ab.spec.RouteGroups {
		if ab.spec.RouteGroups[i].Name == section {
			idx = i
		}
	}
	if idx < 0 {
		ab.spec.RouteGroups = append(ab.spec.RouteGroups, RouteGroup{
			Name:   section,
			Routes: []Route{route},
		})
	} else {
		ab.spec.RouteGroups[idx].Routes = append(ab.spec.RouteGroups[idx].Routes, route)
	}
}

func (ab *ApiSpecBuilder) parseRouteInfo(r RouteInfo) (Route, map[string]schemaX) {
	route := Route{
		ID:          r.ID,
		Name:        r.Name,
		Desc:        r.Desc,
		Method:      r.Method,
		Uri:         r.Uri,
		Permissions: r.Permissions,
	}

	schemas := make(map[string]schemaX)
	if r.Request != nil {
		route.Request = &Request{
			Format: r.RequestFormat,
		}
		if r.RequestFormat == ReqFormatJson {
			bb, err := json.MarshalIndent(r.Request, "", "  ")
			if err != nil {
				log.Error("marshal json request error")
			}
			route.Request.JsonExample = string(bb)
		}
		scms := ab.ParseModel(r.Request)
		//pp.Println(scms)
		if len(scms) == 0 {
			log.Error("ParseModel error")
		} else {
			route.Request.Schema = scms[0].Name
			for _, s := range scms {
				schemas[s.Name] = s
			}
		}
	}
	if r.SuccessResponse != nil {
		route.SuccessResponse = Response{
			HttpCode: r.SuccessHttpCode,
			Format:   ResponseFormatJson, // 先写死吧
		}
		if r.RequestFormat == ResponseFormatJson {
			bb, err := json.MarshalIndent(r.SuccessResponse, "", "  ")
			if err != nil {
				log.Error("marshal json SuccessResponse error")
			}
			route.SuccessResponse.JsonExample = string(bb)
		}

		scms := ab.ParseModel(r.SuccessResponse)
		//pp.Println(scms)
		if len(scms) == 0 {
			log.Error("ParseModel error")
		} else {
			route.SuccessResponse.Schema = scms[0].Name
			for _, s := range scms {
				schemas[s.Name] = s
			}
		}
	} else {
		// 没有response类型的算返回文件
		route.SuccessResponse = Response{
			HttpCode: r.SuccessHttpCode,
			Format:   ResponseFormatFile,
		}
	}

	for httpCode, resp := range r.ErrorResponses {
		eRsp := Response{
			HttpCode: httpCode,
			Format:   ResponseFormatJson, // 先写死吧
		}

		bb, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			log.Error("marshal json ErrorResponse error")
		}
		eRsp.JsonExample = string(bb)

		scms := ab.ParseModel(resp)
		//pp.Println(scms)
		if len(scms) == 0 {
			log.Error("ParseModel error")
		} else {
			eRsp.Schema = scms[0].Name
			for _, s := range scms {
				schemas[s.Name] = s
			}
		}

		route.ErrorResponses = append(route.ErrorResponses, eRsp)
	}

	return route, schemas
}

type schemaX struct {
	idx int
	pkg string

	Schema
}

func (ab *ApiSpecBuilder) ParseModel(v interface{}) []schemaX {
	tables := make(map[string]schemaX)
	modelName := ParseStruct(v, tables)
	_ = modelName

	scms := make([]schemaX, 0, len(tables))
	for _, s := range tables {
		scms = append(scms, s)
	}
	sort.Slice(scms, func(i, j int) bool {
		return scms[i].idx < scms[j].idx
	})

	return scms
}

func (ab *ApiSpecBuilder) Build() ApiSpec {
	// json deep copy -_-!!
	bb, _ := json.Marshal(ab.spec)
	var spec ApiSpec
	json.Unmarshal(bb, &spec)

	mapping := schemaNameMapping(spec.Schemas)
	for i := range spec.RouteGroups {
		for j, r := range spec.RouteGroups[i].Routes {
			if r.Request != nil {
				spec.RouteGroups[i].Routes[j].Request.Schema = mapping[r.Request.Schema]
			}
			spec.RouteGroups[i].Routes[j].SuccessResponse.Schema = mapping[r.SuccessResponse.Schema]
			for k, e := range r.ErrorResponses {
				spec.RouteGroups[i].Routes[j].ErrorResponses[k].Schema = mapping[e.Schema]
			}
		}
	}
	scms := make(map[string]Schema)
	for k, scm := range spec.Schemas {
		scm.Name = mapping[scm.Name]
		for i, f := range scm.Fields {
			switch f.Type.Type {
			case FieldTypeObject:
				scm.Fields[i].Type.Name = mapping[f.Type.Name]
				scm.Fields[i].Type.Ref = mapping[f.Type.Ref]
			case FieldTypeArray:
				//if f.Type.Ref != "" {
				//}
				if ref, ok := mapping[f.Type.Ref]; ok {
					scm.Fields[i].Type.Name = "[]" + ref
					scm.Fields[i].Type.Ref = ref
				}
			case FieldTypeMap:
				//if f.Type.Ref != "" {
				if ref, ok := mapping[f.Type.Ref]; ok {
					scm.Fields[i].Type.Ref = ref
					scm.Fields[i].Type.Name = fmt.Sprintf("map[%s]%s", f.Type.Key, ref)
				}
			}
			scms[mapping[k]] = scm
		}
	}
	spec.Schemas = scms

	return spec
}

// 把类似 gitlab.sz.sensetime.com/yuansongxian/backend_scaffold/apispec.Item 这样长的名字
// 换成 apispec.Item 这样的短名字
func schemaNameMapping(scms map[string]Schema) map[string]string {
	nameMap := make(map[string]int)
	mapping := make(map[string]string)
	for k, _ := range scms {
		name := getSchemaName(k)
		tn := name
		for {
			i, ok := nameMap[name]
			if !ok {
				nameMap[name] = 0
				break
			}
			tn = name + strconv.Itoa(i)
			if _, ok := nameMap[tn]; ok {
				nameMap[name]++
			} else {
				nameMap[tn] = 0
				name = tn
				break
			}
		}

		mapping[k] = name
	}
	return mapping
}

func getSchemaName(name string) string {
	if strings.HasPrefix(name, "struct {") {
		// 匿名结构体
		return "DataModel"
	}

	_, tn := filepath.Split(name)
	return tn
}

func ParseStruct(v interface{}, res map[string]schemaX) string {
	objType := reflect.TypeOf(v)
	//objVal  := reflect.ValueOf(v)

	return parseStruct(objType, res)
}

func getTypeName(t reflect.Type) string {
	fn := t.PkgPath() + "." + t.Name()
	if t.PkgPath() == "" {
		fn = t.String()
	}
	return fn
}

func parseStruct(objType reflect.Type, res map[string]schemaX) string {

	pkg := objType.PkgPath()
	pn := getTypeName(objType)
	//pn := getStructName(objType, len(res))

	if _, ok := res[pn]; ok {
		return pn
	}

	t := schemaX{
		idx: len(res),
		pkg: pkg,
		Schema: Schema{
			//idx:       len(res),
			Name: pn,
			//fieldName: fieldName,
			Fields: make([]Field, 0, objType.NumField()),
		},
	}

	innerStructs := make([]y, 0)
	fields := getFields(objType)

	//for i := 0; i < objType.NumField(); i++ {
	for _, field := range fields {
		//field := objType.Field(i) // 获取字段类型

		name := getJsonName(field)
		//typ := getFieldType(field, t.idx)
		typ := getTypeName(field.Type)
		note := getComment(field)
		enum := getEnum(field)

		//pp.Println(field, field.Anonymous)
		//pp.Println(getPkgName(field.Type), field.Type.String())

		if field.Anonymous {
			field.Type.Elem()
		}
		// form表单文件
		if typ == "*multipart.FileHeader" {
			t.Fields = append(t.Fields, Field{
				Name: name,
				Type: FieldType{
					Type: FieldTypeFile,
				},
				Comment: note,
				Enum:    enum,
			})
			continue
		}

		ft, ok := derefType(field.Type)
		if !ok {
			pp.Println("unsupported type", field.Type.Kind().String())
			continue
		}

		if ft.Kind() == reflect.Struct {
			if _, ok := res[getTypeName(ft)]; !ok {
				if !hasType(innerStructs, ft) {
					innerStructs = append(innerStructs, y{t: ft, name: typ})
				}
			}
		}

		fft := FieldType{
			Name: typ,
			Type: typ,
		}
		switch field.Type.Kind() {
		case reflect.Struct:
			fft.Type = FieldTypeObject
			fft.Ref = getTypeName(ft)
		case reflect.Slice:
			fft.Type = FieldTypeArray
			fft.Ref = getTypeName(ft)
		case reflect.Map:
			fft.Type = FieldTypeMap
			fft.Ref = getTypeName(ft)
			fft.Key = field.Type.Key().String()
		}

		t.Fields = append(t.Fields, Field{
			Name:    name,
			Type:    fft,
			Comment: note,
			Enum:    enum,
		})
	}
	res[pn] = t

	for _, t := range innerStructs {
		//pp.Println("inner type", getPkgName(t.t))
		parseStruct(t.t, res)
	}

	return pn
}

// 展开内嵌结构体
func getFields(objType reflect.Type) []reflect.StructField {
	fields := make([]reflect.StructField, 0, objType.NumField())
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i) // 获取字段类型

		if field.Anonymous {
			fields = append(fields, getFields(field.Type)...)
		} else {
			fields = append(fields, field)
		}
	}

	return fields
}

type y struct {
	t    reflect.Type
	name string
}

func getComment(f reflect.StructField) string {
	return f.Tag.Get("comment")
}

func getEnum(f reflect.StructField) string {
	return f.Tag.Get("enum")
}

func getJsonName(f reflect.StructField) string {
	note := f.Tag.Get("json")
	ss := strings.Split(note, ",")
	if ss[0] == "" {
		note = f.Tag.Get("form")
		ss = strings.Split(note, ",")
		if ss[0] == "" {
			note = f.Tag.Get("uri")
			ss = strings.Split(note, ",")
		}
	}
	return ss[0]
}

func derefType(t reflect.Type) (reflect.Type, bool) {
	if normalType[t.Kind()] {
		return t, true
	}

	switch t.Kind() {
	case reflect.Ptr:
		//fmt.Println("*ptf->", t.Elem())
		return derefType(t.Elem())
	case reflect.Struct:
		//fmt.Println("is a struct", t)
		return t, true
	case reflect.Slice, reflect.Array:
		//fmt.Println("[]->", t.Elem())
		return derefType(t.Elem())
	case reflect.Map:
		//fmt.Println("map->", t.Key(), t.Elem())
		if !normalType[t.Key().Kind()] {
			fmt.Println("map key should be normal type")
			return nil, false
		}
		return derefType(t.Elem())
	}

	return nil, false
}

func hasType(arr []y, typ reflect.Type) bool {
	for _, t := range arr {
		if t.t == typ {
			return true
		}
	}
	return false
}

var normalType = map[reflect.Kind]bool{
	reflect.Bool:   true,
	reflect.Int:    true,
	reflect.Int8:   true,
	reflect.Int16:  true,
	reflect.Int32:  true,
	reflect.Int64:  true,
	reflect.Uint:   true,
	reflect.Uint8:  true,
	reflect.Uint16: true,
	reflect.Uint32: true,
	reflect.Uint64: true,
	//reflect.Uintptr: true,
	reflect.Float32: true,
	reflect.Float64: true,
	reflect.String:  true,
}
