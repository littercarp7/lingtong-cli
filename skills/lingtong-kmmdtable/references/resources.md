# 资源与文件夹

`GET /application/km/multi-table/id` 无参数，`result` 为绑定应用的整数 `appId`。`GET /app/resource/list` 传 query `{appId}`，`result` 为资源数组。每项包含整数 `id`、`appId`、`sortNum`，字符串 `name`、`icon`、`resourceType`（`table`、`dashboard`、`folder`、`page`），以及整数或 `null` 的 `parentFolderId`。根目录传 `null`，文件夹允许嵌套。

## 创建与命名

| 资源 | 创建路径 | 参数与结果 |
| --- | --- | --- |
| 文件夹 | `POST /app/resource/folder/create` | `appId`、`name`、`parentFolderId`、`icon`；`result.id` 为文件夹 ID |
| 页面 | `POST /page/create` | 同上，`result.id` 为 `pageId`；内容开发见 [自定义页面](custom-page.md) |
| 表格 | `POST /basicdata/save` | 沿用本模块建表约定；命令和字段流程见 [表格 Skill](../../lingtong-table/SKILL.md) |
| 仪表盘 | `POST /dashboard/save` | 新建传 `appId`，可传 `name`；更新需传 `id`，见 [仪表盘](dashboards.md) |

文件夹和页面默认放根目录，默认图标分别为 `i-folder`、`i-page`。未指定名称时使用“文件夹”“文件夹 2”或“页面”“页面 2”等，在应用资源范围避重；可见列表不保证包含全部重名资源，以服务端校验为准。

```bash
lingtong-cli api POST /app/resource/folder/create --data '{"appId":942,"name":"业务分析","parentFolderId":null,"icon":"i-folder"}' --envelope
```

## 移动与排序

`POST /app/resource/sort/update` 的 body 为 `{appId, items}`。每项包含 `resourceType`、`resourceId`、`parentFolderId`、`sortNum`，同级顺序从 `0` 开始。

1. 读取当前资源树，确定移动资源、原父目录和目标父目录。
2. 计算移动后的两个同级列表。`items` 仅提交父目录或顺序实际变化的项，包括原目录和目标目录中受影响的兄弟资源；不提交全量列表。
3. 移动文件夹时保留其内部子资源的父目录与顺序。禁止移入自身或后代形成循环。
4. 被移动资源需要对应编辑权限；可见文件夹可以作为目标。提交成功后重新查询确认。

下面仅展示单项结构，实际 `items` 必须包含本次移动涉及的全部变化项：

```json
{"appId":942,"items":[{"resourceType":"page","resourceId":123,"parentFolderId":45,"sortNum":0}]}
```

## 改名、图标与删除

| 类型 | 改名 | 删除 |
| --- | --- | --- |
| 文件夹 | `POST /app/resource/folder/update`，body `{id,name}` | `POST /app/resource/folder/delete`，body `{id}` |
| 页面 | `POST /page/update`，body `{id,name}` | `POST /page/delete`，body `{id}` |
| 表格 | `POST /basicdata/update`，保留原配置并修改 `name` | `GET /basicdata/delete`，query `id` |
| 仪表盘 | `POST /dashboard/save`，body `{id,appId,name}` | `GET /dashboard/delete`，query `id` |

删除操作按业务含义识别，即使接口使用 GET 也属于删除。文件夹非空时先按用户要求移出或删除子资源，不默认递归删除。

四种资源统一通过 `POST /app/resource/icon/update` 更换图标，body 为 `{appId,resourceType,resourceId,icon}`。`icon` 使用资源图标 ID，默认值为：文件夹 `i-folder`、页面 `i-page`、表格 `i-table`、仪表盘 `i-dashboard`。

```bash
lingtong-cli api POST /app/resource/icon/update --data '{"appId":942,"resourceType":"folder","resourceId":45,"icon":"i-folder"}' --envelope
```
