# H5 Virtual Ego Template · Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a reusable, data-driven H5 page that displays a virtual person's ego (star map → constellation → trace → CTA) as a promotional tool for the ego app.

**Architecture:** Standalone vanilla HTML/CSS/JS project at `h5-virtual-ego/`. A single `index.html` loads a `data.json` file and renders four scrollable sections. Canvas API for background star field; DOM manipulation for content rendering; CSS transitions for expand/collapse; Intersection Observer for scroll-triggered animations.

**Tech Stack:** Vanilla HTML5, CSS3 (custom properties, transitions, flexbox), ES2020+ JavaScript, Canvas API, no frameworks.

**Design baseline:** 375px mobile viewport, max-width 480px centered on larger screens.

---

## File Structure

```
h5-virtual-ego/
├── index.html              # Single page, loads all resources
├── css/
│   └── main.css            # All styles, organized by section
├── js/
│   ├── main.js             # Entry point, data loading, section orchestration
│   ├── background.js       # Canvas star field animation
│   ├── starmap.js          # S1 constellation + S2 star list rendering
│   └── trace.js            # S3 trace rendering
├── data/
│   └── example.json        # Example data for 余华
└── README.md               # Usage + data schema docs
```

---

### Task 1: Project scaffolding

**Files:**
- Create: `h5-virtual-ego/index.html`
- Create: `h5-virtual-ego/css/main.css`
- Create: `h5-virtual-ego/js/main.js`
- Create: `h5-virtual-ego/README.md`

- [ ] **Step 1: Create index.html shell**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
  <title>ego · 旁观体验</title>
  <link rel="stylesheet" href="css/main.css">
</head>
<body>
  <div id="app">
    <canvas id="starfield"></canvas>
    <main id="content">
      <section id="s1-starmap"></section>
      <section id="s2-constellation" class="hidden"></section>
      <section id="s3-trace" class="hidden"></section>
      <section id="s4-cta" class="hidden"></section>
    </main>
  </div>
  <script src="js/background.js"></script>
  <script src="js/starmap.js"></script>
  <script src="js/trace.js"></script>
  <script src="js/main.js"></script>
</body>
</html>
```

- [ ] **Step 2: Create main.css with CSS variables and base reset**

```css
/* === CSS Variables === */
:root {
  --bg-deep: #07071a;
  --bg-panel: #0a0a1a;
  --gold: #ffd700;
  --gold-dim: rgba(255, 215, 0, 0.5);
  --gold-soft: rgba(255, 215, 0, 0.12);
  --purple: #c9b0ff;
  --purple-dim: rgba(201, 176, 255, 0.5);
  --purple-soft: rgba(201, 176, 255, 0.12);
  --blue: #7ec8e3;
  --blue-dim: rgba(126, 200, 227, 0.5);
  --text-primary: rgba(255, 255, 255, 0.75);
  --text-secondary: rgba(255, 255, 255, 0.45);
  --text-dim: rgba(255, 255, 255, 0.2);
  --text-faint: rgba(255, 255, 255, 0.12);
  --border-gold: rgba(255, 215, 0, 0.18);
  --border-purple: rgba(201, 176, 255, 0.2);
  --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
}

/* === Reset === */
*, *::before, *::after {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html {
  font-size: 16px;
  -webkit-font-smoothing: antialiased;
}

body {
  font-family: var(--font-sans);
  background: var(--bg-deep);
  color: var(--text-primary);
  overflow-x: hidden;
  -webkit-tap-highlight-color: transparent;
}

/* === App Layout === */
#app {
  position: relative;
  width: 100%;
  max-width: 480px;
  margin: 0 auto;
  min-height: 100vh;
}

#starfield {
  position: fixed;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: 480px;
  height: 100vh;
  z-index: 0;
  pointer-events: none;
}

#content {
  position: relative;
  z-index: 1;
}

section {
  min-height: 100vh;
  padding: 48px 24px;
  position: relative;
}

.hidden {
  display: none;
}

/* === Scroll transition === */
section {
  opacity: 0;
  transform: translateY(20px);
  transition: opacity 0.5s ease, transform 0.5s ease;
}

section.visible {
  opacity: 1;
  transform: translateY(0);
}
```

- [ ] **Step 3: Create main.js shell with initialization**

```javascript
// main.js — Entry point, data loading, section orchestration

let appData = null;

async function init() {
  try {
    const res = await fetch('data/example.json');
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    appData = await res.json();
  } catch (e) {
    console.error('Failed to load data:', e);
    showError('数据加载失败，请稍后重试。');
    return;
  }

  renderAll();
  initScrollObserver();
}

function showError(msg) {
  document.getElementById('content').innerHTML =
    `<div style="padding:80px 24px;text-align:center;color:var(--text-dim)">${msg}</div>`;
}

function initScrollObserver() {
  const sections = document.querySelectorAll('section');
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible');
      }
    });
  }, { threshold: 0.1 });

  sections.forEach(s => observer.observe(s));
}

function renderAll() {
  renderS1();
  renderS2();
  renderS3();
  renderS4();
}

document.addEventListener('DOMContentLoaded', init);
```

- [ ] **Step 4: Create README.md**

```markdown
# h5-virtual-ego

H5 虚拟热点人士 ego 旁观体验页。数据驱动，可复用。

## Quick Start

```sh
# Any static file server
python3 -m http.server 8080
# Open http://localhost:8080
```

## Data Schema

See `data/example.json` for the full example. The template expects this structure:

- `figure` — name and title strings
- `constellations[]` — 3 items, each with `name`, `status` ("formed"|"forming"), `color`, `position`, `insight`, `stars[]`
- `cta` — mirror text and button labels

