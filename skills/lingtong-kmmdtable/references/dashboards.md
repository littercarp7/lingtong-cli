# 仪表盘流程

仪表盘是当前应用下的 `dashboard` 资源，以 `id` 定位。接口和配置 DTO 统一维护在 CLI 共享规范，通过 [接口索引](api-catalog.md) 查询；没有额外复制的仪表盘接口结构。

## 新建与查看

1. 使用当前应用 `appId`，通过 `/dashboard/save` 新建，记录响应中的 `result.id`。
2. 表格数据源、Schema 与视图通过 [表格 Skill](../../lingtong-table/SKILL.md) 查询；透视配置通过 [透视表 Skill](../../lingtong-pivot-table/SKILL.md) 处理。
3. 读取仪表盘详情，再配置组件；通过数据查询接口检查结果。

```bash
lingtong-cli api POST /dashboard/save --data '{"appId":942,"name":"经营概览"}' --envelope
lingtong-cli api GET /dashboard/get --params '{"id":12}' --envelope
lingtong-cli api POST /dashboard/query --data '{"id":12}' --envelope
```

## 配置与布局

`config` 中的 `layoutList` 决定位置与尺寸，`tableConfigList` 保存组件的数据源、图表和样式，`globalFilterList` 保存全局筛选。布局键 `i` 对应组件 ID；位置尺寸为 `x`、`y`、`w`、`h`。

开发组件时先确认数据源表格、视图和字段类型，再组装配置。只调整布局时，读取最新 `config`，按 `i` 合并指定组件的位置尺寸，保留组件定义及全局筛选，通过 `/dashboard/save` 提交 `{id,appId,config}`。

重命名只提交 `{id,appId,name}`。编辑仪表盘需要仪表盘编辑权限；修改底层表格透视配置同时遵循表格编辑权限。保存后重新读取详情或查询数据确认结果。
