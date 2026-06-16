// main.js — Entry point, data loading, section orchestration, S5 chat

window.appData = null;

// ============================================================
// Section transition helpers
// ============================================================

function transitionTo(fromId, toId) {
  const from = document.getElementById(fromId);
  const to = document.getElementById(toId);

  // Fade out current
  from.classList.add('fade-out');

  setTimeout(() => {
    from.classList.add('hidden');
    from.classList.remove('fade-out');

    // Show next
    to.classList.remove('hidden');
    to.style.opacity = '0';
    to.style.transform = 'translateY(10px)';

    requestAnimationFrame(() => {
      to.style.opacity = '1';
      to.style.transform = 'translateY(0)';
    });

    to.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }, 300);
}

// ============================================================
// Init
// ============================================================

async function init() {
  try {
    const res = await fetch('data/example.json');
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    window.appData = await res.json();
  } catch (e) {
    console.error('Failed to load data:', e);
    showError('数据加载失败，请稍后重试。');
    return;
  }

  renderAll();
}

function showError(msg) {
  document.getElementById('content').innerHTML =
    `<div style="padding:80px 24px;text-align:center;color:var(--text-dim)">${msg}</div>`;
}

function renderAll() {
  renderS1();
  renderS4();
}

function renderS4() {
  const section = document.getElementById('s4-cta');
  const { figure, cta } = window.appData;

  let html = '';

  html += '<div class="s4-top">';
  html += `你看完了${figure.name}的星图。<br>`;
  html += '他的 Moment，他的 Echo，他的 Constellation。';
  html += '</div>';

  html += '<div class="s4-mirror">';
  for (const line of cta.mirrorLines) {
    if (line === '') {
      html += '<br>';
    } else {
      html += `${line}<br>`;
    }
  }
  html += '</div>';

  html += '<div class="s4-body">';
  for (const line of cta.bodyLines) {
    html += `${line}<br>`;
  }
  html += '</div>';

  html += `<a class="s4-btn" id="s4-primary" href="https://myego.online/#/now">${cta.primaryButton}</a>`;
  html += `<div class="s4-dismiss" id="s4-dismiss">${cta.secondaryText}</div>`;

  section.innerHTML = html;

  document.getElementById('s4-dismiss').addEventListener('click', () => {
    transitionTo('s4-cta', 's1-starmap');
  });
}

// ============================================================
// S5: Chat with past self
// ============================================================

let chatMessages = [];
let chatRound = 0;
const MAX_ROUNDS = 3;

function renderS5(star) {
  const section = document.getElementById('s5-chat');
  chatMessages = [];
  chatRound = 0;

  // Build system prompt from template in data json
  const template = window.appData.api?.promptTemplate || '';
  const momentsText = star.moments.map((m, i) => `[${i + 1}] ${m.content}`).join('\n');
  const systemPrompt = template.replace('{moments}', momentsText);
  window._chatContext = { star, systemPrompt };

  let html = '';

  html += '<div class="s5-header">';
  html += '<div class="s5-star-info">';
  html += `<span class="s5-topic">${star.topic}</span>`;
  html += `${star.date} · ${star.time}`;
  html += '</div>';
  html += '<div class="s5-exit" id="s5-exit">退出对话</div>';
  html += '</div>';

  html += '<div class="s5-messages" id="s5-messages">';

  const firstMoment = star.moments[0];
  const snippet = firstMoment.content.length > 30
    ? firstMoment.content.substring(0, 30) + '…'
    : firstMoment.content;
  const opening = `我那时候写过这样一句——「${snippet}」你想跟那时的我聊什么？`;
  chatMessages.push({ role: 'assistant', content: opening });

  html += '<div class="s5-msg past-self">';
  html += `<div class="msg-bubble">${opening}</div>`;
  html += `<div class="msg-ref">参考了 ${star.date} 前后</div>`;
  html += '</div>';
  html += '</div>';

  html += '<div class="s5-rounds" id="s5-rounds">可以说 3 次话</div>';

  html += '<div class="s5-input-row">';
  html += '<input class="s5-input" id="s5-input" type="text" placeholder="你想对那时候的 ta 说什么？" />';
  html += '<button class="s5-send" id="s5-send">发送</button>';
  html += '</div>';

  section.innerHTML = html;

  document.getElementById('s5-exit').addEventListener('click', finishChat);
  document.getElementById('s5-send').addEventListener('click', sendMessage);
  document.getElementById('s5-input').addEventListener('keydown', (e) => {
    if (e.key === 'Enter') sendMessage();
  });

  setTimeout(() => document.getElementById('s5-input').focus(), 400);
}

