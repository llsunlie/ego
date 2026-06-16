// background.js — Canvas-based deep space star field

const STAR_COUNT = 120;
const TWINKLE_RATE = 0.3;

let canvas, ctx, stars, width, height, dpr;

function initStarfield() {
  canvas = document.getElementById('starfield');
  ctx = canvas.getContext('2d');
  resize();
  generateStars();
  window.addEventListener('resize', resize);
  requestAnimationFrame(draw);
}

function resize() {
  dpr = window.devicePixelRatio || 1;
  const rect = canvas.getBoundingClientRect();
  width = rect.width;
  height = rect.height;
  canvas.width = width * dpr;
  canvas.height = height * dpr;
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
}

function generateStars() {
  stars = [];
  for (let i = 0; i < STAR_COUNT; i++) {
    stars.push({
      x: Math.random() * (width || 375),
      y: Math.random() * (height || 812),
      r: Math.random() * 1.5 + 0.3,
      baseAlpha: Math.random() * 0.3 + 0.05,
      twinkle: Math.random() < TWINKLE_RATE,
      twinkleSpeed: Math.random() * 0.02 + 0.005,
      twinkleOffset: Math.random() * Math.PI * 2,
    });
  }
}

function draw() {
  ctx.clearRect(0, 0, width, height);

  const now = Date.now() / 1000;

  for (const star of stars) {
    let alpha = star.baseAlpha;
    if (star.twinkle) {
      alpha = star.baseAlpha + Math.sin(now * star.twinkleSpeed * 60 + star.twinkleOffset) * 0.15;
      alpha = Math.max(0.02, Math.min(0.4, alpha));
    }

    ctx.beginPath();
    ctx.arc(star.x, star.y, star.r, 0, Math.PI * 2);
    ctx.fillStyle = `rgba(255, 255, 255, ${alpha})`;

    if (star.r > 1.2) {
      ctx.shadowColor = `rgba(255, 255, 255, ${alpha * 0.5})`;
      ctx.shadowBlur = star.r * 3;
    } else {
      ctx.shadowColor = 'transparent';
      ctx.shadowBlur = 0;
    }

    ctx.fill();
  }

  ctx.shadowColor = 'transparent';
  ctx.shadowBlur = 0;

  requestAnimationFrame(draw);
}

document.addEventListener('DOMContentLoaded', initStarfield);
