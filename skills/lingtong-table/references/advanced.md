# 表格扩展操作

表格、视图、筛选、同步及下载接口的参数和 DTO 统一维护在 CLI 共享 OpenAPI 规范。先用 `lingtong-cli schema path /路径` 和对应 `service ... --help` 查参数；常规表格、Schema、记录及视图操作仍使用 `table ...` 专用命令。

## 筛选与展示

`/table/filterSort/get` 读取已有筛选排序配置，`/table/filterSort/save` 保存修改。分组使用视图配置，分组展开/收起不修改业务记录；字段显示、列顺序、列宽及颜色沿用 Schema 配置，保留未修改的属性。

## Excel 导入与导出

1. 文件先上传到平台，取得 `fileUrl`，不能将本地路径当作 URL。宿主上传辅助接口为 `GET /storage/oss/getStsToken`，可带时间戳 `_t`；临时上传凭据仅用于上传。
2. `POST /basicdata/record/import/columnsSchema` 用 query `fileUrl` 解析列。支持 POST query 的 `service` 命令可按 `--help` 调用；使用通用 `api` 时须把 URL 编码的 query 放入路径。
3. `POST /basicdata/record/import` 用 multipart 表单提交 `basicDataId`、`fileUrl`；当前通用 `api` 和 `service --body` 不能替代表单上传能力。
4. 返回任务 UUID 后通过 `/basicdata/record/getImportProgress` 查询，区分任务创建与实际完成。

导入模板通过 `/basicdata/record/getImportTemplate` 获取。导出通过 `/basicdata/record/export` 发起，再由 `/download/center/queryList` 和 `/download/center/queryFileUrl` 查询任务及下载地址；导出任务已创建不等于文件已生成。

## 字典与连接器同步

- 字典：`/connectorDict/list` 查询，`/basicdata/record/dictionary/import` 导入。
- 连接器配置：`/basicdata/connector/get`、`save`；请求参数通过其 `request/get`、`request/save` 维护。
- 自动同步：`/basicdata/connector/sync/get`、`getRequestField`、`save`；先读取并保留现有规则。
- 手动同步：`/basicdata/connector/sync/manualSync`，按用户要求触发数据变更。
- 开启状态：`POST /basicdata/open/update` 的 `id`、`open`、`openTask` 放在 query。

账户、连接器和模型取数分别复用对应业务 Skill。仅调整同步配置时不隐含立即执行同步。
