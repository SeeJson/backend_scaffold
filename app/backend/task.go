package backend

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
	"gorm.io/gorm"

	"github.com/SeeJson/backend_scaffold/app"
	"github.com/SeeJson/backend_scaffold/cerror"
	"github.com/SeeJson/backend_scaffold/common"
	"github.com/SeeJson/backend_scaffold/docgen"
	"github.com/SeeJson/backend_scaffold/ginplus"
	"github.com/SeeJson/backend_scaffold/model"
	"github.com/SeeJson/backend_scaffold/utils"
)

type NewTaskRequest struct {
	Type     string `json:"type" comment:"任务类型" enum:"TaskType"`
	Priority int8   `json:"priority" comment:"任务优先级"`
	Input    string `json:"input" comment:"任务输入"`
	Note     string `json:"note" comment:"备注"`
}

func (s *Server) NewTaskHandler() Handler {
	return Handler{
		doc: &docgen.DocInfo{
			Section: sectionTask,
			Index:   21,
			Name:    "新增任务",
			Desc:    "新增任务",
			Req:     NewTaskRequest{},
			Rsp:     ginplus.OkResponse,
		},
		handler: func(c *gin.Context) {
			sess := GetSession(c)
			if sess == nil {
				ginplus.ResError(c, cerror.ErrUnauthorized)
				return
			}
			logger := app.GetLogger(c).WithField("uid", sess.Uid)

			var req NewTaskRequest
			err := c.ShouldBindJSON(&req)
			if err != nil {
				logger.Errorf("fail to bind request params: %v", err)
				ginplus.ResError(c, cerror.ErrInvalidRequest)
				return
			}
			logger = logger.WithField("req", req)

			if req.Input == "" {
				ginplus.ResError(c, cerror.ErrInvalidRequest, "空输入")
				return
			}
			taskType := common.GetEnumByName("TaskType", req.Type)
			if taskType.Name == "" {
				ginplus.ResError(c, cerror.ErrInvalidRequest, "非法任务类型")
				return
			}

			now := time.Now()
			task := model.Task{
				Uid:      sess.Uid,
				Status:   common.TaskStatusPending,
				Type:     taskType.Int,
				Input:    req.Input,
				Priority: req.Priority,
				Note:     req.Note,
				CT:       now,
				UT:       now,
			}

			err = s.db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Create(&task).Error; err != nil {
					logger.WithField("err", err).Error("create address error")
					return err
				}
				return nil
			})
			if err != nil {
				ginplus.ResError(c, err)
				return
			}

			logger.Info("new task ok")
			ginplus.ResOK(c)
			s.opWriter.WriteOpLog(c, model.OperationLog{
				Uid:       sess.Uid,
				API:       c.Request.URL.String(),
				Operation: "new_task",
				Value:     utils.MarshalJSON(task),
			})
		},
	}
}

type TaskDetailResponse struct {
	ginplus.CommonResponse

	Data TaskDetail `json:"data"`
}

type TaskDetail struct {
	ID       int64  `json:"id"`
	Status   uint8  `json:"status" comment:"任务状态" enum:"TaskStatus"`
	Type     string `json:"type" comment:"任务类型" enum:"TaskType"`
	Priority int8   `json:"priority" comment:"任务优先级"`
	Input    string `json:"input" comment:"任务输入"`
	Result   string `json:"result" comment:"任务结果"`
	Note     string `json:"note" comment:"备注"`
}

