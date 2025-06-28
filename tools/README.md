# Tools

本目录包含开发、文档和运维相关工具。

## docs（Docusaurus 文档系统）

- 位置：`tools/docs`
- 说明：基于 Docusaurus 的文档系统，所有架构、开发、API 文档请放在此处。
- 本地启动：
  ```bash
  cd tools/docs
  npm start
  ```
- 构建静态站点：
  ```bash
  cd tools/docs
  npm run build
  ```
- 部署到 GitHub Pages：
  ```bash
  cd tools/docs
  npm run deploy
  ```

如需自动化部署，请参考 `.github/workflows/docs.yml`.