async function sendMessage() {
  const input = document.getElementById('s5-input');
  const text = input.value.trim();
  if (!text) return;
  if (chatRound >= MAX_ROUNDS) return;

  // Disable input during response
  input.disabled = true;
  document.getElementById('s5-send').disabled = true;

  // Add user message
  addMessage('user', text);
  input.value = '';
  chatRound++;

  // Update round counter
  document.getElementById('s5-rounds').textContent =
    `还剩 ${MAX_ROUNDS - chatRound} 次`;

  // Get LLM response
  const messagesDiv = document.getElementById('s5-messages');
  const thinkingId = addThinking(messagesDiv);

  try {
    const reply = await callLLM();
    removeThinking(thinkingId);
    addMessage('past-self', reply, window._chatContext.star.date);
  } catch (e) {
    removeThinking(thinkingId);
    addMessage('past-self', '（我那时候好像不太想说这个……换个话题？）', '');
  }

  // Re-enable input or show next button
  input.disabled = false;
  document.getElementById('s5-send').disabled = false;
  input.focus();

  if (chatRound >= MAX_ROUNDS) {
    // Show continue button, hide input
    document.getElementById('s5-input-row').style.display = 'none';
    document.getElementById('s5-rounds').textContent = '对话结束';

    const nextBtn = document.createElement('div');
    nextBtn.className = 's5-next-btn';
    nextBtn.id = 's5-next';
    nextBtn.textContent = '继续 →';
    nextBtn.addEventListener('click', finishChat);
    document.getElementById('s5-chat').appendChild(nextBtn);
  }
}

function addMessage(role, text, refDate) {
  // Track in conversation history for LLM context
  chatMessages.push({ role, content: text });

  const container = document.getElementById('s5-messages');
  const div = document.createElement('div');
  div.className = `s5-msg ${role}`;

  let html = `<div class="msg-bubble">${text}</div>`;
  if (refDate) {
    html += `<div class="msg-ref">参考了 ${refDate} 前后</div>`;
  }

  div.innerHTML = html;
  container.appendChild(div);
  container.scrollTop = container.scrollHeight;
}

function addThinking(container) {
  const div = document.createElement('div');
  div.className = 's5-msg past-self';
  div.id = 's5-thinking';
  div.innerHTML = '<div class="msg-bubble" style="opacity:0.4">…</div>';
  container.appendChild(div);
  return 's5-thinking';
}

function removeThinking(id) {
  const el = document.getElementById(id);
  if (el) el.remove();
}

async function callLLM() {
  const ctx = window._chatContext;
  const api = window.appData.api || {};

  const conversationMessages = chatMessages.map(msg => ({
    role: msg.role === 'user' ? 'user' : 'assistant',
    content: msg.content,
  }));

  if (!api.endpoint) {
    return '（未配置 API 端点，请启动 proxy 服务）';
  }

  const res = await fetch(api.endpoint, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      system: ctx.systemPrompt,
      messages: conversationMessages,
    }),
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || `HTTP ${res.status}`);
  }

  const data = await res.json();
  return data.content || '（对方沉默了……）';
}

function finishChat() {
  transitionTo('s5-chat', 's4-cta');
  chatMessages = [];
  chatRound = 0;
}

// ============================================================
// Boot
// ============================================================

document.addEventListener('DOMContentLoaded', init);