See schema details in Task 2.
```

- [ ] **Step 5: Commit**

```bash
git add h5-virtual-ego/
git commit -m "feat(h5): scaffold virtual ego H5 template project"
```

---

### Task 2: Data schema and example data

**Files:**
- Create: `h5-virtual-ego/data/example.json`

- [ ] **Step 1: Create example.json with 余华 data**

```json
{
  "figure": {
    "name": "余华",
    "title": "如果 余华 有 ego",
    "subtitle": "他的星图长这样"
  },
  "constellations": [
    {
      "id": "c1",
      "name": "关于等待",
      "status": "formed",
      "color": "gold",
      "position": {
        "stars": [
          { "x": 20, "y": 22 },
          { "x": 34, "y": 16 },
          { "x": 48, "y": 26 }
        ],
        "labelX": 22,
        "labelY": 8
      },
      "insight": "三次都是"他不说话，我也不说"。第一次你用的是"等"，第二次是"内耗"，第三次是"别扭"——同一件事，你叫它的名字在变。",
      "stars": [
        {
          "id": "s1",
          "topic": "等一个人看见我",
          "date": "2023年3月",
          "time": "凌晨 2:14",
          "duration": "47 分钟",
          "momentCount": 5,
          "moments": [
            {
              "content": "其实我没有真的生气，我只是想让他先开口。",
              "echo": {
                "content": "每次吵完架我都在等他先发消息。不是放不下脸——是怕发了也没有用。",
                "date": "2022年11月",
                "candidates": 2
              },
              "insight": "四个月前说"怕发了也没有用"，现在说"想让他先开口"。上次在说怕什么——这次在说自己想要什么。"
            },
            {
              "content": "我一直在想，为什么先开口的那个人不能是我。后来我意识到——我怕的不是开口，是开口之后他什么都没变。",
              "echo": {
                "content": "一段关系里最累的不是吵架——是你在心里翻了几百页，对方连封面都没打开。",
                "date": "2021年6月",
                "candidates": 1
              },
              "insight": "刚才还在说"让他先开口"，现在你把它推得更远了——你发现"开口"只是一个动作，真正让你怕的，是开口之后对方纹丝不动。你在往自己里面走。"
            },
            {
              "content": "我其实不是想等他先开口。我是想等他把我看清楚。不是等一句话，是等一个人真的看见我。",
              "echo": {
                "content": "有时候我觉得自己站在人群边上，不是在等人过来——是在等有人发现我没过去。",
                "date": "2020年9月",
                "candidates": 2
              },
              "insight": "你从"让他先开口"，到"怕他什么都没变"，到这里——你说你要的不是一句话，是"被看见"。三次转弯，最后停下来的地方，和起点已经是两回事了。"
            },
            {
              "content": "然后我想到——我一直在等人看见我。但我自己看清过自己吗？",
              "echo": {
                "content": "写作的时候最怕的不是写不出来——是写出来才发现自己不认识那个人。",
                "date": "2019年4月",
                "candidates": 1
              },
              "insight": "这是一个转折——你从"他看不看见我"转到了"我看不看见自己"。前面的三次都是向外看的，这里你第一次向内看了。"
            },
            {
              "content": "也许等一个人看见我，和我看见我自己，是同一件事。也许我就是想等他看见我的时候，我能说——我已经开始看自己了。",
              "echo": {
                "content": "我写小说从来不是为了讲好一个故事——是为了搞清楚自己到底在想什么。写到一百页的时候才发现，哦，原来我是这样的人。",
                "date": "2018年2月",
                "candidates": 2
              },
              "insight": "你从"等一个人看见我"走到了"我开始看自己了"。47分钟、5次表达——你把"等"这件事从外到内翻了个面。外面的人可能还没来，但你先到了自己面前。这就是写作的用处。"
            }
          ]
        },
        {
          "id": "s2",
          "topic": "翻了几百页，封面都没打开",
          "date": "2023年6月",
          "time": "下午 4:30",
          "duration": "35 分钟",
          "momentCount": 4,
          "moments": [
            {
              "content": "我发现自己总是在等。等一个回复、等一个决定、等别人先做什么。好像我的生活是一本翻不完的书——翻了几百页，封面都没打开。",
              "echo": {
                "content": "人生就是一直在等。等天亮，等天黑，等一个不会来的人。",
                "date": "2021年3月",
                "candidates": 2
              },
              "insight": "三个月前你说"等一个人看见我"，这一次你把"等"的范围拉开了——不再只是等一个人，而是等所有事情开始。"
            },
            {
              "content": "然后我问自己：我在等什么？等一个对的时机？但时机从来不对——对的是你决定不等的那一刻。",
              "echo": {
                "content": "我总觉得明天会不一样。但每一个明天都是今天的复制。",
                "date": "2020年12月",
                "candidates": 1
              },
              "insight": "你在推自己——不是别人推你，是你自己在推自己。开始意识到"等"这个动作本身就是问题。"
            },
            {
              "content": "我觉得我等的东西其实不是一件事，是一个人——一个先迈出一步的人。但我不知道那个人是不是我。",
              "echo": {
                "content": "有时候不是没有勇气，是没有理由。你为什么要做那个先开口的人？",
                "date": "2019年8月",
                "candidates": 2
              },
              "insight": "你把问题具体化了——等的不再是模糊的"时机"，而是"有人先迈步"。但你也承认，那个人可能就是自己。"
            },
            {
              "content": "也许我等的从来不是一个别人。我等的，是自己愿意停下来，承认自己也在害怕。",
              "echo": {
                "content": "承认自己不行这件事，比逞强难一百倍。",
                "date": "2019年2月",
                "candidates": 1
              },
              "insight": "你从等别人走到了等自己——愿意面对害怕的自己。三次说"等"，最后说的其实是"承认"。"
            }
          ]
        },
        {
          "id": "s3",
          "topic": "别扭",
          "date": "2023年12月",
          "time": "晚上 11:08",
          "duration": "22 分钟",
          "momentCount": 4,
          "moments": [
            {
              "content": "今天又开始了那种感觉——不是生气，不是难过，就是别扭。像穿了一件反的衣服，没人告诉你你也不会脱，但它就是不对。",
              "echo": {
                "content": "有时候胃里不舒服，不是吃坏了东西，是心里有事没说出来。",
                "date": "2022年8月",
                "candidates": 1
              },
              "insight": "你用了一个新词——"别扭"。它比"等"更身体化，比"内耗"更安静。你在给自己的感受找一个越来越准确的名字。"
            },
            {
              "content": "以前我会分析为什么别扭，找原因，找解决办法。现在我不会了。别扭就是别扭，我得坐着和它待一会儿。",
              "echo": {
                "content": "我学会了一件事：不是每种情绪都需要解决。有些情绪只是需要被看见。",
                "date": "2020年5月",
                "candidates": 2
              },
              "insight": "今年的你说"坐着和它待一会儿"。三年前的你说"需要被看见"。同一个方向——对待自己的方式在变温和。"
            },
            {
              "content": "别扭可能是在告诉我——有什么东西不对了，但我还不知道是什么。它不是一个问题，它是一个信号。",
              "echo": {
                "content": "恐惧不一定是坏事。它只是告诉你，这里有个你还不知道怎么处理的东西。",
                "date": "2019年11月",
                "candidates": 1
              },
              "insight": "你把"别扭"重新定义为"信号"——这是你今年的变化。不再和情绪对抗，而是听它想说什么。"
            },
            {
              "content": "也许每一个别扭，都是在让我停下来看看自己。不是修理自己——就是看看。",
              "echo": {
                "content": "写小说的人最怕的不是读者不喜欢——是自己不敢面对自己写出来的东西。",
                "date": "2018年6月",
                "candidates": 1
              },
              "insight": "你看，你从"等"变成了"内耗"，又从"内耗"变成了"别扭"。每一次换词，你都离自己近了一步。不是离答案近，是离自己近。"
            }
          ]
        }
      ]
    },
    {
      "id": "c2",
      "name": "关于写作",
      "status": "formed",
      "color": "purple",
      "position": {
        "stars": [
          { "x": 55, "y": 48 },
          { "x": 42, "y": 54 },
          { "x": 62, "y": 56 }
        ],
        "labelX": 50,
        "labelY": 42
      },
      "insight": "写作对你来说不是手艺，不是职业，是一个办法——搞清楚自己到底在想什么的办法。从"废纸"到"怕"到"就这样吧"，你写的不是小说，是你自己。",
      "stars": [
        {
          "id": "s4",
          "topic": "下一页会不会是废纸",
          "date": "2022年7月",
          "time": "凌晨 3:02",
          "duration": "52 分钟",
          "momentCount": 5,
          "moments": [
            {
              "content": "写作就是这样，你永远不知道下一页会不会是废纸。但你只能往下写。",
              "echo": {
                "content": "活着本身就是一种写作。你不知道下一章是什么，但你翻不过去。",
                "date": "2018年10月",
                "candidates": 2
              },
              "insight": "你说"只能往下写"——不是"想往下写"，是"只能"。写作对你来说不是选择，是必须。"
            },
            {
              "content": "有时候打开文档，光标在闪，我在等。等什么呢？等一个句子自己跳出来。但它从来不会。你得伸手去抓。",
              "echo": {
                "content": "灵感不是等来的，是写来的。你写着写着，它就来了。",
                "date": "2019年3月",
                "candidates": 1
              },
              "insight": "你把"写作"和"等待"放在了一起——写作也是在等。但你说"得伸手去抓"。你和等的关系在变。"
            },
            {
              "content": "我觉得每一个作家都有一个抽屉，里面装满了不敢拿出来的东西。我的抽屉是满的。",
              "echo": {
                "content": "一个人写出来的东西，只是冰山浮在水面上的那一小块。",
                "date": "2017年7月",
                "candidates": 2
              },
              "insight": "你不只说写作难——你说"抽屉是满的"。你写出来了，但也藏了更多没写的。"
            },
            {
              "content": "今天写了一段，删了。又写了一段，又删了。第三段留着——不是因为写得好，是因为我决定不再删了。",
              "echo": {
                "content": "修改是写作的一部分。但有时候不是修改的问题——是你不敢说真话。",
                "date": "2018年5月",
                "candidates": 1
              },
              "insight": "你说"决定不再删了"——这是一个态度转变。不是写得好了，是你放过自己了。"
            },
            {
              "content": "后来我发现，废纸和好作品之间的差别不是才华——是你有没有勇气在废纸上继续写。",
              "echo": {
                "content": "活着就是这样。不是你活得好不好——是你有没有继续活着。",
                "date": "2016年3月",
                "candidates": 1
              },
              "insight": "你从"废纸"走了一圈，最后说废纸本身也是路的一部分。写作和活着，在你这里是同一件事。"
            }
          ]
        },
        {
          "id": "s5",
          "topic": "写到一百页才认识自己",
          "date": "2023年1月",
          "time": "下午 5:15",
          "duration": "38 分钟",
          "momentCount": 4,
          "moments": [
            {
              "content": "写了一百页，才发现前面八十页都是在绕路。不是内容错了——是我不敢直接写我想写的。",
              "echo": {
                "content": "修改到第十稿的时候，发现最想说的话在第一稿就有了，只是没敢留着。",
                "date": "2019年6月",
                "candidates": 1
              },
              "insight": "绕路不是迷路。你说"不敢直接写"——问题不在写作，在勇气。八十页绕路是你绕过了自己。"
            },
            {
              "content": "但我绕路的时候也写了些东西——那些不是主角的故事，反而更像我。",
              "echo": {
                "content": "最好的角色不是你以为的主角，是写了一半突然跳出来说"我也有话要说"的那个人。",
                "date": "2017年11月",
                "candidates": 2
              },
              "insight": "你说绕路的那些"反而更像我"——你在原谅自己。绕路也是路，藏着的自己也是自己。"
            },
            {
              "content": "是不是每个人都有一个不敢写的主角？不是因为写不好——是因为写出来，你就得承认那是你。",
              "echo": {
                "content": "写《活着》的时候，福贵不是我，但他说的每一句话都像我在说。",
                "date": "2015年8月",
                "candidates": 1
              },
              "insight": "你意识到每部作品背后都有自己。不敢写主角，是不敢面对自己。"
            },
            {
              "content": "也许写作从来不是为了给别人看。是为了写到第一百页的时候，发现自己是谁。",
              "echo": {
                "content": "我写作不是为了讲一个好故事，是为了搞清楚自己到底在想什么。",
                "date": "2014年3月",
                "candidates": 1
              },
              "insight": "你用十年时间确认了同一件事——写作是关于自我认知。从2014到2023，你没变。"
            }
          ]
        },
        {
          "id": "s6",
          "topic": "就这样吧",
          "date": "2023年10月",
          "time": "晚上 9:40",
          "duration": "28 分钟",
          "momentCount": 4,
          "moments": [
            {
              "content": "今天写完了。不是写好了——是写到我能接受的程度了。不是最好，是够了。",
              "echo": {
                "content": "写完比写好重要。你写完了一本，才知道下一本该怎么写。",
                "date": "2019年1月",
                "candidates": 1
              },
              "insight": "你说"够了"——不是妥协，是有了自己的标准。不是外界说的好，是你自己说的够。"
            },
            {
              "content": "以前我会追求完美。现在不会了。不是因为放弃了——是因为发现完美是一个借口。一个让你不用面对完成的东西。",
              "echo": {
                "content": "最可怕的事情是写不完。比写不好可怕一万倍。",
                "date": "2017年4月",
                "candidates": 2
              },
              "insight": "你把"完美"重新定义为"借口"。六年前你说"最怕写不完"，今天你说"完美让你不用面对完成"——你在和过去的自己对话。"
            },
            {
              "content": "有些东西写出来就知道要删的。但写的过程让我知道了我为什么想写它。删的时候不后悔——因为写的时候已经被它改变了。",
              "echo": {
                "content": "写作不是为了被记住，是为了被改变。",
                "date": "2016年9月",
                "candidates": 1
              },
              "insight": "你不再把被删掉的文字当成浪费。写的过程本身就改变了你。这是一个作家最成熟的姿态。"
            },
            {
              "content": "所以就这样吧。不是结束——是这一轮够了。下一轮会来的，但不是现在。",
              "echo": {
                "content": "以前我觉得每本书都是最后一本。现在我知道——不是的。只要还在想，就还有下一本。",
                "date": "2018年12月",
                "candidates": 1
              },
              "insight": "你说"就这样吧"——这是你给自己的允许。允许停下，允许下一轮在以后来，允许现在就已经够了。"
            }
          ]
        }
      ]
    },
    {
      "id": "c3",
      "name": "隐约有什么…",
      "status": "forming",
      "color": "blue",
      "position": {
        "stars": [
          { "x": 30, "y": 66 },
          { "x": 42, "y": 70 }
        ],
        "labelX": 28,
        "labelY": 74
      },
      "insight": "",
      "stars": []
    }
  ],
  "cta": {
    "mirrorLines": [
      "他用"等"写成了一颗星。",
      "他让"等"变成了"内耗"，又变成了"别扭"，",
      "最后变成了"我看见自己"。",
      "",
      "你的"等"，会变成什么？"
    ],
    "bodyLines": [
      "你写下的每一句 Moment，",
      "都会在某一天变成自己的 Echo，",
      "然后被收进一颗 Star，",
      "和其他的 Star 一起，连成只属于你的 Constellation。"
    ],
    "primaryButton": "建立我的 ego →",
    "secondaryText": "嗯，先看看别人的"
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add h5-virtual-ego/data/example.json
git commit -m "feat(h5): add example data schema and 余华 content"
```

---

### Task 3: Background star field animation

**Files:**
- Create: `h5-virtual-ego/js/background.js`
- Modify: `h5-virtual-ego/css/main.css` (add canvas styles)

- [ ] **Step 1: Add canvas styles to main.css**

Append to `main.css`:

```css
/* === Canvas background === */
#starfield {
  position: fixed;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: 480px;
  height: 100vh;
  z-index: 0;
  pointer-events: none;
}
```

- [ ] **Step 2: Create background.js with star field**

```javascript
// background.js — Canvas-based deep space star field

const STAR_COUNT = 120;
const TWINKLE_RATE = 0.3; // fraction of stars that twinkle

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

    // Subtle glow for larger stars
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

// Start animation when DOM is ready
document.addEventListener('DOMContentLoaded', initStarfield);
```

- [ ] **Step 3: Commit**

```bash
git add h5-virtual-ego/js/background.js h5-virtual-ego/css/main.css
git commit -m "feat(h5): add canvas star field background animation"
```

---

### Task 4: S1 Star map rendering

**Files:**
- Create: `h5-virtual-ego/js/starmap.js` (starmap + constellation rendering)
- Modify: `h5-virtual-ego/css/main.css` (S1 + S2 styles)
- Modify: `h5-virtual-ego/js/main.js` (add renderS1, renderS2)

- [ ] **Step 1: Add S1 + S2 styles to main.css**

Append to `main.css`:

```css
/* === S1: Star Map === */
#s1-starmap {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  text-align: center;
  background: radial-gradient(ellipse at 50% 45%, rgba(180, 160, 120, 0.05) 0%, transparent 55%);
}

