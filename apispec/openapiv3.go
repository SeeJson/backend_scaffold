package apispec

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

func (a *ApiSpec) ToOpenApiV3() (openapi3.T, error) {
	root := openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:       a.Title,
			Description: a.Description,
			Version:     a.Version,
		},
	}

	root.Components = openapi3.NewComponents()
	root.Components.Schemas = make(openapi3.Schemas)
	for _, scm := range a.Schemas {
		sch := openapi3.NewObjectSchema()
		for _, f := range scm.Fields {

			fs, ok := getNormalTypeSchema(f.Type.Type)
			if !ok {
				switch f.Type.Type {
				case FieldTypeArray:
					itemSchema, ok := getNormalTypeSchema(f.Type.Ref)
					if ok {
						fs = openapi3.NewArraySchema().WithItems(itemSchema)
					} else {
						// TODO [][]int 这种怎么处理??

						ref := ref2ComponentsSchemas(f.Type.Ref)
						fs = openapi3.NewArraySchema()
						fs.Items = openapi3.NewSchemaRef(ref, openapi3.NewObjectSchema())
					}
				case FieldTypeObject:
					fmt.Println("todo object item", f.Name, f.Type.Ref)

					ref := ref2ComponentsSchemas(f.Type.Ref)
					sch.WithPropertyRef(f.Name, openapi3.NewSchemaRef(ref, openapi3.NewObjectSchema()))
				case FieldTypeMap:
					// TODO
				}
			}

			if fs != nil {
				sch.WithPropertyRef(f.Name, fs.NewRef())
			}
		}

		root.Components.Schemas[scm.Name] = openapi3.NewSchemaRef("", sch)
	}

	for _, rg := range a.RouteGroups {
		for _, route := range rg.Routes {
			op := openapi3.NewOperation()
			op.Tags = append(op.Tags, rg.Name)
			op.Summary = route.Name
			op.Description = route.Desc

			if route.Request != nil {
				switch route.Request.Format {
				case ReqFormatQuery:
					scm, ok := a.Schemas[route.Request.Schema]
					if !ok {
						return root, fmt.Errorf("schema(%s) not found", route.Request.Schema)
					}
					for _, f := range scm.Fields {
						p := openapi3.NewQueryParameter(f.Name).
							WithDescription(f.Comment)
						fs, ok := getNormalTypeSchema(f.Type.Type)
						if !ok {
							fmt.Println("not normal field in query")
							continue
						}
						p.WithSchema(fs)
						op.AddParameter(p)
					}
				case ReqFormatJson:
					rb := openapi3.NewRequestBody().WithRequired(true)

					ref := ref2ComponentsSchemas(route.Request.Schema)
					rb.WithJSONSchemaRef(openapi3.NewSchemaRef(ref, openapi3.NewObjectSchema()))
					op.RequestBody = &openapi3.RequestBodyRef{"", rb}
				case ReqFormatForm:
					// TODO
				}
			}
			if route.SuccessResponse.Format == ResponseFormatJson {

				sch, ok := root.Components.Schemas[route.SuccessResponse.Schema]
				if !ok {
					return root, fmt.Errorf("schema(%s) not found", route.SuccessResponse.Schema)
				}

				rb := openapi3.NewResponse().WithDescription("success response")
				ref := ref2ComponentsSchemas(route.SuccessResponse.Schema)
				rb.WithJSONSchemaRef(openapi3.NewSchemaRef(ref, sch.Value))

				op.AddResponse(200, rb)
			}
			for _, rsp := range route.ErrorResponses {

				sch, ok := root.Components.Schemas[rsp.Schema]
				if !ok {
					return root, fmt.Errorf("schema(%s) not found", rsp.Schema)
				}

				rb := openapi3.NewResponse().WithDescription("error response")
				ref := ref2ComponentsSchemas(rsp.Schema)
				rb.WithJSONSchemaRef(openapi3.NewSchemaRef(ref, sch.Value))

				op.AddResponse(rsp.HttpCode, rb) // TODO
			}

			root.AddOperation(route.Uri, route.Method, op)
		}
	}
	if err := root.Validate(context.Background()); err != nil {
		return root, err
	}

	return root, nil
}

func ref2ComponentsSchemas(s string) string {
	return "#/components/schemas/" + s
}

func getNormalTypeSchema(typ string) (*openapi3.Schema, bool) {
	switch typ {
	case "string":
		return openapi3.NewStringSchema(), true
	case "bool":
		return openapi3.NewBoolSchema(), true
	case "int":
		return openapi3.NewIntegerSchema(), true
	case "int8":
		return openapi3.NewIntegerSchema(), true
	case "int16":
		return openapi3.NewIntegerSchema(), true
	case "int32":
		return openapi3.NewInt32Schema(), true
	case "int64":
		return openapi3.NewInt64Schema(), true
	case "float64", "float32":
		return openapi3.NewFloat64Schema(), true
	}
	return nil, false
}
