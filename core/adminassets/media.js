'use strict';
window.createAdminMedia = ({ api, endpoint, t, date }) => {
  const $ = id => document.getElementById(id);
  let listing = null, selected = new Set(), page = 1, serial = 0, previewSerial = 0, pending = null, busy = false, sessionGeneration = 0, lastResult = null, loading = false;
  const kindLabels = { text: '文本', image: '图片', video: '视频', audio: '音频', file: '其他文件' };
  const bytes = size => {
    const unit = size < 1024 ? 0 : Math.min(4, Math.floor(Math.log(size) / Math.log(1024)));
    return new Intl.NumberFormat(window.adminPreferences.language, { maximumFractionDigits: unit ? 1 : 0 }).format(size / 1024 ** unit) + ' ' + ['B', 'KiB', 'MiB', 'GiB', 'TiB'][unit];
  };
  const name = item => item.names?.at(-1) || item.id.split('/')[1];
  function node(tag, text, className) {
    const el = document.createElement(tag); if (text !== undefined) el.textContent = text;
    if (className) el.className = className; return el;
  }
  function url(item, preview = false) {
    const u = new URL('media/content', endpoint); u.searchParams.set('id', item.id);
    if (preview) u.searchParams.set('preview', '1'); return u.href;
  }
  function selection() {
    const items = listing?.items || [], chosen = items.filter(item => selected.has(item.id));
    $('media-selection').textContent = t('已选当前页：') + chosen.length + ' · ' + bytes(chosen.reduce((sum, item) => sum + item.size, 0));
    $('media-delete-selected').disabled = busy || loading || !chosen.length;
    $('media-select-page').checked = !!items.length && chosen.length === items.length;
    $('media-select-page').indeterminate = chosen.length > 0 && chosen.length < items.length;
    $('media-select-page').disabled = busy || loading || !items.length;
  }
  function render() {
    $('media-result').hidden = !lastResult;
    $('media-result').classList.toggle('has-failures', !!lastResult?.failed.length);
    if (lastResult) $('media-result').textContent = t('已删除：') + lastResult.deleted + ' · ' + t('已释放：') + bytes(lastResult.bytes) + (lastResult.failed.length ? ' · ' + t('删除失败：') + lastResult.failed.join(', ') : '');
    $('media-count').textContent = listing?.count ?? '—';
    $('media-bytes').textContent = listing ? bytes(listing.bytes) : '—';
    $('media-matched').textContent = listing ? listing.total + ' · ' + bytes(listing.filtered_bytes) : '—';
    $('media-body').replaceChildren();
    for (const item of listing?.items || []) {
      const row = node('tr'); const cell = () => { const el = node('td'); row.append(el); return el; };
      const checkbox = node('input'); checkbox.type = 'checkbox'; checkbox.checked = selected.has(item.id); checkbox.setAttribute('aria-label', t('选择文件：') + name(item));
      checkbox.addEventListener('change', () => { checkbox.checked ? selected.add(item.id) : selected.delete(item.id); selection(); }); cell().append(checkbox);
      const title = cell(); title.append(node('div', name(item), 'cell-title media-name'), node('small', item.id, 'cell-sub mono media-file-id'));
      if (item.names?.length > 1) title.append(node('div', item.names.slice(0, -1).join(' · '), 'cell-sub media-name'));
      if (item.legacy) title.append(node('span', t('旧文件'), 'badge revoked'));
      cell().textContent = t(kindLabels[item.kind] || '其他文件'); cell().textContent = bytes(item.size);
      const time = cell(); time.append(node('div', date(item.uploaded || item.modified)), node('small', t(item.uploaded ? '上传时间' : '文件修改时间'), 'cell-sub'));
      const actions = node('div', undefined, 'media-actions');
      if (!item.legacy && ['text', 'image', 'audio', 'video'].includes(item.kind)) {
        const preview = node('button', t('预览'), 'action-link'); preview.type = 'button'; preview.addEventListener('click', () => openPreview(item)); actions.append(preview);
      }
      const download = node('a', t('下载'), 'download-link'); download.href = url(item); download.setAttribute('download', name(item)); actions.append(download);
      const remove = node('button', t('删除'), 'danger-link'); remove.type = 'button'; remove.disabled = busy || loading; remove.addEventListener('click', () => confirmDelete([item], false)); actions.append(remove); cell().append(actions);
      $('media-body').append(row);
    }
    $('media-empty').hidden = !listing || listing.items.length !== 0;
    $('media-clear').disabled = busy || loading || !listing?.count;
    renderPagination();
    selection();
  }
  const format = (key, values) => Object.entries(values).reduce((text, [name, value]) => text.replaceAll('{' + name + '}', new Intl.NumberFormat(window.adminPreferences.language).format(value)), t(key));
  function renderPagination() {
    const current = listing?.page || 1, pages = listing?.pages || 1;
    $('media-prev').disabled = busy || loading || !listing || current <= 1;
    $('media-next').disabled = busy || loading || !listing || current >= pages;
    $('media-pagination-controls').hidden = !listing || pages <= 1;
    $('media-page-label').textContent = listing ? format('第 {start}–{end} 项，共 {total} 项', { start: listing.total ? (current - 1) * 50 + 1 : 0, end: Math.min(current * 50, listing.total), total: listing.total }) : '';
    const numbers = $('media-pages'); numbers.replaceChildren();
    if (!listing || pages <= 1) return;
    // Keep a compact sliding window while leaving the first and last pages reachable.
    const shown = pages <= 7 ? Array.from({ length: pages }, (_, i) => i + 1) : [...new Set([1, pages, ...Array.from({ length: 3 }, (_, i) => Math.min(pages - 2, Math.max(3, current)) + i - 1)])].sort((a, b) => a - b);
    let previous = 0;
    for (const value of shown) {
      if (previous && value - previous > 1) {
        if (value - previous === 2) appendPage(previous + 1);
        else { const gap = node('span', '…', 'page-gap'); gap.setAttribute('aria-hidden', 'true'); numbers.append(gap); }
      }
      appendPage(value); previous = value;
    }
    function appendPage(value) {
      const button = node('button', new Intl.NumberFormat(window.adminPreferences.language).format(value), 'page-number');
      button.type = 'button'; button.disabled = busy || loading;
      button.setAttribute('aria-label', format('第 {page} 页', { page: value }));
      if (value === current) button.setAttribute('aria-current', 'page');
      button.addEventListener('click', () => goTo(value, true)); numbers.append(button);
    }
  }
  async function goTo(target, focusNumber = false) {
    if (busy || loading || !listing || target < 1 || target > listing.pages || target === listing.page) return;
    page = target; await refresh();
    if (focusNumber && listing?.page === target) $('media-pages').querySelector('[aria-current="page"]')?.focus();
  }
  async function refresh() {
    const request = ++serial; loading = true; render();
    const query = new URLSearchParams({ q: $('media-search').value.trim(), kind: $('media-kind').value, sort: $('media-sort').value, page: String(page) });
    $('media-view').setAttribute('aria-busy', 'true');
    try {
      const result = await api('media?' + query);
      if (request !== serial) return;
      listing = result; page = result.page; selected.clear(); $('page-message').textContent = ''; render();
    } catch (error) { if (request === serial) $('page-message').textContent = t(error instanceof TypeError ? '请求失败，请重试' : error.message); }
    finally { if (request === serial) { loading = false; $('media-view').removeAttribute('aria-busy'); render(); } }
  }
  async function openPreview(item) {
    const request = ++previewSerial; const content = $('media-preview-content'); content.replaceChildren();
    $('media-preview-title').textContent = name(item); $('media-preview-download').href = url(item); $('media-preview-download').setAttribute('download', name(item));
    $('media-preview-dialog').showModal();
    try {
      // Check the real content type: names and uploaded MIME hints never authorize active content.
      const response = await fetch(url(item, true), { credentials: 'same-origin', cache: 'no-store', headers: item.kind === 'text' ? {} : { Range: 'bytes=0-511' } });
      if (!response.ok) { await response.body?.cancel(); if (response.status === 401) await api('session'); throw new Error(t('无法预览文件')); }
      const mime = response.headers.get('Content-Type') || '';
      if (request !== previewSerial) { await response.body?.cancel(); return; }
      if (mime.startsWith('text/plain')) {
        const text = await response.text(); if (request === previewSerial) content.append(node('pre', text));
      } else {
        await response.body?.cancel();
        if (!/^(image\/(png|jpeg|gif|webp|bmp|x-icon)|audio\/|video\/|application\/ogg)/.test(mime)) throw new Error(t('此格式不支持预览，请下载查看'));
        const tag = mime.startsWith('image/') ? 'img' : mime.startsWith('video/') ? 'video' : 'audio';
        const el = node(tag); if (tag === 'img') el.alt = name(item); else { el.controls = true; el.preload = 'metadata'; }
        el.addEventListener('error', () => { if (request === previewSerial) content.replaceChildren(node('p', t('此格式不支持预览，请下载查看'), 'muted')); });
        el.src = url(item, true); content.append(el);
      }
    } catch (error) { if (request === previewSerial) content.replaceChildren(node('p', t(error instanceof TypeError ? '请求失败，请重试' : error.message), 'muted')); }
  }
  $('media-preview-dialog').addEventListener('close', () => {
    ++previewSerial; $('media-preview-content').querySelectorAll('audio,video').forEach(el => { el.pause(); el.removeAttribute('src'); el.load(); }); $('media-preview-content').replaceChildren();
  });
  function confirmDelete(items, clear) {
    if (busy || loading || !listing) return;
    pending = { action: clear ? 'clear' : 'delete', ...(clear ? { snapshot: listing.snapshot } : { ids: items.map(item => item.id) }), count: clear ? listing.count : items.length, bytes: clear ? listing.bytes : items.reduce((sum, item) => sum + item.size, 0) };
    renderConfirmation(); $('media-clear-confirmation').hidden = !clear;
    $('media-clear-input').value = ''; $('media-clear-input').required = clear;
    $('media-delete-dialog').querySelector('.dialog-message').textContent = ''; $('media-delete-dialog').showModal();
  }
  function renderConfirmation() { if (pending) $('media-delete-description').textContent = t(pending.action === 'clear' ? '即将清空服务器全部文件：' : '即将删除文件：') + pending.count + ' · ' + bytes(pending.bytes); }
  $('media-delete-form').addEventListener('submit', async event => {
    event.preventDefault(); if (!pending || busy) return;
    if (pending.action === 'clear' && $('media-clear-input').value !== 'CLEAR') { $('media-clear-input').focus(); return; }
    const dialog = $('media-delete-dialog'); const request = pending.action === 'clear' ? { action: 'clear', snapshot: pending.snapshot } : { action: 'delete', ids: pending.ids };
    const generation = sessionGeneration; busy = true; dialog.dataset.busy = 'true'; $('media-confirm-delete').disabled = true; render();
    try {
      const result = await api('media', request);
      if (generation !== sessionGeneration) return;
      dialog.close(); pending = null;
      lastResult = result; $('page-message').textContent = '';
      await refresh();
    } catch (error) { dialog.querySelector('.dialog-message').textContent = t(error instanceof TypeError ? '请求失败，请重试' : error.message); }
    finally { busy = false; delete dialog.dataset.busy; $('media-confirm-delete').disabled = false; render(); }
  });
  $('media-delete-selected').addEventListener('click', () => confirmDelete((listing?.items || []).filter(item => selected.has(item.id)), false));
  $('media-clear').addEventListener('click', () => confirmDelete([], true));
  $('media-select-page').addEventListener('change', event => { selected = new Set(event.target.checked ? (listing?.items || []).map(item => item.id) : []); render(); });
  $('media-search-form').addEventListener('submit', event => { event.preventDefault(); page = 1; refresh(); });
  ['media-kind', 'media-sort'].forEach(id => $(id).addEventListener('change', () => { page = 1; refresh(); }));
  $('media-prev').addEventListener('click', () => goTo(page - 1)); $('media-next').addEventListener('click', () => goTo(page + 1));
  window.addEventListener('admin-language-change', () => { render(); renderConfirmation(); });
  return { refresh, render, reset() { ++sessionGeneration; ++serial; ++previewSerial; listing = null; selected.clear(); page = 1; pending = null; lastResult = null; loading = false; $('media-view').removeAttribute('aria-busy'); render(); } };
};