.s1-title-top {
  font-size: 10px;
  color: var(--text-dim);
  letter-spacing: 3px;
  margin-bottom: 6px;
}

.s1-title {
  font-size: 20px;
  color: rgba(255, 255, 255, 0.6);
  font-weight: 300;
  letter-spacing: 1px;
}

.s1-hint {
  margin-top: 32px;
  font-size: 10px;
  color: var(--text-faint);
}

/* Constellation container */
.starmap-canvas {
  position: relative;
  width: 100%;
  height: 320px;
  margin: 16px 0;
}

/* Constellation star */
.c-star {
  position: absolute;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  cursor: pointer;
  transition: box-shadow 0.3s ease;
}

.c-star.formed.gold {
  background: var(--gold);
  box-shadow: 0 0 14px rgba(255, 215, 0, 0.7), 0 0 28px rgba(255, 215, 0, 0.25);
}

.c-star.formed.purple {
  background: var(--purple);
  box-shadow: 0 0 14px rgba(201, 176, 255, 0.7), 0 0 28px rgba(201, 176, 255, 0.25);
}

.c-star.forming {
  background: var(--blue);
  box-shadow: 0 0 8px rgba(126, 200, 227, 0.5);
}

/* Constellation line */
.c-line {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.c-line line.formed {
  stroke-opacity: 0.3;
  stroke-width: 0.5;
}

.c-line line.forming {
  stroke-opacity: 0.2;
  stroke-width: 0.5;
  stroke-dasharray: 4, 4;
}

/* Constellation label */
.c-label {
  position: absolute;
  font-size: 10px;
  letter-spacing: 1px;
  transform: translate(-50%, -50%);
  pointer-events: none;
  text-shadow: 0 0 8px rgba(255, 215, 0, 0.3);
}

.c-label.formed.gold { color: rgba(255, 215, 0, 0.65); }
.c-label.formed.purple { color: rgba(201, 176, 255, 0.6); }
.c-label.forming { color: rgba(126, 200, 227, 0.4); font-size: 9px; }

/* === S2: Constellation Detail === */
#s2-constellation {
  background: var(--bg-panel);
}

