import { Hono } from 'hono'

const app = new Hono()

// 错误处理中间件
app.use('*', async (c, next) => {
  try {
    await next()
  } catch (err) {
    console.error('Error details:', {
      message: err.message,
      stack: err.stack,
      name: err.name
    })
    return c.text('Internal Server Error: ' + err.message, 500)
  }
})

// 静态文件服务
app.get('/', (c) => {
  return c.redirect('/index.html')
})

app.get('/index.html', async (c) => {
  const html = await c.env.ASSETS.fetch(new URL('/index.html', c.req.url))
  return html
})

// API 路由
app.get('/api/health', (c) => {
  return c.json({ status: 'ok' })
})

// 文章相关 API
app.get('/api/articles', async (c) => {
  try {
    const { DB } = c.env
    const articles = await DB.prepare('SELECT * FROM articles ORDER BY created_at DESC').all()
    return c.json(articles)
  } catch (err) {
    console.error('Error fetching articles:', err)
    throw err
  }
})

app.post('/api/articles', async (c) => {
  try {
    const { DB } = c.env
    const { title, content } = await c.req.json()
    
    const result = await DB.prepare(
      'INSERT INTO articles (title, content, created_at) VALUES (?, ?, datetime("now"))'
    ).bind(title, content).run()
    
    return c.json({ id: result.meta.last_row_id })
  } catch (err) {
    console.error('Error creating article:', err)
    throw err
  }
})

app.get('/api/articles/:id', async (c) => {
  try {
    const { DB } = c.env
    const id = c.req.param('id')
    
    const article = await DB.prepare('SELECT * FROM articles WHERE id = ?').bind(id).first()
    if (!article) {
      return c.json({ error: 'Article not found' }, 404)
    }
    
    return c.json(article)
  } catch (err) {
    console.error('Error fetching article:', err)
    throw err
  }
})

// KV 存储相关 API
app.get('/api/cache/:key', async (c) => {
  try {
    const { AIRTEST_KV } = c.env
    const key = c.req.param('key')
    
    const value = await AIRTEST_KV.get(key)
    if (!value) {
      return c.json({ error: 'Key not found' }, 404)
    }
    
    return c.json({ value })
  } catch (err) {
    console.error('Error getting cache:', err)
    throw err
  }
})

app.put('/api/cache/:key', async (c) => {
  try {
    const { AIRTEST_KV } = c.env
    const key = c.req.param('key')
    const { value } = await c.req.json()
    
    await AIRTEST_KV.put(key, value)
    return c.json({ success: true })
  } catch (err) {
    console.error('Error setting cache:', err)
    throw err
  }
})

export default app 