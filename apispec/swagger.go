package apispec

import (
	"fmt"
	"strings"

	"github.com/go-openapi/spec"
)

func (a *ApiSpec) ToSwagger() (spec.Swagger, error) {
	root := spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Title:       a.Title,
					Description: a.Description,
					Version:     a.Version,
				},
			},
		},
	}

	root.Definitions = make(spec.Definitions)
	for _, scm := range a.Schemas {
		sch := spec.Schema{}
		for _, f := range scm.Fields {

			fs, ok := getSwaggerNormalTypeSchema(f.Type.Type)
			if !ok {
				switch f.Type.Type {
				case FieldTypeArray:
					itemSchema, ok := getSwaggerNormalTypeSchema(f.Type.Ref)
					if ok {
						fs = spec.ArrayProperty(itemSchema)
					} else {
						// TODO [][]int 这种怎么处理??

						ref := ref2DefinitionsSchemas(f.Type.Ref)
						fs = spec.ArrayProperty(spec.RefSchema(ref))
					}
				case FieldTypeObject:
					fmt.Println("todo object item", f.Name, f.Type.Ref)

					ref := ref2DefinitionsSchemas(f.Type.Ref)
					fs = spec.RefSchema(ref)
				case FieldTypeMap:
					fmt.Println("todo map item", f.Name, f.Type.Ref)
					// TODO
				}
			}

			if fs != nil {
				sch.SetProperty(f.Name, *fs)
			}
		}
		root.Definitions[scm.Name] = sch
	}

	root.Paths = &spec.Paths{
		Paths: make(map[string]spec.PathItem),
	}
	for _, rg := range a.RouteGroups {
		for _, route := range rg.Routes {
			op := spec.Operation{}
			op.Tags = append(op.Tags, rg.Name)
			op.Summary = route.Name
			op.Description = route.Desc
			op.ID = route.ID
			if route.ID == "" {
				op.ID = "api" + getDefaultRouteID(route.Method, route.Uri)
			}

			if route.Request != nil {
				switch route.Request.Format {
				case ReqFormatQuery:
					scm, ok := a.Schemas[route.Request.Schema]
					if !ok {
						return root, fmt.Errorf("schema(%s) not found", route.Request.Schema)
					}
					for _, f := range scm.Fields {
						p := spec.QueryParam(f.Name).
							WithDescription(f.Comment)

						fs, ok := getSwaggerNormalTypeSchema(f.Type.Type)
						if !ok {
							fmt.Println("not normal field in query")
							continue
						}
						p.Typed(fs.Type[0], fs.Format)

						op.AddParam(p)
					}
				case ReqFormatJson:
					ref := ref2DefinitionsSchemas(route.Request.Schema)
					p := spec.BodyParam("", spec.RefSchema(ref))
					op.AddParam(p)
				case ReqFormatForm:
					// TODO
				}
			}

			if route.SuccessResponse.Format == ResponseFormatJson {

				_, ok := root.Definitions[route.SuccessResponse.Schema]
				if !ok {
					return root, fmt.Errorf("schema(%s) not found", route.SuccessResponse.Schema)
				}

				ref := ref2DefinitionsSchemas(route.SuccessResponse.Schema)
				resp := spec.NewResponse().
					WithSchema(spec.RefSchema(ref))

				op.WithDefaultResponse(resp)
			}
			for _, rsp := range route.ErrorResponses {
				_, ok := root.Definitions[rsp.Schema]
				if !ok {
					return root, fmt.Errorf("schema(%s) not found", rsp.Schema)
				}

				ref := ref2DefinitionsSchemas(rsp.Schema)

				resp := spec.NewResponse().
					WithSchema(spec.RefSchema(ref))

				op.RespondsWith(rsp.HttpCode, resp) // TODO
			}

			p, ok := root.Paths.Paths[route.Uri]
			if !ok {
				p = spec.PathItem{}
			}

			switch route.Method {
			case "GET":
				p.Get = &op
			case "PUT":
				p.Put = &op
			case "POST":
				p.Post = &op
			case "DELETE":
				p.Delete = &op
			case "OPTIONS":
				p.Options = &op
			case "HEAD":
				p.Head = &op
			case "PATCH":
				p.Patch = &op
			}
			root.Paths.Paths[route.Uri] = p
		}
	}

	return root, nil
}

func ref2DefinitionsSchemas(s string) string {
	return "#/definitions/" + s
}

func getSwaggerNormalTypeSchema(typ string) (*spec.Schema, bool) {
	switch typ {
	case "string":
		return spec.StringProperty(), true
	case "bool":
		return spec.BooleanProperty(), true
	case "int":
		return spec.Int32Property(), true
	case "int8":
		return spec.Int8Property(), true
	case "int16":
		return spec.Int16Property(), true
	case "int32":
		return spec.Int32Property(), true
	case "int64":
		return spec.Int64Property(), true
	case "float32":
		return spec.Float32Property(), true
	case "float64":
		return spec.Float64Property(), true
	}
	return nil, false
}

func getDefaultRouteID(method string, uri string) string {
	ss := strings.Split(uri, "/")
	for i := range ss {
		ss[i] = strings.Title(Camelcase(ss[i]))
	}
	return strings.Title(strings.ToLower(method)) + strings.Join(ss, "")
}