.s2-back {
  font-size: 9px;
  color: var(--text-faint);
  margin-bottom: 24px;
  cursor: pointer;
}

.s2-header {
  text-align: center;
  margin-bottom: 28px;
}

.s2-header-name {
  font-size: 16px;
  color: rgba(255, 215, 0, 0.75);
  letter-spacing: 1px;
  margin-bottom: 4px;
}

.s2-header-meta {
  font-size: 9px;
  color: var(--text-dim);
}

/* Insight card used in S2 and S3 */
.insight-card {
  padding: 16px;
  border: 1px solid var(--border-gold);
  border-radius: 10px;
  margin-bottom: 24px;
  background: rgba(255, 215, 0, 0.03);
}

.insight-card .insight-label {
  font-size: 9px;
  color: rgba(255, 215, 0, 0.45);
  margin-bottom: 8px;
  letter-spacing: 1px;
}

.insight-card .insight-text {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.65);
  line-height: 1.7;
}

/* Star list section label */
.s2-stars-label {
  font-size: 9px;
  color: var(--text-dim);
  letter-spacing: 1px;
  margin-bottom: 12px;
}

/* Star card */
.star-card {
  padding: 14px 16px;
  border: 1px solid rgba(255, 215, 0, 0.15);
  border-radius: 8px;
  margin-bottom: 10px;
  background: rgba(255, 215, 0, 0.02);
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  transition: border-color 0.2s ease, background 0.2s ease;
}

