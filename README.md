# Hacker News AI 助手

一个基于 Hacker News 的中文日报项目，每天自动抓取 Hacker News 热门文章及评论，通过 AI 生成中文解读与总结，传递科技前沿信息。

数据库保存结合了 https://github.com/TwoThreeWang/go_simple_forum 这个项目，直接保存到了论坛数据库，实现了联动发布文章。

## 功能特性

- 🔄 自动抓取 Hacker News 热门文章
- 🤖 使用 Google Gemini AI 生成中文摘要
- 📝 自动生成每日科技新闻精选
- 💾 支持 PostgreSQL 数据持久化
- 🎯 支持自定义文章抓取数量

## 技术栈

- Go 1.23.4
- Google Gemini AI
- PostgreSQL
- GORM

## 安装说明

1. 克隆项目
```bash
git clone https://github.com/TwoThreeWang/hacker-news-ai.git
cd hacker-news-ai
```

2. 安装依赖
```bash
go mod download
```

3. 配置数据库
- 创建 PostgreSQL 数据库
- 导入数据库表结构（需要自行创建）

4. 配置项目
- 复制 `config/config_ex.json` 为 `config/config.json`
- 修改配置文件中的相关参数：
  - `gemini_api_key`: Google Gemini API 密钥
  - `hn_api_base_url`: Hacker News API 地址
  - `top_stories_limit`: 每日获取的热门文章数量
  - 数据库相关配置

## 使用说明

1. 启动项目
```bash
go run main.go
```

2. 项目会自动执行以下操作：
- 从 Hacker News 获取热门文章
- 使用 AI 生成中文摘要
- 生成每日科技新闻精选
- 保存到数据库

## 配置说明

```json
{
  "gemini_api_key": "your_api_key",
  "hn_api_base_url": "https://hacker-news.firebaseio.com/v0",
  "top_stories_limit": 30,
  "db_host": "localhost",
  "db_port": 5432,
  "db_user": "postgres",
  "db_password": "your_password",
  "db_name": "your_database"
}
```

## 项目结构

```
.
├── config/          # 配置文件和配置管理
├── database/        # 数据库操作封装
├── models/          # 数据模型定义
├── services/        # 业务逻辑服务
└── main.go         # 程序入口
```

## 许可证

MIT License