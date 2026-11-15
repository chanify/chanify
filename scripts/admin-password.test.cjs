// Run with: node --test scripts/admin-password.test.cjs
const { test } = require('node:test');
const assert = require('node:assert/strict');
const { createHash, webcrypto } = require('node:crypto');
const { readFileSync } = require('node:fs');
const vm = require('node:vm');
const source = readFileSync(new URL('../core/adminassets/password.js', `file://${__filename}`), 'utf8');
for (const native of [false, true]) {
  test(`admin password SHA-256 (${native ? 'Web Crypto' : 'HTTP fallback'})`, async () => {
    const context = vm.createContext({ crypto: native ? webcrypto : undefined });
    vm.runInContext(source, context);
    for (const value of ['', 'abc', 'synthetic-admin-password', '管理密码-🔔-synthetic', ...[55,56,63,64,65,72,128,1024].map(n => 'a'.repeat(n))]) {
      const input = new TextEncoder().encode(value), original = input.slice();
      const digest = await context.ChanifyPassword.digest(input);
      assert.equal(Buffer.from(digest).toString('hex'), createHash('sha256').update(input).digest('hex'));
      assert.deepEqual(input, original, 'caller retains ownership of input');
    }
  });
}

const app = readFileSync(new URL('../core/adminassets/app.js', `file://${__filename}`), 'utf8');
for (const allow of [true, false]) {
  test(`HTTP login integration (fallback ${allow ? 'enabled' : 'disabled'})`, async () => {
    const elements = new Map(), requests = [];
    const element = id => {
      if (!elements.has(id)) elements.set(id, { value: '', hidden: false, textContent: '', addEventListener(event, handler) { this[event] = handler; }, replaceChildren() {} });
      return elements.get(id);
    };
    const context = vm.createContext({
      URL, TextEncoder, Uint8Array, crypto: undefined,
      location: { href: 'http://node.internal/admin' },
      document: { getElementById: element, querySelectorAll: () => [] },
      fetch: async (url, options) => {
        requests.push({ path: url.pathname, ...options });
        if (url.pathname.endsWith('/config')) return { ok: true, json: async () => ({ allow_insecure_http: allow }) };
        return { ok: false, status: 401, json: async () => ({ error: 'synthetic rejection' }) };
      },
      adminPreferences: { t: value => value },
      createAdminMedia: () => ({ reset() {} }),
      addEventListener() {}
    });
    context.window = context;
    vm.runInContext(source, context);
    vm.runInContext(app, context);
    await new Promise(resolve => setImmediate(resolve));
    element('password').value = '管理密码-synthetic-123';
    const button = { disabled: false };
    await element('login-form').submit({ preventDefault() {}, submitter: button });
    const logins = requests.filter(r => r.path.endsWith('/login'));
    assert.equal(logins.length, allow ? 1 : 0);
    if (allow) {
      assert.deepEqual(JSON.parse(logins[0].body), { password_digest: createHash('sha256').update('管理密码-synthetic-123').digest('hex') });
      assert.equal(logins[0].credentials, 'same-origin');
    } else assert.equal(element('login-message').textContent, '请使用 HTTPS 或 localhost 打开管理页');
    assert.equal(element('password').value, '');
    assert.equal(button.disabled, false);
  });
}
