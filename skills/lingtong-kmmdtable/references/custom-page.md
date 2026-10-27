# 自定义页面开发与草稿回传

页面资源类型为 `page`，资源的 `id` 就是 `pageId`。开发前按主 Skill 完成 CLI 授权。

## 数据流程与页面开发

1. 按需求使用 CLI 的表格、连接器、连接器工厂、场景、模型等模块及对应业务 Skill 查找接口，获取真实数据，确认参数、字段和分页范围。
2. 基于真实数据开发页面；页面打开后，按平台登录态直接请求业务后端获取数据。开发取数通过 CLI，页面运行时取数通过浏览器业务接口。
3. 平台提供公共后端服务，Agent 只需交付一个 UTF-8 HTML 文件。技术栈不限，JS、CSS 和所需框架依赖全部内联，不依赖外部 JS/CSS 或本机文件。
4. HTML 只实现业务内容，不重复实现宿主菜单和页面管理功能。适配桌面与移动端，检查主要交互、数据展示及加载、空数据、失败状态。

用户要求先预览时，先提供预览，按其确认顺序回传；已明确要求回传时直接完成授权范围内的操作。

## 页面运行时请求约定

- 使用当前部署环境的平台 Web 业务接口，例如 `/admin/ltapp/...`；不使用 `/gw` 网关，也不把 `/gw` 作为失败重试或备用前缀。
- 登录态由浏览器携带当前 Cookie。页面不从 Cookie、`localStorage` 或 `sessionStorage` 中读取、缓存 Token，不手动拼接 `Authorization: Bearer ...` 或设置 `Cookie` 请求头，也不将 CLI Token 写入 HTML。
- 使用 Fetch 时设置 `credentials: 'include'`；使用 Axios 时设置 `withCredentials: true`。

下面的 `businessUrl` 是按当前部署环境确认的 Web 业务接口地址：

```js
const response = await fetch(businessUrl, {
  method: 'GET',
  headers: { Accept: 'application/json' },
  credentials: 'include'
});
```

以上约定仅用于生成页面中的浏览器请求；Agent 开发取数和保存草稿仍通过 CLI，并沿用 CLI 自身的授权流程。

## 新建页面

1. 从用户上下文或绑定应用查询取得目标 `appId`，开发并验证 HTML。
2. 调用 `POST /page/create`，传 `appId`、`name`、`parentFolderId`、`icon`。默认根目录为 `parentFolderId: null`，默认图标为 `i-page`；名称按应用避重，未指定时可使用接口默认命名。
3. 创建成功后，从 **`result.id`** 取得并记录 `pageId`。
4. 调用 `POST /page/draft/save`，传该 `pageId`、首次基准版本 `baseVersionNo: 0` 和完整代码 `htmlContent`。创建成功但保存失败时，继续使用已有 `pageId`，不重复创建。

## 更新已有页面

1. 用户提供了 `pageId`，或从 `通过 AI 创建（pageId: xxx）` 进入时，直接更新这个页面，不调用创建接口。
2. 调用 `GET /page/get`，查询参数为 **`id`**，确认页面归属与草稿状态。
3. 已有草稿时调用 `GET /page/draft/source`，查询参数为 **`pageId`**，读取 `htmlContent` 和 **`versionNo`**，在原内容上按需求修改。确认从未保存过草稿时，基准版本为 `0`；读取失败不能当作没有草稿。
4. 调用 `POST /page/draft/save`，传原 `pageId`、**编辑前读取的 `versionNo` 作为 `baseVersionNo`**、完整 `htmlContent`。版本冲突时重新读取并合并修改，不仅替换版本号强行覆盖。

## 回传与结果

接口路径、参数、返回结构和本地 HTML 回传示例统一维护在 [页面接口与回传](page-api.md) 中，按其中的 `lingtong-cli api` 命令执行。

- 使用结构化 JSON 提交完整 HTML；检查业务响应 `success`，不能只看 CLI 退出码或 HTTP 状态。
- 保存后回读草稿，核对 `pageId`、版本和 HTML 内容。结果不明确时先查询确认，不直接重复创建或保存。
- 默认只保存草稿，不自动发布。
- 完成后告知本地 HTML 文件、`appId`、`pageId`、实际返回的草稿版本和保存结果；尚未成功的步骤如实说明。
