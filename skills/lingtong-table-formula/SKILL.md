---
name: lingtong-table-formula
description: "绫通表格公式字段：创建和更新计算列、编写公式、查询结果、排查 Elasticsearch Runtime Field/Painless 转换问题。当用户需要金额、利润率、条件判断、文本处理、日期差、数组计算，或使用 type=function、functionFieldType 时使用。关键词：公式、计算字段、formula、SUM、IF、ROUND、DATEDIFF、DATEADD、Painless、ES。"
---

# 绫通表格公式字段

通过现有 `lingtong-cli table schema` 命令配置公式，后端将表达式转换为 Elasticsearch Runtime Field 的 Painless 脚本，查询时计算。

## 导航

- 新建计算列、保留旧 Schema、核验结果：[配置与查询](references/configuration.md)
- 公式语法、常用函数、注册目录、日期及空值边界：[公式与函数](references/functions.md)
- 表格 CRUD、记录读写：`lingtong-table`
- 工作流/连接器 JavaScript 脚本：`lingtong-script`

## 字段契约

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
  "primaryKey": false,
  "dataDisabled": true
}
```

- 字段类型为小写 `function`；表达式字段名为 `function`，不是 `formula` 或 `type: "FORMULA"`。
- **公式引用字段 `title`**，如 `[单价]`；记录 `data` 的 key 才是字段 ID 字符串。不能将 `[price]`、`[1]`、`[data123.1]` 当作 `[单价]` 的等价写法。改名后应检查所有引用公式。
- `functionFieldType` 可为 `normal`、`text`、`textarea`、`number`、`date`。数字计算推荐显式 `number`；该属性不会把任意文本或日期字符串自动转换为正确的底层结果。
- `numberFormat`、`dateFormat` 是显示配置，不改变公式运算。日期列的 `dateFormat` 用 `YYYY-MM-DD` 等前端格式。

## 标准操作顺序

1. 使用同一 Host 和身份查询表格及 Schema：

   ```bash
   lingtong-cli table list --app-id 165
   lingtong-cli table schema query --basic-data-id 1568 --omit-null=false
   ```

2. 核对 `openHighMode: 1`、源字段名称/类型、现有 `columnsSchema`、Schema `id` 和 `version`。ES 查询还依赖服务端搜索配置、租户 `openEsQuery` 和数据同步；仅打开高性能模式不代表查询已走 ES。
3. 在原 Schema 上追加/修改公式列，保留其余列、字段 ID、key、主键和配置。**`columnsSchema` 是完整替换，不能只提交新增列。**
4. 用读取时的 `version` 提交完整 Schema；真实写入前确认目标与意图：

   ```bash
   lingtong-cli table schema update --schema-id 2367 --schema '<保留原配置并合并公式列后的完整Schema JSON>'
   ```

   版本冲突时重新读取并合并。`--basic-data-id --columns-schema` 只会自动取版本，不会合并字段；不要用最新版本覆盖基于旧快照生成的列数组。
5. 再查 Schema，核对公式、版本及旧列；用有明确预期值的记录查询计算结果：

   ```bash
   lingtong-cli table data query --basic-data-id 1568 --ids 1001 --omit-null=false
   ```

   按公式列 ID 检查 `result.data[].data`。若单价为 `12.5`、数量为 `4`，金额应为 `50`。仅保存成功或页面显示成功，不足以证明 ES 查询、过滤和排序已正确计算。

新建表格先创建普通源字段、建立映射，再追加公式列。完整示例见 [配置与查询](references/configuration.md)。

## 常用写法

| 目标 | 表达式 | 结果说明 |
|------|--------|----------|
| 金额 | `ROUND([单价]*[数量],2)` | 数值；显式 `number` |
| 状态 | `IF([数量]>0,"有货","缺货")` | 两个分支都返回文本 |
| 拼接 | `CONCATENATE("客户：",[客户])` | 文本 |
| 年份 | `YEAR([下单日期])` | 数值 |
| 发货日减下单日 | `DATEDIFF([发货日期],[下单日期],"days")` | 日历天差；参数为结束、开始 |
| 精确时长（小时） | `([发货日期]-[下单日期])/3600000` | 日期相减按毫秒计算 |
| 金额文本 | `TEXT([单价],"0.00")` | 简化数字格式化；不是日期格式化 |

## 使用边界

- 使用大写函数名、英文逗号/括号、引号包围字符串，不加 Excel 的前导 `=`。
- 日期单位用 `"days"`、`"months"`、`"years"` 等小写复数；`"DAY"` 不是当前 ES 实现的等价值。`DATEDIFF` 的小时/分钟/秒目前按日历天数近似，精确时长用日期相减。
- `TEXT([日期],"yyyy-MM-dd")` 当前没有对应的日期格式化实现；保留日期类型并配置 `dateFormat`。`DATEADD` 当前生成日期字符串，不能直接承诺它适配 `functionFieldType: "date"`。
- 数值缺失默认通常为 `0.0`，文本为空字符串，日期为 `null`；按实际函数核验业务上的缺失值语义。
- `IF` 当前生成辅助函数调用，不能依赖分支短路来保护除零或非法函数参数。分支结果类型应一致。
- 不直接假定飞书/Excel 的 `ISBLANK`、`IFERROR`、`LOOKUP`、lambda 语法可用。当前注册目录、实现限制见 [公式与函数](references/functions.md)。
- 当前 CLI/OpenAPI 没有独立的表格公式目录或公式校验接口；不要编造 `table formula validate` 等命令。`/function/desc/list` 是连接器/脚本函数说明，不是该公式引擎的目录。

## 排查顺序

1. **列不显示**：确认 `show: true`、前端可见配置及保存后的完整 Schema。
2. **构建 ES 运行时字段异常**：核对字段标题和源字段类型、映射是否已建立、函数名/参数、生成脚本及 ES mapping/编译错误。
3. **有列但无计算值**：检查 ES 查询开关、租户、同步和索引中的源数据，避免把数据库回退查询当成 Runtime Field 查询。
4. **日期差/类型不对**：先检查单位、参数顺序、时区、返回类型；区分日历差与毫秒时长。
5. **页面正确、CLI/筛选错误**：前端另有公式引擎及依赖排序，应验证后端 ES 结果。公式互相引用、子表格、数组和正则组合需在目标环境单独核验。
