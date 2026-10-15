# 公式语法、函数与实现边界

## 1. 语法

| 内容 | 推荐写法 |
|------|----------|
| 字段引用 | `[单价]`、`[数量]`，名称取 Schema 的 `title` |
| 数值/字符串 | `12.5`、`"有货"` |
| 算术 | `+`、`-`、`*`、`/`；用括号明确优先级 |
| 比较 | `>`、`<`、`>=`、`<=`、`=`、`<>` |
| 函数嵌套 | `ROUND(SUM([单价],[运费]),2)` |

不加前导 `=`；函数名使用大写，参数用英文逗号。不能直接套用所有 Excel/飞书函数、数组字面量或 JavaScript lambda 语法。引用子表格/对象需要匹配真实 Schema 与索引结构，不按标题拼接猜测内部 ES 路径。

## 2. 常用函数

这些是当前源码的常用写法；部署版本、ES 编译及业务结果仍需通过目标表查询核验。

| 函数 | 写法 | 注意事项 |
|------|------|----------|
| SUM | `SUM([单价],[运费])` | 当前行参数求和，不是全表聚合 |
| AVERAGE | `AVERAGE([单价],[成本])` | 数值参数 |
| MAX / MIN | `MAX([数量],0)` | 当前行参数比较 |
| ABS | `ABS([差额])` | 一个参数 |
| ROUND | `ROUND([单价]*[数量],2)` | 两个参数；显示位数另设 `numberFormat` |
| POWER / SQRT | `POWER([边长],2)`、`SQRT([面积])` | 注意值域 |
| IF | `IF([数量]>0,"有货","缺货")` | 三个参数，分支类型保持一致；不保证短路 |
| AND / OR / NOT | `AND([数量]>0,[单价]>0)` | 条件组合 |
| SWITCH | `SWITCH([状态],"完成",1,"取消",0,2)` | 匹配项及返回值成对，末尾可放默认值 |
| CONCATENATE | `CONCATENATE("客户：",[客户])` | 文本拼接 |
| LEFT / RIGHT | `LEFT([客户],2)` | 两个参数 |
| MID | `MID([客户],1,2)` | 三个参数；起始位置从 1 开始 |
| LEN / UPPER / LOWER / TRIM | `LEN([客户])` | 各一个参数 |
| SUBSTITUTE | `SUBSTITUTE([客户],"旧","新")` | 另有第四参数指定次数的实现，需单独核验 |
| VALUE | `VALUE([金额文本])` | 无法解析时当前实现返回 `0.0` |
| TEXT | `TEXT([单价],"0.00")` | 仅简化数字格式；返回文本 |
| YEAR / MONTH / DAY | `YEAR([下单日期])` | 日期字段，不是普通日期文本列 |
| NOW / TODAY | `NOW()`、`TODAY()` | 时区和查询时刻影响结果 |
| DATEDIFF | `DATEDIFF([发货日期],[下单日期],"days")` | 结束日期在前、开始日期在后 |
| DATEADD | `DATEADD([下单日期],7,"days")` | 当前 ES 辅助函数返回 `yyyy-MM-dd` 字符串 |
| ARRAYLENGTH | `ARRAYLENGTH(ARRAYSPLIT([客户],","))` | 数组处理；需核验真实输入类型 |

## 3. 日期与时长

当前表格后端实际调用 `PainlessScriptGenerator`，其日期实现应优先于旧的 `FormulaConverter` 示例、前端或 Java 内存执行引擎说明。

- `DATEDIFF(结束,开始,"days")` 计算 `ChronoUnit.DAYS.between(开始日期,结束日期)`；先取 `LocalDate`，是日历天差。
- `DATEDIFF` 单位为 `"days"`、`"months"`、`"years"`、`"hours"`、`"minutes"`、`"seconds"`。小时/分钟/秒目前由日历天差乘系数近似；未知单位返回 `0`。不要写成 `"DAY"`、`"YEAR"`。
- 精确经过时长优先使用日期相减：`([发货日期]-[下单日期])/86400000` 为天数，除以 `3600000` 为小时，除以 `60000` 为分钟。源字段必须具有日期语义，并验证时区及空值策略。
- `DATEADD` 当前识别 `"years"`、`"months"`、`"days"`、`"hours"`、`"minutes"`，输出仅保留日期部分。不把它描述成保留完整时分秒的日期型结果。
- 公式函数 `TEXT` 目前没有日期格式化分支；`TEXT([下单日期],"yyyy-MM-dd")` 会落到字符串回退，不能保证输出要求的日期格式。
- 日期显示使用 Schema 的 `dateFormat: "YYYY-MM-DD"`；这是前端显示格式，与 Java 的 `yyyy-MM-dd` 模式不是同一契约。`DATE`/`DATEADD` 的字符串结果也不能仅靠设置 `functionFieldType: "date"` 就保证通过 ES `emit` 类型检查。

## 4. 空值、类型与依赖

