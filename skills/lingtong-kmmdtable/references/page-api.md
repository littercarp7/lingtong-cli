# 页面接口与回传

页面接口通过 `lingtong-cli api <method> <path>` 调用，命令、参数和返回结构随本 Skill 一起维护。查询参数传 `--params`，JSON 请求体传 `--data`，沿用已授权的 CLI 身份。带 `--envelope` 时，先检查外层 `ok`，再检查 `data.success`。

## 接口与关键字段

| 目的 | 方法与路径 | 参数 | 关键返回字段 |
| --- | --- | --- | --- |
| 查询当前公司绑定应用 | `GET /application/km/multi-table/id` | 无 | `result` 为 `appId` |
| 查询应用资源 | `GET /app/resource/list` | query：`appId` | `result` 为资源列表；页面的 `resourceType` 为 `page`，`id` 即 `pageId` |
| 创建页面 | `POST /page/create` | JSON：`appId`、`name`、`parentFolderId`、`icon` | **`result.id` 即 `pageId`** |
| 查询页面 | `GET /page/get` | query：**`id`** | `result.id`、`appId`、`draftVersionNo` 等 |
| 读取草稿 | `GET /page/draft/source` | query：**`pageId`** | `result.pageId`、`versionNo`、`htmlContent` 等 |
| 保存草稿 | `POST /page/draft/save` | JSON：**`pageId`、`baseVersionNo`、`htmlContent`** | `result.pageId`、`versionId`、`versionNo`、`version`、`htmlContent` |

页面 ID、应用 ID、版本号和分页参数均使用整数。创建、查询及发布返回的页面元数据可包含 `id`、`appId`、`name`、`draftVersionNo`、`draftVersionId`、`published`、`publishedVersionNo`、`publishedVersionId`；`published` 为布尔值，版本 ID 可为空。版本冲突校验使用版本号，不使用版本 ID。

## 查询与创建

以下 ID 仅为参数示例，执行时替换为用户指定或查询确认的目标。

```bash
# 上下文中没有 appId 时查询绑定应用
lingtong-cli api GET /application/km/multi-table/id
lingtong-cli api GET /app/resource/list --params '{"appId":942}'

# 新建页面；已有 pageId 时跳过
lingtong-cli api POST /page/create --data '{"appId":942,"name":"订单概览","parentFolderId":null,"icon":"i-page"}'

# 查询已有页面及草稿
lingtong-cli api GET /page/get --params '{"id":123}'
lingtong-cli api GET /page/draft/source --params '{"pageId":123}'
```

创建参数：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `appId` | integer | 必填，目标应用 ID |
| `name` | string | 最长 100 字符；未指定时由后端默认命名 |
| `parentFolderId` | integer / null | 根目录传 `null`；指定文件夹时传其 ID |
| `icon` | string | 资源图标 ID，默认传 `i-page` |

创建成功的关键返回结构：

```json
{
  "success": true,
  "result": {
    "id": 123,
    "appId": 942,
    "name": "订单概览",
    "draftVersionNo": 0
  }
}
```

后续回传使用 `result.id` 作为 `pageId`。`draftVersionNo: 0` 表示尚无已保存的草稿。

## 保存草稿

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `pageId` | integer | 必填，创建响应的 `result.id` 或用户提供的已有页面 ID |
| `baseVersionNo` | integer | 必填，首次保存传 `0`；更新时传**编辑开始时读取的 `result.versionNo`**，不是 `versionId` |
| `htmlContent` | string | 必填，完整 UTF-8 单文件 HTML，JS、CSS 内联 |

下面的示例从本地文件读取 HTML，用 JSON 序列化和参数数组调用 CLI。三个参数分别是文件路径、pageId、编辑基准版本；执行会实际保存草稿。

```bash
python3 - ./page.html 123 0 <<'PY'
import json
import subprocess
import sys
from pathlib import Path

html_path = Path(sys.argv[1])
page_id = int(sys.argv[2])
base_version_no = int(sys.argv[3])
if page_id <= 0 or base_version_no < 0:
    raise SystemExit('pageId 必须为正整数，baseVersionNo 必须非负')
html = html_path.read_text(encoding='utf-8')
if html_path.suffix.lower() != '.html' or not html.strip():
    raise SystemExit('需要非空 UTF-8 HTML 文件')
body = {'pageId': page_id, 'baseVersionNo': base_version_no, 'htmlContent': html}
command = [
    'lingtong-cli', 'api', 'POST', '/page/draft/save',
    '--data', json.dumps(body, ensure_ascii=False), '--format', 'json', '--envelope'
]
# 使用命名环境/身份时，在 command 中追加与读取阶段相同的 --profile、--auth。
completed = subprocess.run(command, check=True, capture_output=True, text=True)
envelope = json.loads(completed.stdout)
if envelope.get('ok') is not True:
    raise SystemExit('CLI 未确认成功，请查询草稿确认结果，不要直接重发')
response = envelope.get('data') or {}
if response.get('success') is not True:
    raise SystemExit('草稿保存失败：' + str(response.get('msg') or response.get('message') or '业务响应未成功'))
result = response.get('result') or {}
if str(result.get('pageId')) != str(page_id) or not isinstance(result.get('versionNo'), int):
    raise SystemExit('保存回执不完整或页面不匹配，请查询草稿确认结果，不要直接重发')
print(json.dumps({
    'status': 'draft_saved', 'pageId': result['pageId'],
    'versionId': result.get('versionId'), 'versionNo': result['versionNo'],
    'version': result.get('version')
}, ensure_ascii=False))
PY
```

示例使用 `--data` JSON 字符串，不把 HTML 拼接进 shell 命令。若使用命名身份，读取和保存时使用相同的 `--profile`、`--auth`。

## 保存结果

保存草稿及读取草稿的关键返回结构如下。ID 和版本号为示例，`htmlContent` 在真实响应中为完整代码。

```json
{
  "success": true,
  "result": {
    "pageId": 123,
    "versionId": 456,
    "versionNo": 1,
    "htmlContent": "完整 HTML 代码"
  }
}
```

每次成功保存生成新版本。检查业务成功及返回的 `pageId`、`versionNo`，再通过 `/page/draft/source` 回读确认版本和完整内容。

创建成功后保留 `pageId`，后续保存失败不重复创建。版本冲突先读取并合并最新草稿；超时或结果不明确时先查询确认。最终报告页面 ID、草稿版本和保存结果；保存草稿不等于发布。

## 页面管理与发布

| 操作 | 方法与路径 | 参数 | 返回 |
| --- | --- | --- | --- |
| 重命名 | `POST /page/update` | body：`id`、`name` | 检查业务成功 |
| 删除 | `POST /page/delete` | body：`id` | 检查业务成功 |
| 发布指定草稿 | `POST /page/publish` | body：`pageId`、`versionNo` | `result` 为页面元数据 |
| 发布历史 | `GET /page/publish/history` | query：`pageId`、`pageNum`、`pageSize` | `result.data`、`result.total` |
| 草稿 HTML | `GET /page/draft/html` | query：`pageId` | 完整 HTML，由平台加载 |

发布仅在用户明确要求时执行；先查询页面，核对所确认的草稿版本，再以该 `versionNo` 发布，不能悄悄改用新的草稿。发布结果不改变默认的草稿展示流程。删除页面不等同于清空草稿。

```bash
# 仅在用户已明确要求发布对应版本时执行
lingtong-cli api POST /page/publish --data '{"pageId":123,"versionNo":1}' --envelope
lingtong-cli api GET /page/publish/history --params '{"pageId":123,"pageNum":1,"pageSize":10}' --envelope
```
