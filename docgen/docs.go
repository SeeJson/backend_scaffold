package docgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/k0kubun/pp"
	"github.com/sirupsen/logrus"
)

type DocSection struct {
	Index int
	Name  string
}

const (
	DocInfoReqFormatQuery = "query"
	DocInfoReqFormatJson  = "json"
	DocInfoReqFormatForm  = "form-data"

	DocInfoResponseFormatJson = "json"
	DocInfoResponseFormatFile = "file"
)

type DocInfo struct {
	Index   int
	Name    string
	Section string
	Desc    string

	Method        string
	Uri           string
	Roles         uint64
	RequestFormat string

	Permissions string

	Param interface{}
	Req   interface{}
	Rsp   interface{}
}

type DocGenerator struct {
	Title      string
	Dir        string
	Sections   map[string]DocSection
	EnumGetter func(string) []string

	sds []ShowDocModel
}

func (g *DocGenerator) GenDoc(outPath string, docs []*DocInfo) {
	docFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		logrus.WithField("err", err).Error("open doc file fail")
		return
	}

	ms := make(map[string][]*DocInfo)
	for i := range docs {
		if docs[i].Method == "GET" {
			docs[i].RequestFormat = DocInfoReqFormatQuery
		} else if docs[i].RequestFormat == "" {
			docs[i].RequestFormat = DocInfoReqFormatJson
		}
		ms[docs[i].Section] = append(ms[docs[i].Section], docs[i])
	}

	type sect struct {
		section DocSection
		docs    []*DocInfo
	}
	sects := make([]sect, 0, len(ms))
	for k, v := range ms {
		sort.Slice(v, func(i, j int) bool {
			return v[i].Index < v[j].Index
		})
		sects = append(sects, sect{
			section: g.Sections[k],
			docs:    v,
		})
	}
	sort.Slice(sects, func(i, j int) bool {
		return sects[i].section.Index < sects[j].section.Index
	})

	docFile.WriteString(fmt.Sprintf("# %s\n\n", g.Title))

	for _, sect := range sects {
		docFile.WriteString(g.formatSect(sect.section.Name, sect.docs))
		docFile.WriteString("\n\n")
	}

	//pp.Println(g.sds)
	bb, _ := json.MarshalIndent(g.sds, "", "  ")
	ioutil.WriteFile("showdoc_data.json", bb, 0766)

	docFile.Sync()
	docFile.Close()
}

var apiTemplate = template.Must(template.New("api").Parse(`### {{ .Name }}

{{ .Desc }}

#### 请求URL:
- {{ .Uri }}

#### 请求方式:
- {{ .Method }}

#### 请求格式:
- {{ .RequestFormat }}

#### 所需权限:
- {{ .Permissions }}

#### 请求参数:

{{ .RequestModel }}

#### 请求示例:

{{ .RequestExample }}

#### 返回参数:

{{ .ResponseModel }}

#### 返回示例:

{{ .ResponseExample }}

------
`))

func (g *DocGenerator) formatSect(name string, docs []*DocInfo) string {
	s := "## " + name + "\n\n"

	for i, doc := range docs {
		//pp.Println(doc)
		var bb bytes.Buffer

		param := struct {
			DocInfo

			RequestExample  string
			ResponseExample string

			RequestModel  string
			ResponseModel string
		}{
			DocInfo: *doc,
		}

		if doc.Req != nil {
			buf, err := json.MarshalIndent(doc.Req, "", "  ")
			if err != nil {
				logrus.WithField("err", err).
					WithField("doc", doc).
					Error("req marshal error")
			}

			if doc.Method != "GET" && doc.RequestFormat != DocInfoReqFormatForm {
				param.RequestExample = "```\n" + string(buf) + "\n```"
			}

			ss := g.ParseModel(doc.Req)
			param.RequestModel = strings.Join(ss, "\n\n")
		}

		if doc.Rsp != nil {
			buf, err := json.MarshalIndent(doc.Rsp, "", "  ")
			if err != nil {
				logrus.WithField("err", err).
					WithField("doc", doc).
					Error("rsp marshal error")
			}
			param.ResponseExample = "```\n" + string(buf) + "\n```\n"

			ss := g.ParseModel(doc.Rsp)
			param.ResponseModel = strings.Join(ss, "\n\n")
		}

		//pp.Println(param)

		if err := apiTemplate.Execute(&bb, param); err != nil {
			logrus.WithField("err", err).
				WithField("doc", doc).
				Error("execute api template error")
		}

		if g.Dir != "" {
			sdm := ShowDocModel{
				CatName:     name,
				PageTitle:   doc.Name,
				PageContent: bb.String(),
				SNumber:     i,
			}
			g.sds = append(g.sds, sdm)
		}

		s += bb.String()
	}

	return s
}

type ShowDocModel struct {
	CatName     string `json:"cat_name"`
	PageTitle   string `json:"page_title"`
	PageContent string `json:"page_content"`
	SNumber     int    `json:"s_number"`
}

var header = []string{"字段名", "类型", "备注"}

