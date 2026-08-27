# 项目长期备忘 - wcj-go-text

## 项目概况
- Wails v2.14.0 桌面应用（Go + Vue 前端），仓库 github.com/wyzzgzhdcxy/wcj-go-text
- CI：`.github/workflows/build.yml`，三平台矩阵构建（windows/macos/ubuntu），push 到 main 或打 v* tag 触发；tag 构建会走 release job 发 Release

## 环境特性（重要）
- 有后台自动同步工具：每隔几分钟自动 commit（message 为 "Sync: <时间戳>"）并 push；它会清掉 `.git/refs/remotes/`，`origin/main` 显示 [gone] 属正常，不影响操作
- 本机无 gh CLI；查 CI 状态用 GitHub 公开 API `api.github.com/repos/wyzzgzhdcxy/wcj-go-text/actions/runs`
- go.mod 要求 go 1.27.0，CI 中 setup-go 需保持 '1.27' 对齐

## CI 历史坑
- 2026-08-26：`dAppServer/wails-build-action@v2` 内部用了废弃的 upload-artifact@v2，导致三平台秒拒。已改为直接 `wails build`（不依赖该三方 action），upload-artifact/download-artifact 用 v4
- 2026-08-26：CI 最终只构建 Windows，且仅在打 v* tag 时触发（避免后台同步工具的 "Sync:" 提交造成空转打包）。发布 Release 流程 = `git tag vX.Y.Z && git push origin vX.Y.Z`