.star-card:active {
  border-color: rgba(255, 215, 0, 0.35);
  background: rgba(255, 215, 0, 0.05);
}

.star-card-topic {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.65);
}

.star-card-meta {
  font-size: 9px;
  color: var(--text-dim);
  margin-top: 2px;
}

.star-card-arrow {
  font-size: 10px;
  color: rgba(255, 215, 0, 0.35);
}
```

- [ ] **Step 2: Create starmap.js with S1 and S2 rendering**

```javascript
// starmap.js — S1 Star map + S2 Constellation rendering

function renderS1() {
  const section = document.getElementById('s1-starmap');
  const { figure, constellations } = window.appData;

  let html = '';

  // Star map canvas with constellations
  html += '<div class="starmap-canvas" id="starmap-canvas">';

  for (const c of constellations) {
    const { id, status, color, position } = c;
    const pos = position;

    // Draw stars
    for (let i = 0; i < pos.stars.length; i++) {
      const star = pos.stars[i];
      const size = status === 'formed' ? (7 - i) : 5;
      html += `<div class="c-star ${status} ${color}"
        style="left:${star.x}%;top:${star.y}%;width:${size}px;height:${size}px;"
        data-constellation="${id}"></div>`;
    }

    // Draw lines between stars
    if (pos.stars.length >= 2) {
      html += '<svg class="c-line" viewBox="0 0 100 100" preserveAspectRatio="none">';
      for (let i = 0; i < pos.stars.length - 1; i++) {
        const a = pos.stars[i];
        const b = pos.stars[i + 1];
        html += `<line class="${status}" x1="${a.x}" y1="${a.y}" x2="${b.x}" y2="${b.y}" />`;
      }
      html += '</svg>';
    }

    // Label
    html += `<div class="c-label ${status} ${color}" style="left:${pos.labelX}%;top:${pos.labelY}%;">${c.name}</div>`;
  }

  html += '</div>';

  // Title
  html += `<div class="s1-title-top">${figure.title}</div>`;
  html += `<div class="s1-title">${figure.subtitle}</div>`;
  html += '<div class="s1-hint">↓ 点击星座，看里面的星星</div>';

  section.innerHTML = html;

  // Attach click handlers to constellation stars
  const stars = section.querySelectorAll('.c-star.formed');
  stars.forEach(star => {
    star.addEventListener('click', () => {
      const cid = star.dataset.constellation;
      showS2(cid);
    });
  });
}

function showS2(constellationId) {
  const c = window.appData.constellations.find(c => c.id === constellationId);
  if (!c || c.status !== 'formed') return;

  // Hide S1, show S2
  document.getElementById('s1-starmap').classList.add('hidden');
  const s2 = document.getElementById('s2-constellation');
  s2.classList.remove('hidden');
  s2.classList.add('visible');
  s2.scrollIntoView({ behavior: 'smooth' });

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
    html += '<div class="star-card" data-star-id="' + star.id + '">';
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
    section.classList.add('hidden');
    const s1 = document.getElementById('s1-starmap');
    s1.classList.remove('hidden');
    s1.classList.add('visible');
    s1.scrollIntoView({ behavior: 'smooth' });
  });

  // Attach star click handlers
  section.querySelectorAll('.star-card').forEach(card => {
    card.addEventListener('click', () => {
      const starId = card.dataset.starId;
      // Find the star in all constellations
      for (const con of window.appData.constellations) {
        const found = con.stars.find(s => s.id === starId);
        if (found) {
          showS3(con, found);
          return;
        }
      }
    });
  });
}

