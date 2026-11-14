'use strict';
(() => {
  const { t } = window.adminPreferences;
  const $ = id => document.getElementById(id);
  const endpoint = new URL('admin/api/', location.href);
  let csrf = '', state = null, view = 'sending', selected = null;
  const date = seconds => seconds ? new Date(seconds * 1000).toLocaleString(window.adminPreferences.language) : '—';
  const short = value => value.length > 18 ? value.slice(0, 10) + '…' + value.slice(-6) : value;
  function node(tag, text, className) {
    const el = document.createElement(tag);
    if (text !== undefined) el.textContent = text;
    if (className) el.className = className;
    return el;
  }
  function loginView() {
    csrf = ''; state = null; selected = null;
    document.querySelectorAll('dialog[open]').forEach(dialog => dialog.close());
    $('new-token').value = ''; $('table-body').replaceChildren();
    $('app-view').hidden = true; $('login-view').hidden = false;
  }
  async function api(path, body) {
    const response = await fetch(new URL(path, endpoint), {
      method: body === undefined ? 'GET' : 'POST', credentials: 'same-origin', cache: 'no-store',
      headers: body === undefined ? {} : { 'Content-Type': 'application/json', 'X-CSRF-Token': csrf },
      body: body === undefined ? undefined : JSON.stringify(body)
    });
    const result = await response.json();
    if (!response.ok) {
      if (response.status === 401 && path !== 'login') loginView();
      throw new Error(t(result.error || '请求失败，请重试'));
    }
    return result;
  }
  function options(select, users, placeholder) {
    const previous = select.value;
    select.replaceChildren(new Option(placeholder, ''));
    users.forEach(user => select.add(new Option(user.id + ' · ' + user.devices + t(" 台设备"), user.id)));
    if (users.some(user => user.id === previous)) select.value = previous;
  }
  async function refresh() {
    $('refresh').disabled = true;
    try {
      const snapshot = await api('state');
      if (!csrf) return;
      state = snapshot;
      options($('user-filter'), state.users || [], t("全部用户"));
      options($('create-user'), state.users || [], t("选择用户"));
      $('node-name').textContent = state.node.name || 'Chanify';
      $('node-version').textContent = state.node.version || '';
      $('page-message').textContent = '';
      render();
    } catch (error) { $('page-message').textContent = t(error instanceof TypeError ? '请求失败，请重试' : error.message); }
    finally { $('refresh').disabled = false; }
  }
  function action(label, token, callback, destructive) {
    const button = node('button', label, destructive ? 'danger-link' : 'action-link');
    button.addEventListener('click', () => { selected = token; callback(); });
    return button;
  }
  function openDialog(id) {
    const dialog = $(id);
    dialog.querySelector('.dialog-message').textContent = '';
    dialog.showModal();
  }
  function render() {
    if (!state) return;
    const tokens = state.tokens || [], clients = state.clients || [], now = state.now;
    const user = $('user-filter').value;
    const matches = item => !user || item.user === user;
    const names = { sending: t("发送令牌"), mcp: t("MCP 令牌"), clients: t("客户端活动") };
    const descriptions = {
      sending: t("由 iOS 手动生成的服务器级发送令牌，可修改名称或单独吊销。"),
      mcp: t("为 coding agent 管理独立凭据，名称与有效期清晰可查。"),
      clients: t("查看已认证请求的最近活动，以及当前保持的 SSE 监听。")
    };
    $('page-title').textContent = names[view]; $('page-description').textContent = descriptions[view];
    $('create-mcp').hidden = view !== 'mcp';
    $('sending-count').textContent = tokens.filter(t => t.kind === 'sending' && !t.revoked && t.expires > now).length;
    $('mcp-count').textContent = tokens.filter(t => t.kind === 'mcp' && !t.revoked && t.expires > now).length;
    $('client-count').textContent = clients.length;
    $('active-count').textContent = tokens.filter(t => matches(t) && !t.revoked && t.expires > now).length;
    $('revoked-count').textContent = tokens.filter(t => matches(t) && t.revoked).length;
    $('listener-count').textContent = clients.filter(matches).reduce((total, c) => total + c.listeners, 0);
    $('updated-at').textContent = t("更新于 ") + date(now);
    $('table-title').textContent = view === 'clients' ? t("本次启动期间的活动") : t("令牌列表");
    $('scope-note').textContent = view === 'clients'
      ? t("仅记录本次启动期间的已登记发送令牌与 MCP 令牌的已认证请求，最多保留 2048 条。IP 为服务器直接连接地址；客户端名称由 User-Agent 提供。普通请求结束后仅表示最近使用。")
      : t("列表仅展示元数据。发送令牌由 iOS 生成；完整 MCP 令牌仅在创建时展示。历史频道令牌不在此列表中。");
    document.querySelectorAll('[data-view]').forEach(button => button.classList.toggle('active', button.dataset.view === view));
    const heads = view === 'clients' ? [t("客户端 / 来源"), t("令牌 / 用户"), t("最近使用"), t("连接状态")] : [t("名称 / ID"), t("用户"), t("创建时间"), t("过期时间"), t("最近使用"), t("状态"), t("操作")];
    const heading = node('tr'); heads.forEach(text => heading.append(node('th', text)));
    $('table-head').replaceChildren(heading); $('table-body').replaceChildren();
    const items = (view === 'clients' ? clients : tokens.filter(t => t.kind === view)).filter(matches);
    items.forEach(item => {
      const row = node('tr');
      const cell = (primary, secondary) => {
        const td = node('td'); td.append(node('div', primary, 'cell-title'));
        if (secondary) td.append(node('small', secondary, 'cell-sub'));
        row.append(td); return td;
      };
      if (view === 'clients') {
        cell(item.user_agent || t("未提供客户端名称"), item.ip);
        cell((item.kind === 'mcp' ? 'MCP · ' : t("发送 · ")) + (short(item.credential_id) || t("频道令牌")), short(item.user));
        cell(date(item.last_seen));
        const td = cell('');
        td.append(node('span', item.listeners ? item.listeners + t(" 个 SSE 监听") : item.active_requests ? t("请求处理中") : t("最近使用"), item.listeners ? 'badge' : 'badge revoked'));
        if (item.connected_since) td.append(node('div', t("始于 ") + date(item.connected_since), 'cell-sub'));
      } else {
        cell(item.name, short(item.id)); cell(short(item.user)); cell(date(item.created)); cell(date(item.expires));
        const activity = clients.filter(c => c.kind === item.kind && c.credential_id === item.id && c.user === item.user);
        cell(date(Math.max(0, ...activity.map(c => c.last_seen))));
        const status = cell('');
        status.append(node('span', item.revoked ? t("已吊销") : item.expires <= now ? t("已过期") : t("有效"), item.revoked ? 'badge revoked' : item.expires <= now ? 'badge expired' : 'badge'));
        const actions = cell('');
        if (!item.revoked) {
          actions.append(action(t("改名"), item, () => { $('rename-name').value = item.name; openDialog('rename-dialog'); }));
          actions.append(action(t("吊销"), item, () => { $('revoke-description').textContent = t("确认吊销「") + item.name + t("」？此操作无法撤销。"); openDialog('revoke-dialog'); }, true));
        } else actions.textContent = '—';
      }
      $('table-body').append(row);
    });
    $('empty').hidden = items.length !== 0;
    $('empty-title').textContent = view === 'clients' ? t("暂无客户端活动") : t("暂无令牌");
    $('empty-description').textContent = view === 'clients' ? t("客户端发送请求后，会在这里显示最近使用情况。") : view === 'mcp' ? t("创建一个 MCP 令牌，开始连接你的 coding agent。") : t("在 iOS 节点详情中添加发送令牌后，会在这里显示。");
  }
  $('login-form').addEventListener('submit', async event => {
    event.preventDefault(); const button = event.submitter; button.disabled = true;
    $('login-message').textContent = '';
    try {
      if (!globalThis.crypto || !globalThis.crypto.subtle) throw new Error(t("请使用 HTTPS 或 localhost 打开管理页"));
      const input = new TextEncoder().encode($('password').value);
      $('password').value = '';
      let digest;
      try { digest = await globalThis.crypto.subtle.digest('SHA-256', input); }
      finally { input.fill(0); }
      const passwordDigest = Array.from(new Uint8Array(digest), byte => byte.toString(16).padStart(2, '0')).join('');
      const session = await api('login', { password_digest: passwordDigest });
      csrf = session.csrf; $('login-view').hidden = true; $('app-view').hidden = false;
      await refresh();
    } catch (error) { $('login-message').textContent = t(error instanceof TypeError ? '请求失败，请重试' : error.message); }
    finally { $('password').value = ''; button.disabled = false; }
  });
  document.querySelectorAll('[data-action="logout"]').forEach(button => button.addEventListener('click', async () => {
    try { await api('logout', {}); loginView(); }
    catch (error) { $('page-message').textContent = t(error instanceof TypeError ? '请求失败，请重试' : error.message); }
  }));
  $('refresh').addEventListener('click', refresh);
  $('user-filter').addEventListener('change', render);
  document.querySelectorAll('[data-view]').forEach(button => button.addEventListener('click', () => { view = button.dataset.view; render(); }));
  document.querySelectorAll('.close-dialog').forEach(button => button.addEventListener('click', () => { const dialog = button.closest('dialog'); if (!dialog.dataset.busy) dialog.close(); }));
  document.querySelectorAll('dialog').forEach(dialog => dialog.addEventListener('cancel', event => { if (dialog.dataset.busy) event.preventDefault(); }));
  $('secret-dialog').addEventListener('close', () => { $('new-token').value = ''; });
  $('create-mcp').addEventListener('click', () => {
    $('create-user').value = $('user-filter').value; $('create-name').value = ''; $('create-days').value = 90; openDialog('create-dialog');
  });
  async function mutate(event, body, after) {
    event.preventDefault(); const dialog = event.target.closest('dialog');
    if (dialog.dataset.busy) return;
    dialog.dataset.busy = 'true'; event.submitter.disabled = true;
    try { const result = await api('tokens', body); dialog.close(); after(result); await refresh(); }
    catch (error) { dialog.querySelector('.dialog-message').textContent = t(error instanceof TypeError ? '请求失败，请重试' : error.message); }
    finally { delete dialog.dataset.busy; event.submitter.disabled = false; }
  }
  $('create-form').addEventListener('submit', event => mutate(event, { action: 'create', kind: 'mcp', user: $('create-user').value, name: $('create-name').value.trim(), expires: Math.floor(Date.now() / 1000) + Number($('create-days').value) * 86400 }, result => {
    $('new-token').value = result.token; openDialog('secret-dialog');
  }));
  ['rename', 'revoke'].forEach(actionName => $(actionName + '-form').addEventListener('submit', event => mutate(event, { action: actionName, kind: selected.kind, user: selected.user, id: selected.id, name: actionName === 'rename' ? $('rename-name').value.trim() : '' }, () => {})));
  $('copy-token').addEventListener('click', async () => {
    try { await navigator.clipboard.writeText($('new-token').value); $('secret-dialog').querySelector('.dialog-message').textContent = t("已复制"); }
    catch (_) { $('new-token').select(); $('secret-dialog').querySelector('.dialog-message').textContent = t("请复制已选中的令牌"); }
  });
  window.addEventListener('admin-language-change', () => {
    document.querySelectorAll('.message').forEach(el => { el.textContent = window.adminPreferences.translateMessage(el.textContent); });
    if (state) {
      options($('user-filter'), state.users || [], t('全部用户'));
      options($('create-user'), state.users || [], t('选择用户'));
      render();
    }
    if (selected) $('revoke-description').textContent = t('确认吊销「') + selected.name + t('」？此操作无法撤销。');
  });
  api('session').then(session => { csrf = session.csrf; $('login-view').hidden = true; $('app-view').hidden = false; return refresh(); }).catch(() => loginView());
})();
