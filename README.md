# Airtest - Cloudflare Pages + Workers 版本

这是一个使用 Cloudflare Pages 和 Workers 构建的文章管理系统。

## 项目结构

```
.
├── functions/           # Cloudflare Functions
│   └── api/            # API 路由
├── public/             # 静态文件
├── schema.sql          # D1 数据库 schema
└── wrangler.toml       # Cloudflare Workers 配置
```

## 部署步骤

1. 安装 Wrangler CLI:
```bash
npm install -g wrangler
```

2. 登录到 Cloudflare:
```bash
wrangler login
```

3. 创建 D1 数据库:
```bash
wrangler d1 create airtest_db
```

4. 创建 KV 命名空间:
```bash
wrangler kv:namespace create "KV"
```

5. 更新 wrangler.toml 中的数据库 ID 和 KV 命名空间 ID

6. 初始化数据库:
```bash
wrangler d1 execute airtest_db --file=./schema.sql
```

7. 部署到 Cloudflare:
```bash
wrangler deploy
```

## 开发

本地开发:
```bash
wrangler dev
```

## 环境变量

- `DB`: D1 数据库绑定
- `KV`: KV 存储绑定

## API 端点

- GET /api/articles - 获取文章列表
- GET /api/articles/:slug - 获取单篇文章
- POST /api/articles - 创建文章
- PUT /api/articles/:slug - 更新文章
- DELETE /api/articles/:slug - 删除文章
- POST /api/articles/:slug/publish - 发布文章 