function renderS2() {
  // Initial empty state — rendered after user clicks a constellation in S1
  const section = document.getElementById('s2-constellation');
  section.innerHTML = '';
}
```

- [ ] **Step 3: Commit**

```bash
git add h5-virtual-ego/js/starmap.js h5-virtual-ego/css/main.css h5-virtual-ego/js/main.js
git commit -m "feat(h5): add S1 star map and S2 constellation rendering"
```

---

### Task 5: S3 Trace rendering

**Files:**
- Create: `h5-virtual-ego/js/trace.js`
- Modify: `h5-virtual-ego/css/main.css` (S3 styles)
- Modify: `h5-virtual-ego/js/main.js` (add renderS3)

- [ ] **Step 1: Add S3 styles to main.css**

Append to `main.css`:

```css
/* === S3: Trace inside a Star === */
#s3-trace {
  background: var(--bg-panel);
}

.s3-back {
  font-size: 9px;
  color: var(--text-faint);
  margin-bottom: 8px;
  cursor: pointer;
  text-align: center;
}

.s3-star-header {
  text-align: center;
  margin-bottom: 28px;
}

.s3-star-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  background: var(--gold);
  border-radius: 50%;
  box-shadow: 0 0 16px rgba(255, 215, 0, 0.8);
  margin-bottom: 8px;
}

.s3-star-topic {
  font-size: 18px;
  color: var(--text-primary);
  letter-spacing: 1px;
}

.s3-star-meta {
  font-size: 9px;
  color: var(--text-dim);
  margin-top: 4px;
}

.s3-trace-badge {
  padding: 10px 16px;
  border: 1px solid rgba(255, 215, 0, 0.1);
  border-radius: 20px;
  text-align: center;
  margin-bottom: 36px;
  background: rgba(255, 215, 0, 0.015);
  font-size: 10px;
  color: rgba(255, 255, 255, 0.25);
  letter-spacing: 1px;
}

/* Moment block */
.moment-block {
  margin-bottom: 24px;
  position: relative;
}

.moment-num {
  position: absolute;
  left: -8px;
  top: -2px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 1px solid rgba(255, 215, 0, 0.25);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 9px;
  color: rgba(255, 215, 0, 0.4);
}

.moment-content {
  padding: 16px;
  border-left: 2px solid rgba(255, 215, 0, 0.25);
  margin: 0 0 10px 8px;
  background: rgba(255, 255, 255, 0.01);
}

.moment-content .label {
  font-size: 9px;
  color: var(--text-dim);
  margin-bottom: 6px;
  letter-spacing: 1px;
}

.moment-content .text {
  font-size: 14px;
  color: var(--text-primary);
  line-height: 1.7;
  font-style: italic;
}

/* Echo card */
.echo-card {
  padding: 12px;
  border: 1px solid rgba(255, 215, 0, 0.1);
  border-radius: 8px;
  margin: 0 0 6px 8px;
  background: rgba(255, 215, 0, 0.015);
}

.echo-card .echo-label {
  font-size: 9px;
  color: rgba(255, 215, 0, 0.35);
  margin-bottom: 4px;
}

.echo-card .echo-text {
  font-size: 11px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.echo-card .echo-date {
  font-size: 8px;
  color: var(--text-faint);
}

.echo-card .echo-candidates {
  margin-top: 4px;
  font-size: 9px;
  color: rgba(255, 215, 0, 0.2);
}

/* S3 insight card (narrower variant) */
.insight-card-sm {
  padding: 12px;
  margin: 0 0 0 8px;
  border: 1px solid rgba(255, 215, 0, 0.12);
  border-radius: 8px;
  background: rgba(255, 215, 0, 0.025);
}

.insight-card-sm .insight-label-sm {
  font-size: 9px;
  color: rgba(255, 215, 0, 0.4);
  margin-bottom: 4px;
}

.insight-card-sm .insight-text-sm {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.55);
  line-height: 1.7;
}

/* Connector */
.trace-connector {
  text-align: center;
  margin: 4px 0 20px 8px;
  font-size: 9px;
  color: rgba(255, 215, 0, 0.18);
}

/* Star stash marker */
.trace-stash {
  text-align: center;
  margin: 12px 0 20px 8px;
  font-size: 9px;
  color: rgba(255, 215, 0, 0.22);
}

/* Star topic summary */
.star-topic-summary {
  padding: 14px;
  border: 1px solid rgba(255, 215, 0, 0.16);
  border-radius: 10px;
  margin: 0 0 20px 8px;
  background: rgba(255, 215, 0, 0.025);
  text-align: center;
}

.star-topic-summary .summary-label {
  font-size: 9px;
  color: var(--text-dim);
  margin-bottom: 4px;
}

.star-topic-summary .summary-topic {
  font-size: 15px;
  color: rgba(255, 215, 0, 0.65);
  letter-spacing: 1px;
}

.star-topic-summary .summary-note {
  font-size: 9px;
  color: var(--text-dim);
  margin-top: 6px;
}

/* Chat entry */
.chat-entry {
  padding: 14px 16px;
  border: 1px solid var(--border-purple);
  border-radius: 8px;
  margin: 0 0 0 8px;
  background: rgba(201, 176, 255, 0.035);
  text-align: center;
}

.chat-entry .chat-title {
  font-size: 11px;
  color: rgba(201, 176, 255, 0.55);
}

