package docgen

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"

	"github.com/SeeJson/backend_scaffold/apispec"
)

func ApiSpec2Markdown(spec apispec.ApiSpec) (string, error) {
	sds, err := ApiSpec2ShowDoc(spec)
	if err != nil {
		return "", err
	}

	sort.Slice(sds, func(i, j int) bool {
		return sds[i].SNumber < sds[j].SNumber
	})

	type S struct {
		name string
		ms   []ShowDocModel
	}
	sects := make([]S, 0)
	catMap := make(map[string]int)
	for _, m := range sds {
		i, ok := catMap[m.CatName]
		if !ok {
			i = len(sects)
			catMap[m.CatName] = i
			sects = append(sects, S{name: m.CatName})
		}

		sects[i].ms = append(sects[i].ms, m)
	}

	s := "# " + spec.Title + "\n\n"
	s += "version: " + spec.Version + "\n\n"
	s += spec.Description + "\n\n"
	for _, sect := range sects {
		s += "## " + sect.name + "\n\n"
		for _, m := range sect.ms {
			s += m.PageContent
		}
	}

	return s, nil
}

func ApiSpec2ShowDoc(spec apispec.ApiSpec) (sds []ShowDocModel, err error) {
	sds = make([]ShowDocModel, 0)
	for _, rg := range spec.RouteGroups {
		for i, r := range rg.Routes {
			var bb bytes.Buffer

			param := struct {
				Name        string
				Desc        string
				Uri         string
				Method      string
				Permissions string

				RequestFormat string

				RequestExample  string
				ResponseExample string

				RequestModel  string
				ResponseModel string
			}{
				Name:        r.Name,
				Desc:        r.Desc,
				Uri:         r.Uri,
				Method:      r.Method,
				Permissions: r.Permissions,
			}

			if r.Request != nil {
				param.RequestFormat = r.Request.Format
				if param.RequestFormat == apispec.ResponseFormatJson {
					param.RequestExample = fmt.Sprintf("```\n%s\n```", r.Request.JsonExample)
				}

				param.RequestModel, err = getSchemaModel(spec, r.Request.Schema)
				if err != nil {
					return nil, err
				}
			}
			if r.SuccessResponse.Format == apispec.ResponseFormatJson {
				param.ResponseExample = fmt.Sprintf("```\n%s\n```", r.SuccessResponse.JsonExample)

				param.ResponseModel, err = getSchemaModel(spec, r.SuccessResponse.Schema)
				if err != nil {
					return nil, err
				}
			} else if r.SuccessResponse.Format == apispec.ResponseFormatFile {
				param.ResponseModel = "文件"
			}

			if err := apiTemplate2.Execute(&bb, param); err != nil {
				logrus.WithField("err", err).
					WithField("route", r).
					Error("execute api template error")
			} else {
				sdm := ShowDocModel{
					CatName:     rg.Name,
					PageTitle:   r.Name,
					PageContent: bb.String(),
					SNumber:     i + 1,
				}
				sds = append(sds, sdm)
			}
		}
	}

	ecHeacer := []string{"错误码", "错误信息", "备注"}
	rows := make([][]string, 0, 3)
	for _, ec := range spec.ErrorCodes {
		rows = append(rows, []string{
			ec.Code,
			ec.Msg,
			ec.Desc,
		})
	}
	tbl := genMDTable(ecHeacer, rows)
	content := fmt.Sprintf(`## 错误码列表

%s

`, tbl)

	sds = append(sds, ShowDocModel{
		CatName:     "错误码",
		PageTitle:   "错误码列表",
		PageContent: content,
		SNumber:     999,
	})

	return sds, nil
}

func getSchemaModel(spec apispec.ApiSpec, schema string) (string, error) {
	set := make(map[string]struct{})
	ss, err := getSchemaTable(spec, schema, set)
	if err != nil {
		return "", err
	}
	return strings.Join(ss, "\n\n"), nil
}

func formatField(spec apispec.ApiSpec, field apispec.Field) []string {
	typ := field.Type.Name
	if field.Type.Type == apispec.FieldTypeFile {
		typ = "表单文件"
	}
	comment := field.Comment
	if field.Enum != "" {
		comment += fmt.Sprintf(" 枚举类型[%s]", strings.Join(getEnum(spec, field.Enum), " "))
	}

	row := []string{
		field.Name,
		typ,
		comment,
	}

	return row
}

func getEnum(spec apispec.ApiSpec, ent string) []string {
	ens := spec.Enums[ent]

	ss := make([]string, 0, len(ens))
	for _, en := range ens {
		ss = append(ss, fmt.Sprintf("%s:%s", en.Name, en.Desc))
	}

	return ss
}

func getSchemaTable(spec apispec.ApiSpec, schema string, handled map[string]struct{}) ([]string, error) {
	if _, ok := handled[schema]; ok {
		return nil, nil
	}

	rr := make([]string, 0, 2)

	sch, ok := spec.Schemas[schema]
	if !ok {
		return nil, fmt.Errorf("schema:%s not found", schema)
	}

	otherSchemas := make(map[string]struct{})

	rows := make([][]string, 0, len(sch.Fields))
	for _, field := range sch.Fields {
		rows = append(rows, formatField(spec, field))
		if _, ok := spec.Schemas[field.Type.Ref]; ok {
			otherSchemas[field.Type.Ref] = struct{}{}
		}
	}

	md := genMDTable(header, rows)
	// type name
	_, tn := filepath.Split(schema)
	if strings.HasPrefix(tn, "struct {") {
		// 匿名结构体
		//tn = t.fieldName
	}
	// TODO
	//if i == 0 {
	//	tn = ""
	//}

	//r[tn] = md
	rr = append(rr, fmt.Sprintf("%s\n\n%s", tn, md))

	handled[schema] = struct{}{}
	for s := range otherSchemas {
		rs, err := getSchemaTable(spec, s, handled)
		if err != nil {
			return nil, err
		}
		rr = append(rr, rs...)
	}

	return rr, nil
}

var apiTemplate2 = template.Must(template.New("api").Parse(`### {{ .Name }}

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
