package apispec

const (
	ReqFormatQuery = "query"
	ReqFormatJson  = "json"
	ReqFormatForm  = "form-data"

	ResponseFormatJson = "json"
	ResponseFormatFile = "file"
)

const (
	FieldTypeFile   = "file"
	FieldTypeArray  = "array"
	FieldTypeObject = "object"
	FieldTypeMap    = "map"
)

type En struct {
	name string
	int  int8
	desc string
}

func (e *En) Name() string {
	return e.name
}
func (e *En) Int() int8 {
	return e.int
}
func (e *En) Desc() string {
	return e.desc
}

func NewEnum(name string, i int8, desc string) EnumItem {
	return &En{name: name, int: i, desc: desc}
}