.chat-entry .chat-note {
  font-size: 9px;
  color: var(--text-dim);
  margin-top: 4px;
}
```

- [ ] **Step 2: Create trace.js with S3 rendering**

```javascript
// trace.js — S3 Trace rendering inside a Star

function showS3(constellation, star) {
  // Hide S2, show S3
  document.getElementById('s2-constellation').classList.add('hidden');
  const s3 = document.getElementById('s3-trace');
  s3.classList.remove('hidden');
  s3.classList.add('visible');
  s3.scrollIntoView({ behavior: 'smooth' });

  renderS3Content(constellation, star);
}

function renderS3Content(constellation, star) {
  const section = document.getElementById('s3-trace');

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

    // Moment block
    html += '<div class="moment-block">';
    html += `<div class="moment-num">${i + 1}</div>`;

    html += `<div class="moment-content" style="border-left-color:${borderColor}">`;
    html += `<div class="label">${label}</div>`;
    html += `<div class="text">"${escapeHtml(m.content)}"</div>`;
    html += '</div>';

    // Echo card
    html += '<div class="echo-card">';
    html += '<div class="echo-label">你之前也说过类似的</div>';
    html += `<div class="echo-text">"${escapeHtml(m.echo.content)}"</div>`;
    html += `<div class="echo-date">${m.echo.date}</div>`;
    if (m.echo.candidates > 1) {
      html += `<div class="echo-candidates">▸ 之前的你还说过 ${m.echo.candidates} 条</div>`;
    }
    html += '</div>';

    // Insight card
    html += '<div class="insight-card-sm">';
    html += '<div class="insight-label-sm">✦ 我发现</div>';
    html += `<div class="insight-text-sm">${m.insight}</div>`;
    html += '</div>';

    html += '</div>';

    // Connector (between moments, not after last)
    if (i < star.moments.length - 1) {
      html += '<div class="trace-connector">↓ 顺着再想想</div>';
    }
  }

  // Stash marker
  html += '<div class="trace-stash">✦ 收进星图</div>';

  // Star topic summary
  html += '<div class="star-topic-summary">';
  html += '<div class="summary-label">这颗 Star 的主题</div>';
  html += `<div class="summary-topic">${star.topic}</div>`;
  html += `<div class="summary-note">AI 基于 ${star.momentCount} 次连续表达提炼</div>`;
  html += '</div>';

  // Chat entry
  html += '<div class="chat-entry">';
  html += '<div class="chat-title">💬 和那段时间的 ta 说说话</div>';
  html += `<div class="chat-note">对话基于这颗 Star · ${star.date}${star.time}的那 ${star.duration}</div>`;
  html += '</div>';

  section.innerHTML = html;

  // Attach back handler
  document.getElementById('s3-back').addEventListener('click', () => {
    section.classList.add('hidden');
    const s2 = document.getElementById('s2-constellation');
    s2.classList.remove('hidden');
    s2.classList.add('visible');
    s2.scrollIntoView({ behavior: 'smooth' });
  });

  // Animate moment blocks in sequence
  const blocks = section.querySelectorAll('.moment-block');
  blocks.forEach((block, i) => {
    block.style.opacity = '0';
    block.style.transform = 'translateY(16px)';
    block.style.transition = 'opacity 0.4s ease, transform 0.4s ease';
    setTimeout(() => {
      block.style.opacity = '1';
      block.style.transform = 'translateY(0)';
    }, 150 * i);
  });
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}

function renderS3() {
  // Initial empty state
  const section = document.getElementById('s3-trace');
  section.innerHTML = '';
}
```

- [ ] **Step 3: Commit**

```bash
git add h5-virtual-ego/js/trace.js h5-virtual-ego/css/main.css h5-virtual-ego/js/main.js
git commit -m "feat(h5): add S3 trace rendering with moment-echo-insight sequence"
```

---

### Task 6: S4 CTA section

**Files:**
- Modify: `h5-virtual-ego/css/main.css` (S4 styles)
- Modify: `h5-virtual-ego/js/main.js` (add renderS4)

- [ ] **Step 1: Add S4 styles to main.css**

Append to `main.css`:

```css
/* === S4: CTA === */
#s4-cta {
  background: linear-gradient(to bottom, var(--bg-panel), #141428);
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
}

.s4-top {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.35);
  line-height: 2.2;
  margin-bottom: 16px;
}

.s4-mirror {
  margin: 32px 0;
  font-size: 15px;
  color: rgba(255, 255, 255, 0.55);
  line-height: 1.8;
  font-weight: 300;
}

.s4-body {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.25);
  line-height: 2;
  margin-bottom: 36px;
}

.s4-btn {
  display: inline-block;
  padding: 14px 44px;
  border: 1px solid rgba(255, 215, 0, 0.4);
  border-radius: 24px;
  font-size: 13px;
  color: rgba(255, 215, 0, 0.75);
  letter-spacing: 1px;
  cursor: pointer;
  transition: background 0.2s ease, border-color 0.2s ease;
  text-decoration: none;
  margin-bottom: 16px;
}

.s4-btn:active {
  background: rgba(255, 215, 0, 0.08);
  border-color: rgba(255, 215, 0, 0.6);
}

