
# 后端GO开发脚手架

本项目是根据以前做过的项目, copy一些有用的代码拼凑出来的一套web后台开发框架.

主要适用于快速开发中小型的单体web项目.


1. 项目架构

```
.
├── app
│   ├── backend  // 具体服务的server业务实现
│   ├── logger.go
│   ├── middleware.go
│   ├── oplog.go
│   ├── upload.go
│   └── utils.go
├── cache
├── cerror  // 错误定义
├── cmd
│   └── backend // main入口
├── common // common数据定义
├── docgen
├── ginplus // common数据定义
├── model
├── sql  // 数据库表定义及初始化数据
│   ├── data.sql
│   ├── ddl.sql
├── tools
│   └── post_api // api文档更新工具
├── upload
├── utils // 一些工具
└── version

```

# 2. 文档生成

为了懒得写api文档, 所以用反射撸了一套自动生成api文档的工具出来, 可以配合showdoc(https://www.showdoc.com.cn/) 展示.

ps: 之前试过 https://github.com/swaggo/swag , 感觉还是不太灵活, 有些需求还是不能满足. 

## 2.1 写代码 

定义一个api需要写一些额外的代码来
```
func (s *Server) UpdateTaskStatusHandler() Handler {
	return Handler{
		doc: &docgen.DocInfo{
			Section: sectionTask,    // 文档的章节
			Index:   22,             // 文档里的排序, 越大越后
			Name:    "更新任务状态",  // 文档里面的title
			Desc:    "更新任务状态",  // 文档里面的介绍文字
			Req:     UpdateTaskStatusRequest{}, // 请求的定义
			Rsp:     ginplus.OkResponse,        // 返回的定义
		},
		handler: func(c *gin.Context) {
            // 下面写业务代码
            ....

            // 解析请求都这样写, request类型跟上面的Req要一样 
			var req UpdateTaskStatusRequest
			err := c.ShouldBindJSON(&req)
			if err != nil {
				logger.Errorf("fail to bind request params: %v", err)
				ginplus.ResError(c, cerror.ErrInvalidRequest)
				return
			}
        },
    }
}
```

server.go  里面初始化路由
```

func (s *Server) Run(addr string) error {
        ...

		s.setupRoute(gr, "GET", "/tasks", 0, s.ListAddressHandler())
		s.setupRoute(gr, "GET", "/task", 0, s.TaskDetailHandler())
		s.setupRoute(gr, "POST", "/task", 0, s.NewTaskHandler())
		s.setupRoute(gr, "PUT", "/task", 0, s.UpdateTaskHandler())
		s.setupRoute(gr, "PUT", "/task/status", 0, s.UpdateTaskStatusHandler())
		s.setupRoute(gr, "DELETE", "/task", 0, s.DeleteTaskHandler())

        ...
}
```

## 2.2 更新文档

运行backend后会在解析全部的路由, 在生成程序执行目录生成一个api定义文件 `api_spec.yaml`.

配合 tools/showdoc_updater 下面的工具会解析 `api_spec.yaml` 文件, 转换为showdoc的文档并通过showdoc的api接口推送上去.


```
cd tools/showdoc_updater
go build
// ./showdoc_updater -input=../../cmd/backend/api_spec.yaml -host={showdoc_api_address} -app_id={showdoc_project_appkey}  -app_token={showdoc_project_appsecret}
./showdoc_updater -input=../../cmd/backend/api_spec.yaml -host=http://10.152.208.11 -app_id=225764c9d6f32e98f074338bff166d78894352810  -app_token=628471dfce0da6e992bcbbcd2ab17d302005677984
```

推送后可以在showdoc上面看文档, demo地址 http://10.152.208.11/web/#/5

## 2.3 枚举类型

写文档比较烦的就是处理枚举值了, 这里提供一种方法

common/enum.go
```
func init() {
	initTaskStatus()
	initTaskType()
}

const (
	TaskStatusPending    = 0
	TaskStatusDone       = 1
	TaskStatusProcessing = 2
	TaskStatusCancel     = 3
)

// 注册枚举值
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
```

在请求或返回的结构体添加`enum`注解关联枚举值字段

```
type TaskDetail struct {
	ID       int64  `json:"id"`
	Status   uint8  `json:"status" comment:"任务状态" enum:"TaskStatus"` // 
	Type     string `json:"type" comment:"任务类型" enum:"TaskType"`
	Priority int8   `json:"priority" comment:"任务优先级"`
	Input    string `json:"input" comment:"任务输入"`
	Result   string `json:"result" comment:"任务结果"`
	Note     string `json:"note" comment:"备注"`
}
```
 
生成的文档会自动帮你列出枚举值

```
backend.TaskDetail

|字段名|类型|备注|
|---|---|---|
|id|int64||
|status|uint8|任务状态 枚举类型[0:等待 2:处理中 1:完成 3:取消]|
|type|string|任务类型 枚举类型[image:图片 video:视频]|
|priority|int8|任务优先级|
|input|string|任务输入|
|result|string|任务结果|
|note|string|备注|
```