func (s *Server) TaskDetailHandler() Handler {
	return Handler{
		doc: &docgen.DocInfo{
			Section: sectionTask,
			Index:   22,
			Name:    "任务详情",
			Desc:    "任务详情",
			Req:     ginplus.IDRequest{},
			Rsp:     TaskDetailResponse{},
		},
		handler: func(c *gin.Context) {
			sess := GetSession(c)
			if sess == nil {
				ginplus.ResError(c, cerror.ErrUnauthorized)
				return
			}
			logger := app.GetLogger(c).WithField("uid", sess.Uid)

			var req ginplus.IDRequest
			err := c.ShouldBindQuery(&req)
			if err != nil {
				logger.Errorf("fail to bind request params: %v", err)
				ginplus.ResError(c, cerror.ErrInvalidRequest)
				return
			}
			logger = logger.WithField("req", req)

			task, err := model.GetTask(s.db, req.ID)
			if err != nil {
				logger.WithField("err", err).Error("GetTask error")
				ginplus.ResError(c, app.MapErr(err))
				return
			}
			if task.Uid != sess.Uid {
				ginplus.ResError(c, cerror.ErrNoPermission)
				return
			}

			resp := TaskDetailResponse{
				CommonResponse: ginplus.OkResponse,
				Data: TaskDetail{
					ID:       task.ID,
					Status:   task.Status,
					Type:     common.GetEnumByInt("TaskType", task.Type).Name,
					Input:    task.Input,
					Result:   task.Result,
					Note:     task.Note,
					Priority: task.Priority,
				},
			}

			ginplus.ResSuccess(c, resp)
		},
	}
}

type TaskDetailQuery struct {
	Field string `query:"field"`
}

type TaskDetailParam struct {
	ID int64 `json:"id" form:"id" uri:"id"`
}

func (s *Server) TaskDetail2Handler() Handler {
	return Handler{
		doc: &docgen.DocInfo{
			Section: sectionTask,
			Index:   22,
			Name:    "任务详情",
			Desc:    "任务详情",
			Param:   TaskDetailParam{},
			Req:     TaskDetailQuery{},
			Rsp:     TaskDetailResponse{},
		},
		handler: func(c *gin.Context) {
			sess := GetSession(c)
			if sess == nil {
				ginplus.ResError(c, cerror.ErrUnauthorized)
				return
			}
			logger := app.GetLogger(c).WithField("uid", sess.Uid)

			var param TaskDetailParam
			err := c.ShouldBindUri(&param)
			if err != nil {
				logger.Errorf("fail to bind request params: %v", err)
				ginplus.ResError(c, cerror.ErrInvalidRequest)
				return
			}
			logger = logger.WithField("req", param)

			pp.Println(param)

			var req2 TaskDetailQuery
			err = c.ShouldBindQuery(&req2)
			pp.Println(req2, err)

			task, err := model.GetTask(s.db, param.ID)
			if err != nil {
				logger.WithField("err", err).Error("GetTask error")
				ginplus.ResError(c, app.MapErr(err))
				return
			}
			if task.Uid != sess.Uid {
				ginplus.ResError(c, cerror.ErrNoPermission)
				return
			}

			resp := TaskDetailResponse{
				CommonResponse: ginplus.OkResponse,
				Data: TaskDetail{
					ID:       task.ID,
					Status:   task.Status,
					Type:     common.GetEnumByInt("TaskType", task.Type).Name,
					Input:    task.Input,
					Result:   task.Result,
					Note:     task.Note,
					Priority: task.Priority,
				},
			}

			ginplus.ResSuccess(c, resp)
		},
	}
}

type UpdateTaskRequest struct {
	ID       int64  `json:"id"`
	Note     string `json:"note" comment:"备注"`
	Priority int8   `json:"priority" comment:"任务优先级"`
}

func (s *Server) UpdateTaskHandler() Handler {
	return Handler{
		doc: &docgen.DocInfo{
			Section: sectionTask,
			Index:   22,
			Name:    "更新任务信息",
			Desc:    "更新任务信息",
			Req:     UpdateTaskRequest{},
			Rsp:     ginplus.OkResponse,
		},
		handler: func(c *gin.Context) {
			sess := GetSession(c)
			if sess == nil {
				ginplus.ResError(c, cerror.ErrUnauthorized)
				return
			}
			logger := app.GetLogger(c).WithField("uid", sess.Uid)

			var req UpdateTaskRequest
			err := c.ShouldBindJSON(&req)
			if err != nil {
				logger.Errorf("fail to bind request params: %v", err)
				ginplus.ResError(c, cerror.ErrInvalidRequest)
				return
			}
			logger = logger.WithField("req", req)

			update := make(map[string]interface{}, 4)
			if req.Priority != 0 {
				update["priority"] = req.Priority
			}
			if req.Note != "" {
				update["note"] = req.Note
			}

			if len(update) == 0 {
				ginplus.ResOK(c)
				return
			}

			if err := model.UpdateTask(s.db, req.ID, update); err != nil {
				logger.WithField("err", err).Error("update task status error")
				ginplus.ResError(c, err)
				return
			}

			logger.Info("update task ok")
			ginplus.ResOK(c)
			s.opWriter.WriteOpLog(c, model.OperationLog{
				Uid:       sess.Uid,
				API:       c.Request.URL.String(),
				Operation: "update_task",
				Value:     utils.MarshalJSON(req),
			})
		},
	}
}

