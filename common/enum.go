package common

import (
	"fmt"
	"math/rand"
	"strconv"
)

var enumStrings = map[string][]skv{}

func GetEnum(en string) []string {
	ss := getEnum(en)
	if len(ss) > 0 {
		return ss
	}
	return getEnum2(en)
}

func getEnum(en string) []string {
	enums := enumStrings[en]
	ss := make([]string, 0, len(enums))
	for _, en := range enums {
		ss = append(ss, fmt.Sprintf("%s:%s", en.v, en.k))
	}
	return ss
}

func getEnum2(en string) []string {
	enums := allEnums[en]
	ss := make([]string, 0, len(enums))
	for _, en := range enums {
		ss = append(ss, fmt.Sprintf("%s:%s", en.e.Name, en.k))
	}
	return ss
}

func CheckEnum(en string, v string) bool {
	es := enumStrings[en]
	for _, e := range es {
		if e.v == v {
			return true
		}
	}
	return false
}

func RandomEnum(en string) string {
	es := enumStrings[en]
	if len(es) == 0 {
		return ""
	}
	s := es[rand.Intn(len(es))]
	return s.v
}

func RegisterStringEnum(name string, enums []skv) {
	enumStrings[name] = enums
}

type skv struct {
	k string
	v string
}

type Enum struct {
	Name string
	Int  int8
}

var allEnums = map[string][]kEnum{}

func RegisterStringEnum3(name string, enums []kEnum) {
	allEnums[name] = append(allEnums[name], enums...)
}

type kEnum struct {
	k string
	e Enum
}

func GetEnumByName(en string, v string) Enum {
	es := allEnums[en]
	for _, e := range es {
		if e.e.Name == v {
			return e.e
		}
	}
	return Enum{}
}

func GetEnumByInt(en string, v int8) Enum {
	es := allEnums[en]
	for _, e := range es {
		if e.e.Int == v {
			return e.e
		}
	}
	return Enum{}
}

func init() {
	initTaskStatus()
	initTaskType()
}

func initTaskStatus() {
	RegisterStringEnum("TaskStatus", []skv{
		{"等待", strconv.Itoa(TaskStatusPending)},
		{"处理中", strconv.Itoa(TaskStatusProcessing)},
		{"完成", strconv.Itoa(TaskStatusDone)},
		{"取消", strconv.Itoa(TaskStatusCancel)},
	})
}

var (
	TaskTypeImage = Enum{Name: "image", Int: 0}
	TaskTypeVideo = Enum{Name: "video", Int: 1}
)

func initTaskType() {
	RegisterStringEnum3("TaskType", []kEnum{
		{"图片", TaskTypeImage},
		{"视频", TaskTypeVideo},
	})
}