.s4-dismiss {
  font-size: 10px;
  color: rgba(255, 255, 255, 0.1);
  cursor: pointer;
}
```

- [ ] **Step 2: Implement renderS4 in main.js**

Add this function to `main.js`:

```javascript
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

  html += `<a class="s4-btn" id="s4-primary" href="#">${cta.primaryButton}</a>`;
  html += `<div class="s4-dismiss" id="s4-dismiss">${cta.secondaryText}</div>`;

  section.innerHTML = html;

  // Primary CTA - app download link (placeholder)
  document.getElementById('s4-primary').addEventListener('click', (e) => {
    e.preventDefault();
    // Replace with actual app download link
    // window.location.href = 'https://ego.app/download';
  });

  // Dismiss - scroll back to S1
  document.getElementById('s4-dismiss').addEventListener('click', () => {
    document.getElementById('s1-starmap').scrollIntoView({ behavior: 'smooth' });
  });
}
```

- [ ] **Step 3: Commit**

```bash
git add h5-virtual-ego/css/main.css h5-virtual-ego/js/main.js
git commit -m "feat(h5): add S4 CTA section with mirror text and conversion buttons"
```

---

### Task 7: Final integration and README

**Files:**
- Modify: `h5-virtual-ego/js/main.js` (finalize renderAll)
- Modify: `h5-virtual-ego/README.md` (complete documentation)

- [ ] **Step 1: Finalize main.js with complete renderAll and navigation**

Replace `renderAll` in `main.js`:

```javascript
function renderAll() {
  renderS1();
  // S2, S3 rendered on demand via user clicks
  // S4 displayed after user has scrolled through content
  renderS4();
}
```

Add S4 reveal logic to `init`:

```javascript
async function init() {
  try {
    const res = await fetch('data/example.json');
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    appData = await res.json();
  } catch (e) {
    console.error('Failed to load data:', e);
    showError('数据加载失败，请稍后重试。');
    return;
  }

  renderAll();
  initScrollObserver();
  initS4Reveal();
}

function initS4Reveal() {
  // Reveal S4 when user reaches bottom of S3 or after S2 interaction
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible');
        observer.unobserve(entry.target);
      }
    });
  }, { threshold: 0.1 });

  observer.observe(document.getElementById('s4-cta'));
}
```

- [ ] **Step 2: Update README with complete documentation**

```markdown
# h5-virtual-ego

H5 虚拟热点人士 ego 旁观体验页。数据驱动，可复用模板。

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

## Data Schema

```typescript
interface H5Data {
  figure: {
    name: string;        // Figure name, e.g. "余华"
    title: string;       // "如果 XX 有 ego"
    subtitle: string;    // "他的星图长这样"
  };
  constellations: Constellation[];  // Exactly 3 required
  cta: CTA;
}

interface Constellation {
  id: string;            // Unique ID, e.g. "c1"
  name: string;          // e.g. "关于等待"
  status: "formed" | "forming";
  color: "gold" | "purple" | "blue";
  position: {
    stars: { x: number; y: number }[];  // Percentage positions (0-100)
    labelX: number;      // Label position (%)
    labelY: number;
  };
  insight: string;       // Constellation-level AI insight
  stars: Star[];         // 3 stars per formed constellation
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
  mirrorLines: string[];  // Mirror text lines (use "" for blank line)
  bodyLines: string[];    // Body explanation lines
  primaryButton: string;  // e.g. "建立我的 ego →"
  secondaryText: string;  // e.g. "嗯，先看看别人的"
}
```

## Design

- **Baseline:** 375px mobile viewport, max-width 480px centered
- **Visual:** Deep space (#07071a) + gold (#ffd700) + purple (#c9b0ff) + blue (#7ec8e3)
- **Typography:** System font stack with PingFang SC priority for Chinese
- **Animation:** Canvas star twinkling, Intersection Observer section reveals, sequential moment block fade-in
```

- [ ] **Step 3: Commit**

```bash
git add h5-virtual-ego/js/main.js h5-virtual-ego/README.md
git commit -m "docs(h5): finalize main.js integration and README documentation"
```

---

### Task 8: Visual verification

- [ ] **Step 1: Start a local server**

```bash
cd h5-virtual-ego && python3 -m http.server 8080 &
```
Expected: Server starts on port 8080

- [ ] **Step 2: Open in browser and verify S1**

Navigate to `http://localhost:8080`.
Expected:
- Deep space background with twinkling stars
- 3 constellations visible: "关于等待" (gold, formed), "关于写作" (purple, formed), "隐约有什么…" (blue, forming)
- Stars positioned correctly with connecting lines
- "如果 余华 有 ego" title visible
- "↓ 点击星座，看里面的星星" hint visible

- [ ] **Step 3: Click "关于等待" constellation → verify S2**

Expected:
- S1 hidden, S2 visible with smooth scroll
- "✦ 关于等待" header + "3 颗星" meta
- Constellation insight card visible
- 3 star cards: "等一个人看见我", "翻了几百页，封面都没打开", "别扭"
- "← 星图全景" back link functional

- [ ] **Step 4: Click first star → verify S3**

Expected:
- S2 hidden, S3 visible with smooth scroll
- Star header: "等一个人看见我", date/time/duration
- "一次连续的 Trace · 5 个 Moment" badge
- 5 moment blocks with sequential fade-in animation
- Each block: Moment (italic left-border) → Echo card → Insight card
- "↓ 顺着再想想" connectors between blocks
- "✦ 收进星图" marker, Star topic summary, Chat entry
- "← 关于等待" back link functional

- [ ] **Step 5: Scroll to S4 → verify CTA**

Expected:
- CTA section visible with gradient background
- Mirror text referencing the specific constellation content
- Body text explaining Moment → Echo → Star → Constellation chain
- "建立我的 ego →" button visible
- "嗯，先看看别人的" dismiss text visible

- [ ] **Step 6: Commit final verification**

```bash
git add -A
git commit -m "verify(h5): confirm all 4 sections render correctly"
```

---

## Implementation Notes

1. **JSON data validity**: The template depends on well-formed JSON. A `forming` Constellation (id="c3") has no `stars` array content — S2 won't render for it. Only `formed` constellations respond to clicks.

2. **Echo matching time order**: Echo dates must be earlier than the parent Moment date. The data models "past words matching current words."

3. **Navigation state**: Section visibility is managed by toggling `.hidden` class. There is no URL routing — all state is in DOM.

4. **Performance**: 120 canvas stars is intentionally low for mobile performance. Star positions are fixed after generation. No parallax on scroll to keep rendering cheap.

5. **CTA link**: The primary button's `href` is a placeholder (`#`). Replace with the actual app download URL before deployment.

6. **XSS prevention**: The `escapeHtml()` function in trace.js sanitizes user content. All JSON content is rendered via textContent-equivalent patterns.
