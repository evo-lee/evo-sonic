/**
 * AI Panel — injected into the Sonic admin SPA.
 * Activates on post/sheet editor routes, adds a floating sidebar
 * with Summarize, Polish, and Suggest Tags actions backed by SSE streaming.
 */
(function () {
  'use strict';

  var EDITOR_ROUTES = ['#/posts/write', '#/posts/edit', '#/sheets/write'];
  var panel = null;
  var currentRoute = '';

  // ── Token ──────────────────────────────────────────────────────────────────

  function getToken() {
    try {
      var raw = localStorage.getItem('Access-Token');
      if (!raw) return '';
      var obj = JSON.parse(raw);
      return obj && obj.access_token ? obj.access_token : '';
    } catch (_) {
      return '';
    }
  }

  // ── Editor content ─────────────────────────────────────────────────────────

  function getEditorContent() {
    var cm = document.querySelector('.CodeMirror');
    if (cm && cm.CodeMirror) return cm.CodeMirror.getValue();
    var ta = document.querySelector('textarea.editor-content');
    if (ta) return ta.value;
    return '';
  }

  function getPostTitle() {
    var el = document.querySelector('input.ant-input[placeholder]');
    if (el && el.value) return el.value;
    return '';
  }

  // ── Tag injection ──────────────────────────────────────────────────────────

  function appendTags(tagStr) {
    var tags = tagStr.split(',').map(function (t) { return t.trim(); }).filter(Boolean);
    // Ant Design Select component — simulate typing into the select input
    var tagInput = document.querySelector('.ant-select-search__field, .ant-select-selection-search-input');
    if (!tagInput) return;
    tags.forEach(function (tag) {
      var nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
      nativeInputValueSetter.call(tagInput, tag);
      tagInput.dispatchEvent(new Event('input', { bubbles: true }));
      // Simulate Enter to confirm tag
      tagInput.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', keyCode: 13, bubbles: true }));
    });
  }

  // ── SSE helper ─────────────────────────────────────────────────────────────

  function streamRequest(endpoint, body, onChunk, onDone, onError) {
    var token = getToken();
    fetch('/api/admin/ai/' + endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Admin-Authorization': token,
      },
      body: JSON.stringify(body),
    }).then(function (res) {
      var reader = res.body.getReader();
      var decoder = new TextDecoder();
      var buffer = '';

      function read() {
        reader.read().then(function (result) {
          if (result.done) { onDone(); return; }
          buffer += decoder.decode(result.value, { stream: true });
          var lines = buffer.split('\n');
          buffer = lines.pop();
          lines.forEach(function (line) {
            if (line.startsWith('data: ')) {
              var data = line.slice(6).trim();
              if (data === '[DONE]') { onDone(); return; }
              try {
                var obj = JSON.parse(data);
                if (obj.error) { onError(obj.error); }
                else if (obj.text) { onChunk(obj.text); }
              } catch (_) {}
            } else if (line.startsWith('event: error')) {
              // handled by next data line
            }
          });
          read();
        }).catch(function (err) { onError(err.message || 'Stream error'); });
      }
      read();
    }).catch(function (err) { onError(err.message || 'Network error'); });
  }

  // ── UI ─────────────────────────────────────────────────────────────────────

  var PANEL_CSS = `
    #ai-panel-btn {
      position: fixed;
      right: 16px;
      bottom: 80px;
      z-index: 9999;
      background: #1890ff;
      color: #fff;
      border: none;
      border-radius: 24px;
      padding: 8px 16px;
      font-size: 13px;
      cursor: pointer;
      box-shadow: 0 2px 8px rgba(0,0,0,.25);
      transition: background .2s;
    }
    #ai-panel-btn:hover { background: #096dd9; }
    #ai-panel-drawer {
      position: fixed;
      right: 0; top: 0; bottom: 0;
      width: 340px;
      z-index: 10000;
      background: #fff;
      box-shadow: -2px 0 12px rgba(0,0,0,.15);
      display: flex;
      flex-direction: column;
      transform: translateX(100%);
      transition: transform .25s ease;
    }
    #ai-panel-drawer.open { transform: translateX(0); }
    .ai-drawer-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 16px;
      border-bottom: 1px solid #f0f0f0;
      font-weight: 600;
      font-size: 15px;
    }
    .ai-drawer-close {
      background: none;
      border: none;
      font-size: 18px;
      cursor: pointer;
      color: #999;
      line-height: 1;
    }
    .ai-drawer-close:hover { color: #333; }
    .ai-drawer-body {
      flex: 1;
      overflow-y: auto;
      padding: 16px;
      display: flex;
      flex-direction: column;
      gap: 12px;
    }
    .ai-action-btn {
      width: 100%;
      padding: 10px 14px;
      background: #f5f5f5;
      border: 1px solid #d9d9d9;
      border-radius: 6px;
      cursor: pointer;
      text-align: left;
      font-size: 13px;
      display: flex;
      align-items: center;
      gap: 8px;
      transition: background .15s;
    }
    .ai-action-btn:hover { background: #e6f4ff; border-color: #91caff; }
    .ai-action-btn:disabled { opacity: .5; cursor: not-allowed; }
    .ai-result-box {
      background: #fafafa;
      border: 1px solid #e8e8e8;
      border-radius: 6px;
      padding: 10px 12px;
      font-size: 12px;
      line-height: 1.6;
      color: #333;
      min-height: 60px;
      max-height: 300px;
      overflow-y: auto;
      white-space: pre-wrap;
      word-break: break-word;
    }
    .ai-result-label {
      font-size: 11px;
      color: #999;
      margin-bottom: 4px;
    }
    .ai-copy-btn {
      background: none;
      border: 1px solid #d9d9d9;
      border-radius: 4px;
      padding: 2px 8px;
      font-size: 11px;
      cursor: pointer;
      margin-top: 6px;
      color: #666;
    }
    .ai-copy-btn:hover { background: #f0f0f0; }
    .ai-section { display: flex; flex-direction: column; gap: 6px; }
    .ai-divider { border: none; border-top: 1px solid #f0f0f0; margin: 4px 0; }
    .ai-settings-link {
      font-size: 12px;
      color: #1890ff;
      text-decoration: none;
      text-align: center;
      padding: 8px;
      cursor: pointer;
    }
    .ai-settings-link:hover { text-decoration: underline; }
    .ai-modal-overlay {
      position: fixed;
      inset: 0;
      background: rgba(0,0,0,.45);
      z-index: 11000;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .ai-modal {
      background: #fff;
      border-radius: 8px;
      padding: 24px;
      width: 420px;
      max-width: 90vw;
      box-shadow: 0 4px 24px rgba(0,0,0,.2);
    }
    .ai-modal h3 { margin: 0 0 16px; font-size: 16px; }
    .ai-form-row { margin-bottom: 12px; }
    .ai-form-row label { display: block; font-size: 12px; color: #666; margin-bottom: 4px; }
    .ai-form-row input, .ai-form-row select {
      width: 100%;
      padding: 6px 10px;
      border: 1px solid #d9d9d9;
      border-radius: 4px;
      font-size: 13px;
      box-sizing: border-box;
    }
    .ai-modal-btns { display: flex; justify-content: flex-end; gap: 8px; margin-top: 16px; }
    .ai-btn-primary {
      background: #1890ff; color: #fff; border: none;
      border-radius: 4px; padding: 6px 16px; cursor: pointer; font-size: 13px;
    }
    .ai-btn-primary:hover { background: #096dd9; }
    .ai-btn-default {
      background: #fff; border: 1px solid #d9d9d9;
      border-radius: 4px; padding: 6px 16px; cursor: pointer; font-size: 13px;
    }
    .ai-status { font-size: 11px; color: #999; margin-top: 4px; }
  `;

  function injectStyles() {
    if (document.getElementById('ai-panel-styles')) return;
    var style = document.createElement('style');
    style.id = 'ai-panel-styles';
    style.textContent = PANEL_CSS;
    document.head.appendChild(style);
  }

  function createPanel() {
    if (document.getElementById('ai-panel-btn')) return;
    injectStyles();

    // Floating button
    var btn = document.createElement('button');
    btn.id = 'ai-panel-btn';
    btn.textContent = '✦ AI';
    btn.onclick = toggleDrawer;
    document.body.appendChild(btn);

    // Drawer
    var drawer = document.createElement('div');
    drawer.id = 'ai-panel-drawer';
    drawer.innerHTML = `
      <div class="ai-drawer-header">
        <span>AI 助手</span>
        <button class="ai-drawer-close" title="关闭">×</button>
      </div>
      <div class="ai-drawer-body">
        <div class="ai-section">
          <button class="ai-action-btn" id="ai-btn-summarize">📝 生成摘要</button>
          <div id="ai-result-summarize" style="display:none">
            <div class="ai-result-label">摘要</div>
            <div class="ai-result-box" id="ai-text-summarize"></div>
            <button class="ai-copy-btn" data-target="ai-text-summarize">复制</button>
          </div>
        </div>
        <hr class="ai-divider">
        <div class="ai-section">
          <button class="ai-action-btn" id="ai-btn-polish">✨ 润色内容</button>
          <div id="ai-result-polish" style="display:none">
            <div class="ai-result-label">润色结果</div>
            <div class="ai-result-box" id="ai-text-polish"></div>
            <button class="ai-copy-btn" data-target="ai-text-polish">复制</button>
          </div>
        </div>
        <hr class="ai-divider">
        <div class="ai-section">
          <button class="ai-action-btn" id="ai-btn-tags">🏷️ 推荐标签</button>
          <div id="ai-result-tags" style="display:none">
            <div class="ai-result-label">推荐标签</div>
            <div class="ai-result-box" id="ai-text-tags"></div>
            <button class="ai-copy-btn" data-target="ai-text-tags">复制</button>
            <button class="ai-copy-btn" id="ai-apply-tags" style="margin-left:6px">添加到标签框</button>
          </div>
        </div>
        <hr class="ai-divider">
        <a class="ai-settings-link" id="ai-settings-link">⚙ AI 配置</a>
      </div>
    `;
    document.body.appendChild(drawer);

    drawer.querySelector('.ai-drawer-close').onclick = toggleDrawer;
    document.getElementById('ai-btn-summarize').onclick = runSummarize;
    document.getElementById('ai-btn-polish').onclick = runPolish;
    document.getElementById('ai-btn-tags').onclick = runTags;
    document.getElementById('ai-settings-link').onclick = openSettings;
    document.querySelectorAll('.ai-copy-btn[data-target]').forEach(function (b) {
      b.onclick = function () {
        var text = document.getElementById(b.dataset.target).textContent;
        navigator.clipboard.writeText(text).catch(function () {});
      };
    });
    document.getElementById('ai-apply-tags').onclick = function () {
      var text = document.getElementById('ai-text-tags').textContent;
      appendTags(text);
    };
  }

  function removePanel() {
    var b = document.getElementById('ai-panel-btn');
    var d = document.getElementById('ai-panel-drawer');
    if (b) b.remove();
    if (d) d.remove();
  }

  function toggleDrawer() {
    var d = document.getElementById('ai-panel-drawer');
    if (d) d.classList.toggle('open');
  }

  // ── Actions ────────────────────────────────────────────────────────────────

  function runSummarize() {
    var content = getEditorContent();
    if (!content) { alert('编辑器内容为空'); return; }
    var btn = document.getElementById('ai-btn-summarize');
    var resultBox = document.getElementById('ai-result-summarize');
    var textBox = document.getElementById('ai-text-summarize');
    btn.disabled = true;
    textBox.textContent = '';
    resultBox.style.display = 'block';
    streamRequest('stream/summarize', { content: content },
      function (chunk) { textBox.textContent += chunk; },
      function () { btn.disabled = false; },
      function (err) { textBox.textContent = '错误: ' + err; btn.disabled = false; }
    );
  }

  function runPolish() {
    var content = getEditorContent();
    if (!content) { alert('编辑器内容为空'); return; }
    var btn = document.getElementById('ai-btn-polish');
    var resultBox = document.getElementById('ai-result-polish');
    var textBox = document.getElementById('ai-text-polish');
    btn.disabled = true;
    textBox.textContent = '';
    resultBox.style.display = 'block';
    streamRequest('stream/polish', { content: content },
      function (chunk) { textBox.textContent += chunk; },
      function () { btn.disabled = false; },
      function (err) { textBox.textContent = '错误: ' + err; btn.disabled = false; }
    );
  }

  function runTags() {
    var content = getEditorContent();
    if (!content) { alert('编辑器内容为空'); return; }
    var btn = document.getElementById('ai-btn-tags');
    var resultBox = document.getElementById('ai-result-tags');
    var textBox = document.getElementById('ai-text-tags');
    btn.disabled = true;
    textBox.textContent = '';
    resultBox.style.display = 'block';
    var title = getPostTitle();
    streamRequest('stream/suggest-tags', { title: title, content: content },
      function (chunk) { textBox.textContent += chunk; },
      function () { btn.disabled = false; },
      function (err) { textBox.textContent = '错误: ' + err; btn.disabled = false; }
    );
  }

  // ── Settings modal ─────────────────────────────────────────────────────────

  function openSettings() {
    if (document.getElementById('ai-settings-modal')) return;
    var token = getToken();

    var overlay = document.createElement('div');
    overlay.className = 'ai-modal-overlay';
    overlay.id = 'ai-settings-modal';
    overlay.innerHTML = `
      <div class="ai-modal">
        <h3>⚙ AI 配置</h3>
        <div class="ai-form-row">
          <label>提供商</label>
          <select id="ais-provider">
            <option value="anthropic">Anthropic</option>
            <option value="openai">OpenAI</option>
            <option value="ollama">Ollama</option>
          </select>
        </div>
        <div class="ai-form-row">
          <label>API Key（留空保留原值）</label>
          <input id="ais-apikey" type="password" placeholder="sk-..." autocomplete="off">
        </div>
        <div class="ai-form-row">
          <label>模型</label>
          <input id="ais-model" type="text" placeholder="claude-sonnet-4-5">
        </div>
        <div class="ai-form-row">
          <label>Base URL（OpenAI 兼容接口可选）</label>
          <input id="ais-baseurl" type="text" placeholder="https://api.openai.com">
        </div>
        <div class="ai-status" id="ais-status"></div>
        <div class="ai-modal-btns">
          <button class="ai-btn-default" id="ais-cancel">取消</button>
          <button class="ai-btn-primary" id="ais-save">保存</button>
        </div>
      </div>
    `;
    document.body.appendChild(overlay);

    overlay.querySelector('#ais-cancel').onclick = function () { overlay.remove(); };
    overlay.onclick = function (e) { if (e.target === overlay) overlay.remove(); };

    // Load current config
    fetch('/api/admin/ai/config', {
      headers: { 'Admin-Authorization': token }
    }).then(function (r) { return r.json(); }).then(function (data) {
      if (data.data) {
        var d = data.data;
        document.getElementById('ais-provider').value = d.provider || 'anthropic';
        document.getElementById('ais-model').value = d.model || '';
        document.getElementById('ais-baseurl').value = d.base_url || '';
      }
    }).catch(function () {});

    overlay.querySelector('#ais-save').onclick = function () {
      var status = document.getElementById('ais-status');
      status.textContent = '保存中...';
      var body = {
        provider: document.getElementById('ais-provider').value,
        api_key: document.getElementById('ais-apikey').value,
        model: document.getElementById('ais-model').value,
        base_url: document.getElementById('ais-baseurl').value,
      };
      fetch('/api/admin/ai/config', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Admin-Authorization': token,
        },
        body: JSON.stringify(body),
      }).then(function (r) { return r.json(); }).then(function () {
        status.textContent = '✓ 已保存';
        setTimeout(function () { overlay.remove(); }, 800);
      }).catch(function (err) {
        status.textContent = '保存失败: ' + err.message;
      });
    };
  }

  // ── Route watcher ──────────────────────────────────────────────────────────

  function isEditorRoute() {
    var hash = location.hash;
    return EDITOR_ROUTES.some(function (r) { return hash.startsWith(r); });
  }

  function onRouteChange() {
    var newRoute = location.hash;
    if (newRoute === currentRoute) return;
    currentRoute = newRoute;

    if (isEditorRoute()) {
      // Wait for Vue to render the editor
      setTimeout(function () { createPanel(); }, 800);
    } else {
      removePanel();
    }
  }

  window.addEventListener('hashchange', onRouteChange);
  // Initial check
  onRouteChange();

  // Also handle programmatic navigation (Vue router pushState)
  var _pushState = history.pushState;
  history.pushState = function () {
    _pushState.apply(history, arguments);
    onRouteChange();
  };
})();