- ES 字段访问辅助函数通常将缺失数值当 `0.0`，文本当 `""`，日期当 `null`。例如当前 `DATEDIFF` 的日期参数为空时返回 `0`，不等于“没有时间差”。
- `functionFieldType: "number"` 会将 Runtime Field 类型设为 Double，`date` 设为 Date；其他类型主要依赖表达式推断。它不是表达式结果的通用强制转换器。
- `IF` 当前生成 `__if(condition, thenValue, elseValue)`，两边表达式可能在调用前求值。因此不要把 `IF([数量]=0,0,[金额]/[数量])` 当成已验证的除零保护。
- 当前 `FormulaParser` 对无法识别的表达式可能回退为字符串；生成器对未支持的普通函数可能生成 `emit(null)`。保存成功不代表语法、函数或结果完全正确。
- 前端有公式依赖拓扑排序；后端字段类型映射对 `function` 列没有同等完整的依赖类型传播。不要承诺公式间引用在 ES 一定等价于前端，也不要断言绝对禁止互相引用；优先直接引用原始列，依赖链需分别核验。

## 5. 当前源码的注册目录

来源：`function-es-core/src/main/java/com/raycloud/formula/engine/functions/FunctionHandlerFactory.java` 的注册表，核查日期：2026-10-15。**114 个注册名称，包含别名**，不是 114 条已经通过 ES 端到端验收的独立函数；也不沿用旧文档固定的“107 个函数”。

### 数学（39）

`ABS`、`ACOS`、`ASIN`、`ATAN`、`AVERAGE`、`CEIL`、`CEILING`、`COS`、`COUNT`、`DEGREES`、`EXP`、`FACT`、`FACTORIAL`、`FLOOR`、`INT`、`LN`、`LOG`、`LOG10`、`MAX`、`MEDIAN`、`MIN`、`MOD`、`PI`、`POW`、`POWER`、`PRODUCT`、`QUOTIENT`、`RADIANS`、`RAND`、`RANDBETWEEN`、`RANDOM`、`ROUND`、`SIGN`、`SIN`、`SQRT`、`SUM`、`TAN`、`TRUNC`、`TRUNCATE`。

### 文本（23）

`CLEAN`、`CONCATENATE`、`EXACT`、`LEFT`、`LEN`、`LOWER`、`MID`、`PROPER`、`REGEXEXTRACT`、`REGEXEXTRACTALL`、`REGEXMATCH`、`REGEXREPLACE`、`REPEAT`、`REPLACE`、`REPT`、`RIGHT`、`SEARCH`、`SUBSTITUTE`、`T`、`TEXT`、`TRIM`、`UPPER`、`VALUE`。

### 日期（21）

`DATE`、`DATEADD`、`DATEDIFF`、`DATEVALUE`、`DAY`、`EOMONTH`、`HOUR`、`ISOWEEKNUM`、`MINUTE`、`MONTH`、`NETWORKDAYS`、`NOW`、`QUARTER`、`SECOND`、`TIME`、`TIMEVALUE`、`TODAY`、`WEEKDAY`、`WORKDAY`、`YEAR`、`YEARFRAC`。

### 逻辑（11）

`AND`、`BLANK`、`ERROR`、`FALSE`、`IF`、`IS_ERROR`、`NOT`、`OR`、`SWITCH`、`TRUE`、`XOR`。

### 数组（20）

`ARRAYAVERAGE`、`ARRAYCONCAT`、`ARRAYCONTAINS`、`ARRAYFILTER`、`ARRAYFORMULA`、`ARRAYINDEX`、`ARRAYJOIN`、`ARRAYLENGTH`、`ARRAYMAP`、`ARRAYMAX`、`ARRAYMIN`、`ARRAYRANGE`、`ARRAYREDUCE`、`ARRAYREPEAT`、`ARRAYREVERSE`、`ARRAYSLICE`、`ARRAYSORT`、`ARRAYSPLIT`、`ARRAYSUM`、`ARRAYUNIQUE`。

### 注册与运行支持的区别

- `CEIL/CEILING`、`POW/POWER`、`FACT/FACTORIAL`、`RAND/RANDOM`、`TRUNC/TRUNCATE`、`REPT/REPEAT` 为注册别名。
- 生成器还有 `CONCAT`、`LENGTH` 分支，但不在以上注册表中；不能只用注册数量判断全部可调用名称。
- `ARRAYFORMULA` 虽有注册，当前高级生成器的数组定义分支没有相应实现，会报不支持的数组函数。
- `ARRAYMAP`/`ARRAYREDUCE`、对象访问、正则组合需要检查生成的脚本与实际 ES 行为；不推导其支持 JavaScript lambda 或任意正则语法。
- `WORKDAY`/`NETWORKDAYS` 当前为简化工作日估算，不处理完整节假日日历。
- 当前未发现可供 CLI 调用的表格公式函数目录/预校验 HTTP 接口。需要新增函数时，应同时核验注册、解析、类型推断、Painless 生成和 ES 执行。