func (g *DocGenerator) ParseModel(v interface{}) []string {

	tables := make(map[string]X)
	g.ParseStruct(v, tables)
	//pp.Println(tables)

	ts := make([]X, 0, len(tables))
	for _, t := range tables {
		ts = append(ts, t)
	}
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].idx < ts[j].idx
	})

	r := make([]string, 0, len(ts))
	for i, t := range ts {
		md := genMDTable(header, t.data)
		// type name
		_, tn := filepath.Split(t.name)
		if strings.HasPrefix(tn, "struct {") {
			// 匿名结构体
			tn = t.fieldName
		}
		if i == 0 {
			tn = ""
		}

		//r[tn] = md
		r = append(r, fmt.Sprintf("%s\n\n%s", tn, md))

		//println(tn)
		//println(md)
		//println("")
	}
	return r
}

type X struct {
	idx       int
	name      string
	fieldName string
	data      [][]string
}

type y struct {
	t    reflect.Type
	name string
}

func (g *DocGenerator) ParseStruct(v interface{}, res map[string]X) {
	objType := reflect.TypeOf(v)
	//objVal  := reflect.ValueOf(v)

	g.parseStruct("", objType, res)
}

func (g *DocGenerator) parseStruct(fieldName string, objType reflect.Type, res map[string]X) {

	pn := getPkgName(objType)
	if _, ok := res[pn]; ok {
		return
	}

	t := X{
		idx:       len(res),
		name:      pn,
		fieldName: fieldName,
		data:      make([][]string, 0, objType.NumField()),
	}
	//if strings.HasPrefix(pn, "struct {") {
	//	t.name = fmt.Sprintf()
	//}

	innerStructs := make([]y, 0)
	fields := getFields(objType)

	//for i := 0; i < objType.NumField(); i++ {
	for _, field := range fields {
		//field := objType.Field(i) // 获取字段类型

		name := getJsonName(field)
		typ := getFieldType(field, t.idx)
		note := g.getNote(field)

		//pp.Println(field, field.Anonymous)
		//pp.Println(getPkgName(field.Type), field.Type.String())

		if field.Anonymous {
			field.Type.Elem()
		}
		// form表单文件
		if typ == "*multipart.FileHeader" {
			typ = "表单文件"
			t.data = append(t.data, []string{name, typ, note})
			continue
		}

		ft, ok := derefType(field.Type)
		if !ok {
			pp.Println("unsupported type", field.Type.Kind().String())
			continue
		}

		if ft.Kind() == reflect.Struct {
			if _, ok := res[getPkgName(ft)]; !ok {
				if !hasType(innerStructs, ft) {
					innerStructs = append(innerStructs, y{t: ft, name: typ})
				}
			}
		}

		t.data = append(t.data, []string{name, typ, note})
	}
	res[pn] = t

	for _, t := range innerStructs {
		//pp.Println("inner type", getPkgName(t.t))
		g.parseStruct(t.name, t.t, res)
	}
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

func getPkgName0(t reflect.Type) string {
	return t.PkgPath() + "." + t.Name()
}

func getPkgName(t reflect.Type) string {
	//return t.PkgPath() + "." + t.Name()
	fn := getPkgName0(t)
	if fn == "." {
		return t.String()
	}
	return fn
}

func getFieldType(f reflect.StructField, idx int) string {
	//fn := getPkgName0(f.Type)
	//if fn == "." {
	//	pp.Println("xxxxx", fn, f.Type.PkgPath(), f.Type.Name())
	//	return f.Name+strconv.Itoa(idx)
	//}

	ftn := f.Type.String()
	if strings.HasPrefix(ftn, "struct {") {
		return f.Name + strconv.Itoa(idx)
	}
	return ftn
}

func getJsonName(f reflect.StructField) string {
	note := f.Tag.Get("json")
	ss := strings.Split(note, ",")
	if ss[0] == "" {
		note = f.Tag.Get("form")
		ss = strings.Split(note, ",")
	}
	return ss[0]
}

func (g *DocGenerator) getNote(f reflect.StructField) string {
	note := f.Tag.Get("comment")
	//ss := strings.Split(note, ",")
	//n := ss[0]
	n := note
	//return n

	ek := f.Tag.Get("enum")
	if ek == "" {
		return n
	}

	return n + fmt.Sprintf(" 枚举类型[%s]", strings.Join(g.EnumGetter(ek), " "))
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

func genMDTable(header []string, records [][]string) string {
	//maxLen := 0
	columns := len(header)
	for _, rec := range records {
		if len(rec) > columns {
			columns = len(rec)
		}
	}
	md := ``
	for len(header) < columns {
		header = append(header, "")
	}
	md += fmt.Sprintf("|%s|\n|", strings.Join(header, "|"))
	filler := "---"
	for i := 0; i < columns; i++ {
		md += filler + "|"
	}
	md += "\n"

	for _, rec := range records {
		for len(rec) < columns {
			rec = append(rec, "")
		}
		md += fmt.Sprintf("|%s|\n", strings.Join(rec, "|"))
	}
	return md
}
