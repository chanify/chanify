'use strict';
(() => {
  const translations = {
  "节点管理": "Node administration",
  "管理你的通知节点": "Manage your notification node",
  "查看令牌、了解客户端使用情况，随时收回访问权限。": "Manage tokens, review client activity, and revoke access at any time.",
  "管理密码": "Admin password",
  "登录管理端": "Sign in",
  "此密码独立于发送令牌和 MCP 凭据。": "This password is separate from sending tokens and MCP credentials.",
  "退出登录": "Sign out",
  "工作区": "Workspace",
  "管理页面": "Administration",
  "发送令牌": "Sending tokens",
  "MCP 令牌": "MCP tokens",
  "客户端活动": "Client activity",
  "自建节点": "Self-hosted node",
  "访问管理": "Access management",
  "刷新": "Refresh",
  "＋ 创建 MCP 令牌": "＋ Create MCP token",
  "有效令牌": "Active tokens",
  "当前 SSE 监听": "Current SSE listeners",
  "已吊销": "Revoked",
  "节点概况": "Node overview",
  "令牌列表": "Tokens",
  "用户": "User",
  "全部用户": "All users",
  "暂无令牌": "No tokens yet",
  "创建 MCP 令牌": "Create MCP token",
  "关闭": "Close",
  "为一个 agent 创建独立凭据，方便单独吊销。": "Create a separate credential for each agent so you can revoke access individually.",
  "所属用户": "User",
  "选择用户": "Select a user",
  "名称": "Name",
  "例如：开发助手": "e.g. Development assistant",
  "有效期（天）": "Lifetime (days)",
  "可向该用户的频道发送。未指定频道时发送到未分类。": "Can send to this user's channels. Defaults to Uncategorized when no channel is specified.",
  "取消": "Cancel",
  "创建令牌": "Create token",
  "修改名称": "Rename token",
  "仅修改名称。令牌内容、权限与有效期不会改变。": "Only the name changes. Token contents, permissions, and expiration stay the same.",
  "保存名称": "Save name",
  "吊销令牌？": "Revoke token?",
  "后续请求将被拒绝。已提交的通知无法撤回；其他令牌不受影响。": "Future requests will be rejected. Submitted notifications cannot be recalled; other tokens are unaffected.",
  "确认吊销": "Revoke token",
  "令牌已创建": "Token created",
  "完整令牌仅展示这一次，请立即复制并妥善保存。": "The full token is shown only once. Copy it now and store it securely.",
  "复制令牌": "Copy token",
  "完成": "Done",
  "请求失败，请重试": "Request failed. Please try again.",
  " 台设备": " devices",
  "由 iOS 手动生成的服务器级发送令牌，可修改名称或单独吊销。": "Server-wide sending tokens created manually in iOS. Rename or revoke each token individually.",
  "为 coding agent 管理独立凭据，名称与有效期清晰可查。": "Manage separate credentials for coding agents with clear names and expiration dates.",
  "查看已认证请求的最近活动，以及当前保持的 SSE 监听。": "Review recent authenticated requests and current SSE listeners.",
  "更新于 ": "Updated ",
  "本次启动期间的活动": "Activity since startup",
  "仅记录本次启动期间的已登记发送令牌与 MCP 令牌的已认证请求，最多保留 2048 条。IP 为服务器直接连接地址；客户端名称由 User-Agent 提供。普通请求结束后仅表示最近使用。": "Records authenticated requests from registered sending and MCP tokens since startup, up to 2,048 entries. IP is the direct connection address; client names come from User-Agent. Completed requests indicate recent use only.",
  "列表仅展示元数据。发送令牌由 iOS 生成；完整 MCP 令牌仅在创建时展示。历史频道令牌不在此列表中。": "Only metadata is listed. Sending tokens are generated in iOS; full MCP tokens are shown only when created. Legacy channel tokens are not listed.",
  "客户端 / 来源": "Client / source",
  "令牌 / 用户": "Token / user",
  "最近使用": "Last used",
  "连接状态": "Connection status",
  "名称 / ID": "Name / ID",
  "创建时间": "Created",
  "过期时间": "Expires",
  "状态": "Status",
  "操作": "Actions",
  "未提供客户端名称": "Client name unavailable",
  "发送 · ": "Sending · ",
  "频道令牌": "Channel token",
  " 个 SSE 监听": " SSE listeners",
  "请求处理中": "Request in progress",
  "始于 ": "Since ",
  "已过期": "Expired",
  "有效": "Active",
  "改名": "Rename",
  "吊销": "Revoke",
  "确认吊销「": "Revoke “",
  "」？此操作无法撤销。": "”? This cannot be undone.",
  "暂无客户端活动": "No client activity yet",
  "客户端发送请求后，会在这里显示最近使用情况。": "Recent activity appears here after a client sends a request.",
  "创建一个 MCP 令牌，开始连接你的 coding agent。": "Create an MCP token to connect your coding agent.",
  "在 iOS 节点详情中添加发送令牌后，会在这里显示。": "Sending tokens appear here after you add them in the iOS node details.",
  "请使用 HTTPS 或 localhost 打开管理页": "Open this page over HTTPS or localhost.",
  "已复制": "Copied",
  "请复制已选中的令牌": "Copy the selected token.",
  "请先登录": "Please sign in.",
  "请求来源无效": "Invalid request origin.",
  "请输入有效密码": "Enter a valid password.",
  "尝试次数过多，请稍后重试": "Too many attempts. Try again later.",
  "登录繁忙，请稍后重试": "Sign-in is busy. Try again later.",
  "密码不正确": "Incorrect password.",
  "无法创建会话": "Unable to create a session.",
  "会话数量已达上限": "Session limit reached.",
  "登录已过期，请重新登录": "Your session expired. Please sign in again.",
  "请求校验失败，请刷新页面": "Request verification failed. Refresh the page.",
  "无法读取管理数据": "Unable to load administration data.",
  "参数无效": "Invalid parameters.",
  "创建失败，请检查用户、名称和有效期": "Creation failed. Check the user, name, and lifetime.",
  "不支持此操作": "This action is not supported.",
  "操作失败，令牌可能已失效": "Action failed. The token may no longer be valid.",
  "语言": "Language",
  "主题": "Theme",
  "自动": "Auto",
  "浅色": "Light",
  "深色": "Dark",
  "媒体管理": "Media management",
  "查找、下载服务器附件，清理存储空间。": "Find and download server attachments, and free up storage space.",
  "文件总数": "Total files",
  "总占用空间": "Storage used",
  "筛选结果": "Filtered results",
  "搜索文件名或 ID": "Search filename or ID",
  "搜索": "Search",
  "文件类型": "File type",
  "全部类型": "All types",
  "文本": "Text",
  "图片": "Images",
  "视频": "Video",
  "音频": "Audio",
  "其他文件": "Other files",
  "排序": "Sort",
  "最新优先": "Newest first",
  "最早优先": "Oldest first",
  "大文件优先": "Largest first",
  "按名称排序": "Name",
  "删除选中": "Delete selected",
  "清空全部文件": "Clear all files",
  "选择当前页": "Select current page",
  "文件名 / ID": "Filename / ID",
  "大小": "Size",
  "时间": "Time",
  "没有匹配的文件": "No matching files",
  "上一页": "Previous",
  "下一页": "Next",
  "仅管理 Server 文件。删除后无法再从 Server 下载，不会清除客户端已有副本。相同内容可能由多条通知共用。旧文件仅展示已有信息，时间为文件修改时间。": "Manages server files only. Deleted files can no longer be downloaded from this server; existing client copies remain. Multiple notifications may share the same file. Legacy files show available details and filesystem modification time.",
  "文本预览最多显示 256 KiB；完整内容请下载。音视频播放取决于浏览器支持。": "Text previews show up to 256 KiB. Download for the full contents. Audio and video playback depends on browser support.",
  "下载": "Download",
  "删除服务器文件？": "Delete server files?",
  "此操作无法撤销。相关通知的附件将无法再从 Server 下载。": "This cannot be undone. Affected notification attachments can no longer be downloaded from this server.",
  "输入 CLEAR 确认清空": "Type CLEAR to confirm clearing all files",
  "确认删除": "Delete files",
  "已选当前页：": "Selected on this page: ",
  "选择文件：": "Select file: ",
  "旧文件": "Legacy file",
  "上传时间": "Uploaded",
  "文件修改时间": "File modified",
  "预览": "Preview",
  "删除": "Delete",
  "每页最多 50 个文件": "Up to 50 files per page",
  "无法预览文件": "Unable to preview this file.",
  "此格式不支持预览，请下载查看": "Preview unavailable for this format. Download to view it.",
  "即将清空服务器全部文件：": "Clear ALL server files: ",
  "即将删除文件：": "Delete files: ",
  "已删除：": "Deleted: ",
  "已释放：": "Freed: ",
  "删除失败：": "Failed to delete: ",
  "无法读取文件列表": "Unable to load files.",
  "文件列表已变化，请刷新后重新确认": "Files have changed. Refresh and confirm again.",
  "无法删除文件": "Unable to delete files.",
  "文件分页": "File pagination",
  "第 {start}–{end} 项，共 {total} 项": "Showing {start}–{end} of {total} files",
  "第 {page} 页": "Page {page}"
};
  const read = key => { try { return localStorage.getItem(key); } catch (_) { return null; } };
  const save = (key, value) => { try { localStorage.setItem(key, value); } catch (_) {} };
  const storedLanguage = read('chanify.admin.language');
  const browserLanguage = ((navigator.languages || [navigator.language]).find(Boolean) || 'en').toLowerCase();
  let language = ['zh-CN', 'en'].includes(storedLanguage) ? storedLanguage : (browserLanguage.startsWith('zh') ? 'zh-CN' : 'en');
  let theme = read('chanify.admin.theme');
  if (!['auto', 'light', 'dark'].includes(theme)) theme = 'auto';
  const media = matchMedia('(prefers-color-scheme: dark)');
  const t = text => language === 'en' ? translations[text] || text : text;
  function applyTheme() {
    document.querySelectorAll('[data-theme-choice]').forEach(button => button.setAttribute('aria-pressed', String(button.dataset.themeChoice === theme)));
    document.documentElement.dataset.theme = theme === 'auto' ? (media.matches ? 'dark' : 'light') : theme;
  }
  applyTheme();
  media.addEventListener('change', applyTheme);
  window.adminPreferences = { t, get language() { return language; }, translateMessage(text) {
    const key = Object.keys(translations).find(key => key === text || translations[key] === text);
    return key ? t(key) : text;
  } };
  function applyLanguage() {
    document.documentElement.lang = language;
    document.title = 'Chanify · ' + t('节点管理');
    document.querySelectorAll('[data-i18n]').forEach(el => { el.textContent = t(el.dataset.i18n); });
    ['aria-label', 'placeholder', 'title'].forEach(attr => document.querySelectorAll('[data-i18n-' + attr + ']').forEach(el => el.setAttribute(attr, t(el.getAttribute('data-i18n-' + attr)))));
    document.querySelectorAll('[data-language]').forEach(button => button.setAttribute('aria-pressed', String(button.dataset.language === language)));
    window.dispatchEvent(new Event('admin-language-change'));
  }
  document.addEventListener('DOMContentLoaded', () => {
    applyLanguage();
    applyTheme();
    document.querySelectorAll('[data-language]').forEach(button => button.addEventListener('click', () => { language = button.dataset.language; save('chanify.admin.language', language); applyLanguage(); }));
    document.querySelectorAll('[data-theme-choice]').forEach(button => button.addEventListener('click', () => { theme = button.dataset.themeChoice; save('chanify.admin.theme', theme); applyTheme(); }));
  });
})();
