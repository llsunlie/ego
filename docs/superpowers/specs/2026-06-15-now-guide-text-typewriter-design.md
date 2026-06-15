# Now Page Guide Text Typewriter

**Date**: 2026-06-15
**Status**: approved
**Scope**: `client/lib/features/now/widgets/guide_section.dart`

## Problem

The idle state guide text on the now page shows a single static string "有什么想说的吗". It's monotonous and doesn't give users enough variety to spark writing inspiration.

## Solution

Replace the static `GuideText` widget with a multi-text typewriter carousel — 7 grounded, everyday prompts that type out character by character with a blinking cursor, pause for reading, delete, and cycle to the next.

### Text Candidates

| # | Text | Intent |
|---|------|--------|
| 1 | 有什么想说的吗？ | General invitation |
| 2 | 刚才做了什么？ | Prompt to record recent action |
| 3 | 今天有什么新鲜事？ | Find the special in the ordinary |
| 4 | 此刻在想什么？ | Capture current thoughts |
| 5 | 心情怎么样？ | Emotional expression |
| 6 | 有什么想记住的？ | Worth-preserving moments |
| 7 | 随便写点什么吧 | Low-pressure invitation |

### Typewriter State Machine

```
typing → pause → deleting → switch → (next text) → typing → ...
```

| Phase | Behavior | Duration |
|-------|----------|----------|
| typing | Append one character + blink cursor | 120ms/char |
| pause | Full text visible, cursor blinks | 2.5s |
| deleting | Remove one character + blink cursor | 60ms/char |
| switch | Brief empty pause before next text | 400ms |

- Initial delay: 1s after mount before first type
- Start index: random (0–6) to avoid always showing #1

### Widget Architecture

```
GuideText: StatelessWidget → StatefulWidget
  - _currentTextIndex: int
  - _displayedChars: int
  - _phase: enum { typing, pause, deleting, switching }
  - _timer: Timer (chained, not periodic)

build():
  → FractionallySizedBox (same layout as before)
    → Text(_currentText.substring(0, _displayedChars) + cursor)
```

### Lifecycle

- `GuideText` is only mounted when `NowPageStatus == idle` (existing `if (isIdle)` guard in `now_page.dart`)
- `initState()`: pick random start index, begin type cycle after 1s delay
- `dispose()`: cancel timer
- No provider changes needed — animation is purely a presentation concern

### Non-changes

- No proto changes
- No server changes
- No database changes
- No route changes
- Provider (`NowPageNotifier`) unchanged
- Parent `now_page.dart` unchanged
- Existing `WriteButton` unchanged
- Text style (font size, color, letter spacing, font weight) unchanged
- Layout (FractionallySizedBox, Alignment) unchanged