type UpdateTaskStatusRequest struct {
	ID     int64 `json:"id"`
	Status uint8 `json:"status" comment:"任务状态" enum:"TaskStatus"`
}

func (s *Server) UpdateTaskStatusHandler() Handler {
	return Handler{
		doc: &docgen.DocInfo{
			Section: sectionTask,
			Index:   22,
			Name:    "更新任务状态",
			Desc:    "更新任务状态",
			Req:     UpdateTaskStatusRequest{},
			Rsp:     ginplus.OkResponse,
		},
		handler: func(c *gin.Context) {
			sess := GetSession(c)
			if sess == nil {
				ginplus.ResError(c, cerror.ErrUnauthorized)
				return
			}
			logger := app.GetLogger(c).WithField("uid", sess.Uid)

			var req UpdateTaskStatusRequest
			err := c.ShouldBindJSON(&req)
			if err != nil {
				logger.Errorf("fail to bind request params: %v", err)
				ginplus.ResError(c, cerror.ErrInvalidRequest)
				return
			}
			logger = logger.WithField("req", req)

			var errMsg string
			txErr := s.db.Transaction(func(tx *gorm.DB) error {
				task, err := model.GetTask(tx, req.ID)
				if err != nil {
					logger.WithField("err", err).Error("GetTask error")
					return err
				}
				if task.Status == common.TaskStatusCancel {
					errMsg = ""
					return cerror.ErrInvalidStatus
				}
				if task.Status == req.Status {
					logger.Infof("nothing changed")
					return nil
				}

				switch task.Status {
				case common.TaskStatusProcessing:
					if common.TaskStatusPending == req.Status {
						errMsg = "不能回退回pending状态"
						return cerror.ErrInvalidStatus
					}
				case common.TaskStatusDone:
					errMsg = "不能修改已完成的任务状态"
					return cerror.ErrInvalidStatus
				case common.TaskStatusCancel:
					errMsg = "不能修改已取消的任务状态"
					return cerror.ErrInvalidStatus
				}

				task.Status = req.Status
				if err := tx.Save(task).Error; err != nil {
					logger.WithField("err", err).Error("update task status error")
					return err
				}

				return nil
			})
			if txErr != nil {
				ginplus.ResError(c, app.MapErr(txErr), errMsg)
				return
			}

			logger.Info("update task status ok")
			ginplus.ResOK(c)
			s.opWriter.WriteOpLog(c, model.OperationLog{
				Uid:       sess.Uid,
				API:       c.Request.URL.String(),
				Operation: "update_task_status",
				Value:     utils.MarshalJSON(req),
			})
		},
	}
}

