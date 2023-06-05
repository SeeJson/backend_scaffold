# 后台API文档

## 用户

### 测试登录

测试是否正常登录

#### 请求URL:
- /backend/api/v1/hola

#### 请求方式:
- GET

#### 请求格式:
- query

#### 所需权限:
- 无

#### 请求参数:



#### 请求示例:



#### 返回参数:



|字段名|类型|备注|
|---|---|---|
|code|int||
|msg|string||
|error|string||


#### 返回示例:

```
{
  "code": 0,
  "msg": "",
  "error": ""
}
```


------
### 登录

后台登录

#### 请求URL:
- /backend/api/v1/login

#### 请求方式:
- POST

#### 请求格式:
- json

#### 所需权限:
- 无

#### 请求参数:



|字段名|类型|备注|
|---|---|---|
|name|string||
|password|string||


#### 请求示例:

```
{
  "name": "",
  "password": ""
}
```

#### 返回参数:



|字段名|类型|备注|
|---|---|---|
|code|int||
|msg|string||
|error|string||
|data|backend.LoginData||


backend.LoginData

|字段名|类型|备注|
|---|---|---|
|uid|int64||
|name|string||
|role|uint32||
|perm|uint64||


#### 返回示例:

```
{
  "code": 0,
  "msg": "ok",
  "error": "",
  "data": {
    "uid": 0,
    "name": "",
    "role": 0,
    "perm": 0
  }
}
```


------
### 登出

登出

#### 请求URL:
- /backend/api/v1/logout

#### 请求方式:
- POST

#### 请求格式:
- json

#### 所需权限:
- 无

#### 请求参数:



#### 请求示例:



#### 返回参数:



|字段名|类型|备注|
|---|---|---|
|code|int||
|msg|string||
|error|string||


#### 返回示例:

```
{
  "code": 0,
  "msg": "",
  "error": ""
}
```


------


## 任务

### 新增任务

新增任务

#### 请求URL:
- /backend/api/v1/task

#### 请求方式:
- POST

#### 请求格式:
- json

#### 所需权限:
- 无

#### 请求参数:



|字段名|类型|备注|
|---|---|---|
|type|string|任务类型 枚举类型[image:图片 video:视频]|
|priority|int8|任务优先级|
|input|string|任务输入|
|note|string|备注|


#### 请求示例:

```
{
  "type": "",
  "priority": 0,
  "input": "",
  "note": ""
}
```

#### 返回参数:



|字段名|类型|备注|
|---|---|---|
|code|int||
|msg|string||
|error|string||


#### 返回示例:

```
{
  "code": 0,
  "msg": "ok",
  "error": ""
}
```


------
### 任务详情

任务详情

#### 请求URL:
- /backend/api/v1/task

#### 请求方式:
- GET

#### 请求格式:
- query

#### 所需权限:
- 无

#### 请求参数:



|字段名|类型|备注|
|---|---|---|
|id|int64||


#### 请求示例:



#### 返回参数:



|字段名|类型|备注|
|---|---|---|
|code|int||
|msg|string||
|error|string||
|data|backend.TaskDetail||


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


#### 返回示例:

```
{
  "code": 0,
  "msg": "",
  "error": "",
  "data": {
    "id": 0,
    "status": 0,
    "type": "",
    "priority": 0,
    "input": "",
    "result": "",
    "note": ""
  }
}
```


------
### 更新任务信息

更新任务信息

#### 请求URL:
- /backend/api/v1/task

#### 请求方式:
- PUT

#### 请求格式:
- json

#### 所需权限:
- 无

#### 请求参数:



|字段名|类型|备注|
|---|---|---|
|id|int64||
|note|string|备注|
|priority|int8|任务优先级|


#### 请求示例:

```
{
  "id": 0,
  "note": "",
  "priority": 0
}
```

#### 返回参数:



|字段名|类型|备注|
|---|---|---|
|code|int||
|msg|string||
|error|string||


#### 返回示例:

```
{
  "code": 0,
  "msg": "ok",
  "error": ""
}
```


------
### 更新任务状态

更新任务状态

#### 请求URL:
- /backend/api/v1/task/status

#### 请求方式:
- PUT

#### 请求格式:
- json

#### 所需权限:
- 无

#### 请求参数:



|字段名|类型|备注|
|---|---|---|
|id|int64||
|status|uint8|任务状态 枚举类型[0:等待 2:处理中 1:完成 3:取消]|


#### 请求示例:

```
{
  "id": 0,
  "status": 0
}
```

#### 返回参数:



|字段名|类型|备注|
|---|---|---|
|code|int||
|msg|string||
|error|string||


#### 返回示例:

```
{
  "code": 0,
  "msg": "ok",
  "error": ""
}
```


------
### 删除任务

删除任务

#### 请求URL:
- /backend/api/v1/task

#### 请求方式:
- DELETE

#### 请求格式:
- json

#### 所需权限:
- 无

#### 请求参数:



|字段名|类型|备注|
|---|---|---|
|ids|[]int64||


#### 请求示例:

```
{
  "ids": null
}
```

#### 返回参数:



|字段名|类型|备注|
|---|---|---|
|code|int||
|msg|string||
|error|string||


#### 返回示例:

```
{
  "code": 0,
  "msg": "ok",
  "error": ""
}
```


------
### 任务列表

任务列表

#### 请求URL:
- /backend/api/v1/tasks

#### 请求方式:
- GET

#### 请求格式:
- query

#### 所需权限:
- 无

#### 请求参数:



|字段名|类型|备注|
|---|---|---|
|offset|int||
|limit|int||


#### 请求示例:



#### 返回参数:



|字段名|类型|备注|
|---|---|---|
|code|int||
|msg|string||
|error|string||
|total|int64||
|items|[]backend.TaskItem||


backend.TaskItem

|字段名|类型|备注|
|---|---|---|
|id|int64||
|status|uint8|任务状态 枚举类型[0:等待 2:处理中 1:完成 3:取消]|
|type|string|任务类型 枚举类型[image:图片 video:视频]|
|priority|int8|任务优先级|
|note|string|备注|
|user|backend.UidName|用户信息|


backend.UidName

|字段名|类型|备注|
|---|---|---|
|uid|int64|用户名|
|name|string|用户名|


#### 返回示例:

```
{
  "code": 0,
  "msg": "",
  "error": "",
  "total": 0,
  "items": [
    {
      "id": 0,
      "status": 0,
      "type": "",
      "priority": 0,
      "note": "",
      "user": {
        "uid": 0,
        "name": ""
      }
    },
    {
      "id": 0,
      "status": 0,
      "type": "",
      "priority": 0,
      "note": "",
      "user": {
        "uid": 0,
        "name": ""
      }
    }
  ]
}
```


------


## 其它

### 上传图片

上传图片

#### 请求URL:
- /backend/api/v1/upload

#### 请求方式:
- POST

#### 请求格式:
- form-data

#### 所需权限:
- 无

#### 请求参数:



|字段名|类型|备注|
|---|---|---|
|file|表单文件||


#### 请求示例:



#### 返回参数:



|字段名|类型|备注|
|---|---|---|
|code|int||
|msg|string||
|error|string||
|uri|string||
|full_uri|string||


#### 返回示例:

```
{
  "code": 0,
  "msg": "",
  "error": "",
  "uri": "",
  "full_uri": ""
}
```


------


