// trace.js — S3 Trace rendering inside a Star

function showS3(constellation, star) {
  window._currentStar = star;
  window._currentConstellation = constellation;
}

function renderS3Content(constellation, star) {
  const section = document.getElementById('s3-trace');
  window._currentStar = star;

  let html = '';

  // Back link
  html += `<div class="s3-back" id="s3-back">← ${constellation.name}</div>`;

  // Star header
  html += '<div class="s3-star-header">';
  html += '<div class="s3-star-dot"></div>';
  html += `<div class="s3-star-topic">${star.topic}</div>`;
  html += `<div class="s3-star-meta">${star.date} · ${star.time} · 写了 ${star.duration}</div>`;
  html += '</div>';

  // Trace badge
  html += `<div class="s3-trace-badge">一次连续的 Trace · ${star.momentCount} 个 Moment</div>`;

  // Moment blocks
  const borderColors = [
    'rgba(255, 215, 0, 0.25)',
    'rgba(255, 200, 100, 0.3)',
    'rgba(255, 185, 85, 0.35)',
    'rgba(255, 170, 70, 0.4)',
    'rgba(255, 155, 60, 0.45)',
    'rgba(255, 140, 50, 0.5)',
  ];

  for (let i = 0; i < star.moments.length; i++) {
    const m = star.moments[i];
    const borderColor = borderColors[i] || borderColors[borderColors.length - 1];
    const label = i === 0 ? 'MOMENT · 此刻写下' : 'MOMENT · 接着写';

    html += '<div class="moment-block">';
    html += `<div class="moment-num">${i + 1}</div>`;

    html += `<div class="moment-content" style="border-left-color:${borderColor}">`;
    html += `<div class="label">${label}</div>`;
    html += `<div class="text">"${escapeHtml(m.content)}"</div>`;
    html += '</div>';

    html += '<div class="echo-card">';
    html += '<div class="echo-label">你之前也说过类似的</div>';
    html += `<div class="echo-text">"${escapeHtml(m.echo.content)}"</div>`;
    html += `<div class="echo-date">${m.echo.date}</div>`;
    if (m.echo.candidates > 1) {
      html += `<div class="echo-candidates">▸ 之前的你还说过 ${m.echo.candidates} 条</div>`;
    }
    html += '</div>';

    html += '<div class="insight-card-sm">';
    html += '<div class="insight-label-sm">✦ 我发现</div>';
    html += `<div class="insight-text-sm">${m.insight}</div>`;
    html += '</div>';

    html += '</div>';

    if (i < star.moments.length - 1) {
      html += '<div class="trace-connector">↓ 顺着再想想</div>';
    }
  }

  html += '<div class="trace-stash">✦ 收进星图</div>';

  html += '<div class="star-topic-summary">';
  html += '<div class="summary-label">这颗 Star 的主题</div>';
  html += `<div class="summary-topic">${star.topic}</div>`;
  html += `<div class="summary-note">AI 基于 ${star.momentCount} 次连续表达提炼</div>`;
  html += '</div>';

  // Clickable chat entry
  html += '<div class="chat-entry" id="s3-chat-entry">';
  html += '<div class="chat-title">💬 和那段时间的 ta 说说话</div>';
  html += `<div class="chat-note">对话基于这颗 Star · ${star.date}${star.time}的那 ${star.duration}</div>`;
  html += '</div>';

  // Skip to S4
  html += '<div class="s3-skip" id="s3-skip">直接看转化 →</div>';

  section.innerHTML = html;

  // Back handler
  document.getElementById('s3-back').addEventListener('click', () => {
    transitionTo('s3-trace', 's2-constellation');
  });

  // Chat entry → S5
  document.getElementById('s3-chat-entry').addEventListener('click', () => {
    transitionTo('s3-trace', 's5-chat');
    renderS5(star);
  });

  // Skip → S4
  document.getElementById('s3-skip').addEventListener('click', () => {
    transitionTo('s3-trace', 's4-cta');
  });

  // Animate moment blocks
  const blocks = section.querySelectorAll('.moment-block');
  blocks.forEach((block, i) => {
    block.style.opacity = '0';
    block.style.transform = 'translateY(16px)';
    block.style.transition = 'opacity 0.4s ease, transform 0.4s ease';
    setTimeout(() => {
      block.style.opacity = '1';
      block.style.transform = 'translateY(0)';
    }, 100 * i);
  });
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}

function renderS3() {
  document.getElementById('s3-trace').innerHTML = '';
}
