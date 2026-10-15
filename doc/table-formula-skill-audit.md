# 表格公式技能遗漏核查与补充

核查日期：2026-10-15。范围：`lingtong-cli` 的技能分发，以及相邻 `lingtong-app`、`lingtong-app-view`、`function-es-core` 的当前源码。

## 1. 结论

**遗漏属实，发生在 CLI 技能内容与分发链路。**

- 补充前 CLI 二进制的 `skills list` 返回 11 个有效技能，无 `lingtong-table-formula`。
- `skills/lingtong-table/SKILL.md` 已有 `type: "function"` 的简单字段片段，但没有完整公式语法、函数目录、日期边界、ES 查询验证流程，也没有导航到公式专项技能。
- 工作区 `.opencode/skills/lingtong-table-formula/` 虽有旧专项技能，但它不是 CLI `skills/` 源目录的一部分，因此不会由 CLI 自动打包分发。全局与项目级技能还可能存在同名版本遮蔽。
- 当前 CLI 已能透传公式 Schema 属性；不需要先新增表格公式命令才能配置计算列。

## 2. 证据与真实实现链路

下述路径相对本工作区，CLI 自身路径带 `lingtong-cli/` 前缀。

| 位置 | 核查结果 |
|------|----------|
| `lingtong-cli/skills_embed.go` | `//go:embed skills` 打包仓库内的技能目录 |
| `lingtong-cli/internal/skills/skills.go` | 仅将含 `SKILL.md` 的一级目录视为技能，自动列举，无额外硬编码注册清单 |
| `lingtong-cli/internal/skills/install.go` | 复制完整技能子树，并从实际技能列表生成 Agent 受管指令块 |
| `lingtong-cli/cmd/table/table.go` | 创建、查询及更新 Schema；保留 `function`/`functionFieldType`，更新提交完整列数组 |
| `lingtong-app/lingtong-app-api/.../basicdata/domain/ColumnsSchema.java` | 公式属性为 `function`、`functionFieldType`；显示属性为 `numberFormat`、`dateFormat` |
| `lingtong-app/lingtong-app-api-support/.../basicdata/service/BasicDataFormSchemaServiceImpl.java` | 保存/更新 Schema 时调用 ES mapping；更新校验 `version` 后加一 |
| `lingtong-app/lingtong-app-api-support/.../search/es/EsMappingService.java` | 以列 `title` 建立字段映射，实际调用 `FormulaParser`、`ASTTypeAnalyzer`、`PainlessScriptGenerator`，向索引 mapping 注册 `data<businessId>.<columnId>` Runtime Field |
| `lingtong-app/lingtong-app-api-support/.../search/EsSelectAspect.java` | 搜索查询开关及租户 `openEsQuery` 控制 ES 路由；ES 查询时请求公式列 `fields` |
| `lingtong-app-view/src/components/RCFormulaEditor/table/index.vue` | 编辑器插入标题路径 |
| `lingtong-app-view/src/components/LTTable/minix/formula.js` | 前端另有公式引擎、按列标题提取依赖及拓扑排序，页面计算不等于 ES 结果 |
| `function-es-core/src/main/java/com/raycloud/formula/engine/functions/FunctionHandlerFactory.java` | 当前注册 114 个名称，数学 39、文本 23、日期 21、逻辑 11、数组 20，含别名 |
| `function-es-core/src/main/java/com/raycloud/formula/engine/painless/PainlessScriptGenerator.java` | 表格当前实际使用的高级 Painless 生成器；与旧正则转换器、Java 内存执行路径存在语义差别 |

链路为：

```text
table schema update（type=function、function、functionFieldType）
  → BasicDataFormSchemaServiceImpl
  → EsMappingService：标题到 ES 字段路径映射
  → FormulaParser + ASTTypeAnalyzer + PainlessScriptGenerator
  → ES index mapping.runtime
  → 查询时请求公式 fields 并计算
```

## 3. 旧技能中不能直接沿用的内容

| 旧内容/假设 | 当前核查口径 |
|-------------|--------------|
| `type: "FORMULA"`、`formula` | 小写 `type: "function"`、表达式属性 `function` |
| `highPerformance: true` | 表格配置 `openHighMode: 1`，CLI `--open-high-mode 1` |
| 基础资料 `businessType: 1` | 表格/基础资料为 `2`，`1` 为场景 |
| `[title]` 和 `[key]` 可以任意替换 | 当前表格后端按 `title` 建映射，记录数据才以列 ID 为键 |
| 固定“107 个函数” | 当前注册表为 114 个名称（含别名），不等于完整 ES 支持矩阵 |
| `DATEDIFF(...,"DAY")` | 当前高级生成器使用小写复数 `"days"`，结束日期在前 |
| `TEXT` 可直接格式化日期 | 当前实现只有简化数字格式化和 `toString()` 回退 |
| 只提交新增公式列 | 会替换完整列数组，必须在原 Schema 上合并并保留旧列及版本 |
| `table schema update` 自动填默认值 | 仅创建表格时 CLI 补默认列属性；更新需显式设置 |
| 公式绝对不能互相引用 / 公式链完全可用 | 前端有依赖处理，ES 类型传播不同；需分别核验，不能作任一绝对承诺 |
| Java 内存性能或转换耗时可直接代表表格表现 | 当前未做对应基准，不纳入技能能力承诺 |

