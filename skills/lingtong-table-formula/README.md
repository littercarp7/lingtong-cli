# lingtong-table-formula 技能

用于通过 `lingtong-cli` 管理绫通表格的公式字段，覆盖金额、条件、文本、日期差、数组计算及 ES/Painless 排查。

## 从哪里开始

- [SKILL.md](SKILL.md)：字段契约、操作顺序、常用公式和错误排查。
- [配置与查询](references/configuration.md)：新建源字段、追加公式、保留旧 Schema、查询核验。
- [公式与函数](references/functions.md)：语法、常用签名、114 个注册名（含别名）及实现限制。

公式列使用 `type: "function"`、`function`、`functionFieldType`；公式引用字段标题，例如 `ROUND([单价]*[数量],2)`。

## 安装与更新

该目录会由 CLI 根目录的 `//go:embed skills` 自动打包。将包含此目录的新版 CLI 安装后，运行：

```bash
lingtong-cli skills list
lingtong-cli skills install --editor opencode --scope project --dry-run
lingtong-cli skills install --editor opencode --scope project
```

支持的其他编辑器使用 `skills install --help` 查询。旧二进制不会从源码目录动态读取新技能，需要先更新或重新构建 CLI。项目级与全局同名技能可能相互遮蔽，应核对 Agent 实际加载的路径并更新相应 scope。

## 相关技能

- `lingtong-table`：表格、Schema 与记录管理。
- `lingtong-pivot-table`：透视表与聚合。
- `lingtong-script`：工作流/连接器脚本。
