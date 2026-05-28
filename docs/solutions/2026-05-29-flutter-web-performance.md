# Flutter Web 性能优化

> 2026-05-29 | PR #37 #39 #40 #41

## 问题

首次加载 ~11MB，9s+，中文方框乱码，Google CDN 字体请求。

## 优化项

1. **Logo PNG → WebP**：1.4MB → 86KB
2. **字体子集化**：完整 NotoSansSC 4MB → 简体子集 1.1MB（[gwfh](https://gwfh.mranftl.com/fonts)）
3. **Dingbats fallback**：引入 NotoSansSymbols2 382KB，覆盖 `✦` 等符号，配合 `fontFamilyFallback`
4. **Button textStyle 修复**：`styleFrom(textStyle:)` 替换式覆盖，需显式加 `fontFamily` + `fontFamilyFallback`
5. **CustomPainter/Overlay 字体**：`TextPainter` 和 `showDialog` 内需显式指定字体
6. **Flutter build -O4**：最高 JS 优化 + `--no-source-maps`
7. **服务端 Gzip**：`http.FileServer` + gziphandler，main.dart.js 传输 ~1MB
8. **字体并行加载**：`<link rel="preload" as="fetch" crossorigin="anonymous">`，CanvasKit 用 fetch 而非 @font-face 加载字体，as=fetch 匹配加载路径，字体与 JS 并行下载
9. **字体 Preload 无效**：~~`<link rel="preload">`~~ as=font 对 CanvasKit 无效，已移除
10. **Wasm 编译无效**：~~`--wasm`~~ 运行时崩溃，grpc-web 等依赖暂不兼容

## 效果

| 指标 | 优化前 | 优化后 |
|---|---|---|
| 总下载 | ~11MB | ~5MB + gzip |
| main.dart.js | 3.2MB | ~1MB (gzip) |
| 中文字体 | 4MB | 1.5MB |
| Logo | 1.4MB | 86KB |
| CDN 请求 | 多个 | 0 |
| 首帧乱码 | 有 | 无 |