## 4. 补充方案与落地文件

采用“表格主入口 + 可独立发现的公式专项技能”：

- `skills/lingtong-table-formula/SKILL.md`：触发描述、契约、配置/查询流程、公式示例和排查。
- `skills/lingtong-table-formula/references/configuration.md`：先建源字段映射、完整 Schema 合并、版本控制、查询验收与 API 对照。
- `skills/lingtong-table-formula/references/functions.md`：常用签名、日期/空值/类型边界、完整注册名与已知实现限制。
- `skills/lingtong-table-formula/README.md`：入口及安装更新方式。
- `skills/lingtong-table/SKILL.md`、`README.md`：补公式检索词与导航，澄清更新不合并列、不填默认值。
- 主 `README.md`：技能清单由 11 调整为 12，并说明新版二进制重新安装流程。

现有自动发现及嵌入机制会包含新增技能和 reference，无需修改 Go 的安装器或 OpenAPI。

## 5. 实现边界与后续演进

本次技能按当前实现描述，并暴露需要额外核验的边界：

1. **独立预校验/试算能力**：当前 CLI、嵌入的 OpenAPI 及相关控制器中未发现表格公式专用目录/校验接口；`/function/desc/list` 属于连接器/脚本说明。后续可在后端提供使用同一字段映射/生成器的目录、校验、试算 API，再映射 CLI 命令。
2. **注册与实现差异**：注册数不能代表端到端可用性。例如 `ARRAYFORMULA` 已注册，但高级生成器没有数组定义实现；生成器还另有 `CONCAT`、`LENGTH` 分支。
3. **日期语义**：`DATEDIFF` 的时分秒使用日历天差近似，`DATEADD` 返回日期字符串；`TEXT` 暂无日期格式化。应与前端及 Java 引擎建立共同用例后再统一契约。
4. **错误提前暴露**：解析器可能回退为字符串，未知普通函数可生成 `emit(null)`，`IF` 不保证分支短路。仅 Schema 保存成功不能证明公式正确。
5. **新源字段映射时序**：后端在普通 mapping 更新前构建 Runtime Field，技能示例先建普通源列及 mapping，再追加公式。
6. **公式依赖类型**：当前表格 ES 字段映射中公式列未获得完整的依赖类型传播；应对公式链、子表格和复杂数组补充真实 ES 用例。

## 6. 验证记录

- 补充前只读 Agent 应用测试：只能从表格技能获得基础金额公式，无法找到日期单位、函数目录、预校验能力与实现边界。
- 补充后只读 Agent 应用测试：可正确输出按标题引用的金额列、`DATEDIFF(结束,开始,"days")`、TEXT 日期限制、完整 Schema 合并步骤，以及 ARRAYFORMULA 和页面/后端差异。复验发现的默认值/单列示例风险已在主技能相邻位置补充说明。
- CLI：`go test ./internal/skills ./cmd/skills ./cmd/table`，三个包全部通过；`go build -o <临时目录>/lingtong-cli-table-formula .` 构建成功。
- 新二进制分发核验：`skills list` 返回 12 个技能并包含 `lingtong-table-formula`；在临时项目执行 OpenCode project scope 安装，四个公式文档与源码逐字节一致，`.opencode/AGENTS.md` 自动包含公式技能。`--dry-run` 不写入，重复安装后受管块无重复条目。
- 文档静态核验：14 处公式相关链接存在，4 个 JSON 示例可解析，函数参考的 114 个注册名与源码注册表集合完全一致。
- 离线生成探针：使用与表格后端相同的类型解析器、字段映射和高级生成器，12 条金额/条件/文本/日期/数组表达式均生成预期辅助调用与字段路径；另验证 `ISBLANK` 的 null 回退、`TEXT` 无日期格式化分支。这仅证明脚本生成，不是 ES 编译或计算结果通过。
- `function-es-core`：`mvn -o -Dtest=DateArithmeticScriptTest test`，5 项通过。
- 初次联合运行 `DateArithmeticScriptTest,ASTBasedScriptGenerationTest,LogicalFunctionTypeTest`：前者 5 项通过；`ASTBasedScriptGenerationTest` 的 5 项因本地 ES 连接拒绝报错；`LogicalFunctionTypeTest` 无可用 ES，运行 0 项。该结果不能表述为 ES 集成通过。

本次目标是技能契约与分发验证。真实租户中的 ES 编译、查询、筛选、排序及前后端一致性应按技能中的样本流程验收。