func (s *Server) DeleteTaskHandler() Handler {
	return Handler{
		doc: &docgen.DocInfo{
			Section: sectionTask,
			Index:   23,
			Name:    "删除任务",
			Desc:    "删除任务",
			Req:     ginplus.IDsRequest{},
			Rsp:     ginplus.OkResponse,
		},
		handler: func(c *gin.Context) {
			sess := GetSession(c)
			if sess == nil {
				ginplus.ResError(c, cerror.ErrUnauthorized)
				return
			}
			logger := app.GetLogger(c).WithField("uid", sess.Uid)

			var req ginplus.IDsRequest
			err := c.ShouldBindJSON(&req)
			if err != nil {
				logger.Errorf("fail to bind request params: %v", err)
				ginplus.ResError(c, cerror.ErrInvalidRequest)
				return
			}
			logger = logger.WithField("req", req)

			if len(req.IDs) == 0 {
				ginplus.ResOK(c)
				return
			}

			err = s.db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Unscoped().Where("id in (?) and uid=?", req.IDs, sess.Uid).Delete(model.Task{}).Error; err != nil {
					logger.WithField("err", err).Error("delete tasks error")
					return err
				}
				return nil
			})
			if err != nil {
				ginplus.ResError(c, err)
				return
			}

			logger.Info("delete tasks ok")
			ginplus.ResOK(c)
			s.opWriter.WriteOpLog(c, model.OperationLog{
				Uid:       sess.Uid,
				API:       c.Request.URL.String(),
				Operation: "delete_tasks",
				Value:     utils.MarshalJSON(req),
			})
		},
	}
}

type ListTaskResponse struct {
	ginplus.CommonResponse

	Total int64      `json:"total"`
	Items []TaskItem `json:"items"`
}

type TaskItem struct {
	ID       int64  `json:"id"`
	Status   uint8  `json:"status" comment:"任务状态" enum:"TaskStatus"`
	Type     string `json:"type" comment:"任务类型" enum:"TaskType"`
	Priority int8   `json:"priority" comment:"任务优先级"`
	Note     string `json:"note" comment:"备注"`

	User UidName `json:"user" comment:"用户信息"`
}

type UidName struct {
	Uid  int64  `json:"uid" comment:"用户名"`
	Name string `json:"name" comment:"用户名"`
}

func (s *Server) ListAddressHandler() Handler {
	return Handler{
		doc: &docgen.DocInfo{
			Section: sectionTask,
			Index:   24,
			Name:    "任务列表",
			Desc:    "任务列表",
			Req:     ginplus.Pager{},
			Rsp: ListTaskResponse{
				Items: make([]TaskItem, 2),
			},
		},
		handler: func(c *gin.Context) {
			sess := GetSession(c)
			if sess == nil {
				ginplus.ResError(c, cerror.ErrUnauthorized)
				return
			}
			logger := app.GetLogger(c).WithField("uid", sess.Uid)

			var req ginplus.Pager
			err := c.ShouldBindQuery(&req)
			if err != nil {
				logger.Errorf("fail to bind request params: %v", err)
				ginplus.ResError(c, cerror.ErrInvalidRequest)
				return
			}
			logger = logger.WithField("req", req)

			resp := ListTaskResponse{
				CommonResponse: ginplus.OkResponse,
				Items:          make([]TaskItem, 0, req.Limit),
			}
			if req.Limit == 0 {
				ginplus.ResSuccess(c, resp)
				return
			}

			type X struct {
				model.Task

				UName string `gorm:"column:uname"`
			}

			tasks := make([]X, 0, req.Limit)
			db := s.db.Where("tasks.uid=?", sess.Uid)
			db = db.Joins("inner join users on users.uid=tasks.uid")
			qdb := db.WithContext(c).Limit(req.Limit).Offset(req.Offset)
			qdb = qdb.Select("tasks.*, users.name as uname")
			if err := qdb.Find(&tasks).Error; err != nil {
				logger.WithField("err", err).Error("query user tasks error")
				ginplus.ResError(c, app.MapErr(err))
				return
			}
			for _, task := range tasks {
				item := TaskItem{
					ID:       task.ID,
					Priority: task.Priority,
					Note:     task.Note,
					Type:     common.GetEnumByInt("TaskType", task.Type).Name,
					Status:   task.Status,
					User: UidName{
						Uid:  task.Uid,
						Name: task.UName,
					},
				}
				resp.Items = append(resp.Items, item)
			}
			sdb := db.Model(model.Task{}).Select("count(1)")
			if err := sdb.Count(&resp.Total).Error; err != nil {
				logger.WithField("err", err).Error("count user tasks error")
			}

			ginplus.ResSuccess(c, resp)
		},
	}
}
