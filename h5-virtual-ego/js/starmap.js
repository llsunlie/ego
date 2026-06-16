// starmap.js — S1 Star map + S2 Constellation rendering

function renderS1() {
  const section = document.getElementById('s1-starmap');
  const { figure, constellations } = window.appData;

  let html = '';

  // All text content above star map
  html += `<div class="s1-title-top">${figure.title}</div>`;
  html += `<div class="s1-title">${figure.subtitle}</div>`;
  html += '<div class="s1-hint">点击星座名称，看里面的星星 →</div>';

  // Star map canvas with constellations
  html += '<div class="starmap-canvas" id="starmap-canvas">';

  for (const c of constellations) {
    const { id, status, color, position: pos } = c;

    // Draw constellation lines
    if (pos.stars.length >= 2) {
      html += `<svg class="c-line" viewBox="0 0 100 100" preserveAspectRatio="none">`;
      for (let i = 0; i < pos.stars.length - 1; i++) {
        const a = pos.stars[i];
        const b = pos.stars[i + 1];
        html += `<line class="${status}" x1="${a.x}" y1="${a.y}" x2="${b.x}" y2="${b.y}" stroke="${color === 'gold' ? '#ffd700' : color === 'purple' ? '#c9b0ff' : '#7ec8e3'}" />`;
      }
      html += '</svg>';
    }

    // Draw stars
    for (let i = 0; i < pos.stars.length; i++) {
      const star = pos.stars[i];
      const sizes = status === 'formed' ? [14, 11, 10] : [9, 8];
      const size = sizes[i] || 7;
      html += `<div class="c-star ${status} ${color}"
        style="left:${star.x}%;top:${star.y}%;width:${size}px;height:${size}px;animation-delay:${i * 0.4}s;"></div>`;
    }

    // Clickable name button — offset bottom-right by +3% x, +3% y
    const btnClass = `c-name-btn ${color} ${status}`;
    html += `<div class="${btnClass}" style="left:${pos.labelX + 4}%;top:${pos.labelY + 3}%;" data-constellation="${id}">${c.name}</div>`;
  }

  html += '</div>';

  section.innerHTML = html;

  // Attach click handlers to constellation name buttons (only formed)
  const btns = section.querySelectorAll('.c-name-btn.formed');
  btns.forEach(btn => {
    btn.addEventListener('click', () => {
      const cid = btn.dataset.constellation;
      showS2(cid);
    });
  });
}

function showS2(constellationId) {
  const c = window.appData.constellations.find(c => c.id === constellationId);
  if (!c || c.status !== 'formed') return;

  // Fade out S1, show S2
  transitionTo('s1-starmap', 's2-constellation');
  renderS2Content(c);
}

function renderS2Content(c) {
  const section = document.getElementById('s2-constellation');

  let html = '';

  // Back link
  html += '<div class="s2-back" id="s2-back">← 星图全景</div>';

  // Header
  html += '<div class="s2-header">';
  html += `<div class="s2-header-name">✦ ${c.name}</div>`;
  html += `<div class="s2-header-meta">${c.stars.length} 颗星</div>`;
  html += '</div>';

  // Constellation insight
  html += '<div class="insight-card">';
  html += '<div class="insight-label">✦ 我发现</div>';
  html += `<div class="insight-text">${c.insight}</div>`;
  html += '</div>';

  // Stars label
  html += '<div class="s2-stars-label">星座里的星星</div>';

  // Star cards
  for (const star of c.stars) {
    html += `<div class="star-card" data-star-id="${star.id}">`;
    html += '<div>';
    html += `<div class="star-card-topic">${star.topic}</div>`;
    html += `<div class="star-card-meta">${star.date} · ${star.time} · ${star.momentCount} 次连续表达</div>`;
    html += '</div>';
    html += '<div class="star-card-arrow">▸</div>';
    html += '</div>';
  }

  section.innerHTML = html;

  // Attach back handler
  document.getElementById('s2-back').addEventListener('click', () => {
    transitionTo('s2-constellation', 's1-starmap');
  });

  // Attach star click handlers
  section.querySelectorAll('.star-card').forEach(card => {
    card.addEventListener('click', () => {
      const starId = card.dataset.starId;
      for (const con of window.appData.constellations) {
        const found = con.stars.find(s => s.id === starId);
        if (found) {
          transitionTo('s2-constellation', 's3-trace');
          renderS3Content(con, found);
          return;
        }
      }
    });
  });
}

function renderS2() {
  document.getElementById('s2-constellation').innerHTML = '';
}
