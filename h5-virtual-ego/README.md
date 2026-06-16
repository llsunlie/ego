# h5-virtual-ego

H5 虚拟热点人士 ego 旁观体验页。数据驱动，可复用。

## Quick Start

```sh
cd h5-virtual-ego
python3 -m http.server 8080
# Open http://localhost:8080
```

## How to Use

1. Create a new JSON data file in `data/` following the schema below
2. Update `js/main.js` to load your data file: `fetch('data/your-figure.json')`
3. Deploy the `h5-virtual-ego/` directory to any static hosting

## Page Structure

```
S1 · 星图全景        →  3 个 Constellation（2 已成型 + 1 正在成型）
    ↓ 点击星座
S2 · 星座展开        →  1 个 Constellation + 3 颗 Star 列表
    ↓ 点击星星
S3 · 星内 Trace      →  1 颗 Star = 4-6 个 Moment-Echo-Insight 元组
    ↓ 继续下滑
S4 · 转化 CTA        →  镜子式引导
```

## Data Schema

```typescript
interface H5Data {
  figure: {
    name: string;        // Figure name, e.g. "余华"
    title: string;       // "如果 XX 有 ego"
    subtitle: string;    // "他的星图长这样"
  };
  constellations: Constellation[];  // Exactly 3
  cta: CTA;
}

interface Constellation {
  id: string;            // Unique ID
  name: string;          // e.g. "关于等待"
  status: "formed" | "forming";
  color: "gold" | "purple" | "blue";
  position: {
    stars: { x: number; y: number }[];  // Percentage positions (0-100)
    labelX: number;
    labelY: number;
  };
  insight: string;       // Constellation-level AI insight
  stars: Star[];         // 3 per formed constellation
}

interface Star {
  id: string;
  topic: string;         // AI-generated topic
  date: string;          // e.g. "2023年3月"
  time: string;          // e.g. "凌晨 2:14"
  duration: string;      // e.g. "47 分钟"
  momentCount: number;   // 4-6
  moments: Moment[];
}

interface Moment {
  content: string;       // The original words
  echo: {
    content: string;     // Matched past similar words
    date: string;
    candidates: number;  // Additional candidates count
  };
  insight: string;       // AI observation
}

interface CTA {
  mirrorLines: string[];  // Mirror text lines ("" for blank line)
  bodyLines: string[];    // Body explanation
  primaryButton: string;  // e.g. "建立我的 ego →"
  secondaryText: string;  // e.g. "嗯，先看看别人的"
}

interface ApiConfig {
  endpoint: string;       // Proxy endpoint URL (POST {system, messages} → {content})
  apiKey: string;         // API key for proxy endpoint
  anthropicKey: string;   // Direct Anthropic API key (bypasses proxy)
  model: string;          // Model name, e.g. "claude-sonnet-4-6"
}
```

## S5 Chat LLM Configuration

Three modes (checked in order):

1. **Proxy mode** (recommended): Start the included proxy:
   ```sh
   cd proxy
   export ANTHROPIC_AUTH_TOKEN=sk-...
   export ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic  # optional
   export ANTHROPIC_MODEL=deepseek-v4-pro                        # optional
   python3 server.py  # listens on :8090
   ```
   Uses the same env vars and Anthropic-compatible API format as the ego backend. H5 POSTs `{system, messages, model}` to `/chat`, proxy returns `{content}`.

2. **Direct Anthropic-compatible** (dev only): Set `api.anthropicKey`. Keys exposed to browser.

3. **Fallback**: Leave all API fields empty. Uses simple keyword matching for demo.

## Design

- **Baseline:** 375px mobile viewport, max-width 480px centered
- **Visual:** Deep space (#07071a) + gold (#ffd700) + purple (#c9b0ff) + blue (#7ec8e3)
- **Typography:** System font stack with PingFang SC priority for Chinese
- **Animation:** Canvas star twinkling, SVG line breathing, pulse glow on constellation buttons, sequential moment block fade-in
- **Transitions:** Fade-out/fade-in between sections (no scrolling)
