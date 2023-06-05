package backend

import (
	"github.com/SeeJson/backend_scaffold/docgen"
)

const (
	sectionUser  = "用户"
	sectionTask  = "任务"
	sectionOther = "其它"
)

var sections = map[string]docgen.DocSection{
	sectionUser:  {1, sectionUser},
	sectionTask:  {2, sectionTask},
	sectionOther: {8, sectionOther},
}

func (s *Server) setupRouteDoc(method, uri string, roles uint64, info *docgen.DocInfo) {
	if info != nil {
		info.Roles = roles
		info.Method = method
		info.Uri = uri
		switch roles {
		default:
			info.Permissions = "无"
		}

		s.docs = append(s.docs, info)
	}
}
