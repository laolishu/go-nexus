# CHANGELOG

## [Unreleased]

### Added
- 初始项目结构
- 基础配置系统
- Wire 依赖注入框架集成
- Docker 部署支持
- 基础数据模型定义

### Changed

### Deprecated

### Removed

### Fixed

### Security

## 版本说明

本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/) 规范。

## 发布说明

### 如何发布新版本

1. 更新 `VERSION` 文件中的版本号
2. 更新 `CHANGELOG.md`，将 `Unreleased` 部分移动到新版本下
3. 提交更改：`git commit -am "Release vX.Y.Z"`
4. 创建标签：`git tag vX.Y.Z`
5. 推送到远程：`git push && git push --tags`
6. GitHub Actions 将自动构建和发布

### 版本命名规则

- 主版本号：当你做了不兼容的 API 修改
- 次版本号：当你做了向下兼容的功能性新增
- 修订号：当你做了向下兼容的问题修正