# 公式字段配置与查询

## 1. 先确认执行环境

- 先完成 `lingtong-shared` 的 Host/认证配置。表格、Schema、数据查询和更新使用同一身份。
- 面向 ES 查询、过滤、排序的公式场景，先确认表格 `openHighMode: 1`、服务端搜索查询开关、租户 `openEsQuery` 和 ES 同步状态。
- 高性能模式下已有普通字段的类型及主键受后端约束；不要为修复公式而直接修改源字段类型。
- 公式是 Schema 配置；记录数据不需要手工写入计算结果。

## 2. 新建表格：先建立源字段

确认用户要在应用 `165` 中创建表格后执行（ID 为示例，替换为目标应用）：

```bash
lingtong-cli table create --app-id 165 --name "订单公式示例" --source 1 --type 1 --open-high-mode 1 \
  --columns-schema '[
    {"id":1,"title":"订单号","key":"orderNo","type":"text","primaryKey":true},
    {"id":2,"title":"单价","key":"price","type":"number","numberFormat":"1.00"},
    {"id":3,"title":"数量","key":"quantity","type":"number"}
  ]'
```

使用响应 `result.id` 作为表格 ID、`schemaResult.id` 作为 Schema ID，再通过 `table schema query` 读取当前配置和版本。

这里分两步建立公式：先让源字段 mapping 存在，再更新公式。当前后端保存 Schema 时先构建 Runtime Field，之后才更新普通字段 mapping；新表在首次 Schema 中同时带公式，可能因源字段 mapping 尚未存在而失败。已有表若同批新增源字段及依赖它的公式，也应先建立源字段再配置公式。

## 3. 在原 Schema 上合并公式列

```bash
lingtong-cli table schema query --basic-data-id 1568 --omit-null=false
```

默认业务类型是 `2`（基础资料/表格），`1` 是场景。读取响应中的 `result`，保留 Schema 配置和所有 `columnsSchema`，选用未占用的列 ID/key，在数组末尾追加：

```json
{
  "id": 4,
  "title": "金额",
  "key": "amount",
  "type": "function",
  "function": "ROUND([单价]*[数量],2)",
  "functionFieldType": "number",
  "numberFormat": "1.00",
  "show": true,
  "width": 200,
  "align": "right",
  "primaryKey": false,
  "dataDisabled": true
}
```

这是**新增列的定义片段**，不能单独作为 `columnsSchema` 提交。CLI 的 Schema 更新不补齐新列默认属性，也不合并现有列。

假设刚读取的 Schema `id=2367`、`version=5`，且原表确实只有上述三列，下列为保留三列并追加公式的请求结构示例。实际使用必须从原 Schema 合并，保留其余表级和列级属性：

```json
{
  "id": 2367,
  "businessId": 1568,
  "businessType": 2,
  "version": 5,
  "columnsSchema": [
    {"id":1,"title":"订单号","key":"orderNo","type":"text","primaryKey":true,"show":true},
    {"id":2,"title":"单价","key":"price","type":"number","numberFormat":"1.00","show":true},
    {"id":3,"title":"数量","key":"quantity","type":"number","show":true},
    {"id":4,"title":"金额","key":"amount","type":"function","function":"ROUND([单价]*[数量],2)","functionFieldType":"number","numberFormat":"1.00","show":true,"primaryKey":false,"dataDisabled":true}
  ]
}
```

将完整对象作为 JSON 字符串传入现有命令：

```bash
lingtong-cli table schema update --schema-id 2367 --schema '<从当前Schema合并后的完整JSON>'
```

`--schema` 接收 JSON 字符串，不支持 `@文件` 或 `--schema-file`。脚本可读取本地 JSON 后作为一个参数传入。若改用 `--columns-schema`，也必须传完整数组；`--basic-data-id` 只自动获取当前版本。

后端校验 `version` 并加一；遇到“表格列已被其他人更新”时重新读取和合并，不能仅替换版本号重放旧数组。

## 4. 保存与查询核验

1. 检查 CLI 退出状态及平台响应的 `success`/错误信息。
2. 再查 Schema，核对公式列、原列、主键和新版本。
3. 选择已存在且数据已同步的明确样本：单价 `12.5`、数量 `4`，预期金额 `50`。

   ```bash
   lingtong-cli table data query --basic-data-id 1568 --ids 1001 --omit-null=false
   ```

4. 按 Schema 中金额列的 ID 检查 `result.data[].data["4"]`，不要按标题或 `key` 取记录值。
5. 在目标环境验证公式列过滤、排序、空值、日期边界、更新源字段后的结果。确认查询走 ES 后再以查询结果判定后端公式生效。

ES Runtime Field 在查询时执行，配置公式不等于在关系数据库逐行回填数据。前端也有自己的公式计算逻辑；页面结果与后端查询结果需分别核验。

## 5. 底层 API 对照

| CLI 操作 | 内部 API | 关键参数 |
|----------|----------|----------|
| `table create` | `POST /basicdata/save`，可接着调用 `POST /basicdata/schema/save` | `openHighMode: 1`、小写字段类型 |
| `table schema query` | `GET /basicdata/schema/get` | `businessId`、`businessType: 2` |
| `table schema update` | `POST /basicdata/schema/update` | Schema `id`、读取时 `version`、完整 `columnsSchema` |
| `table data query` | `POST /basicdata/record/listNew` | `basicDataId`、分页等；由后端决定查询路由 |

CLI 统一处理代理及认证，通常不需要手写 HTTP。当前没有独立的表格公式目录、预校验或试算命令；现有 `/function/desc/list` 面向连接器/脚本，不应当作 `function-es-core` 的函数目录。
