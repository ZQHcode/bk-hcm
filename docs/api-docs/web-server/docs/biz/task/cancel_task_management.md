### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：业务下任务管理操作。
- 该接口功能描述：取消任务。

### URL

PATCH /api/v1/cloud/bizs/{bk_biz_id}/task_managements/cancel

### 输入参数

| 参数名称       | 参数类型   | 必选 | 描述              |
|------------|--------|----|-----------------|
| ids        | string array    | 是  | 任务id列表，最大长度为100 |

### 调用示例

```json
{
  "ids": ["0000001","0000002"]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |