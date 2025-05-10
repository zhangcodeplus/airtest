export async function onRequest(context) {
  const { request, env, params } = context;
  
  if (request.method !== 'POST') {
    return new Response('Method not allowed', { status: 405 });
  }

  try {
    const slug = params.slug;
    
    // 更新文章状态
    await env.DB.prepare(
      'UPDATE articles SET draft = false, lastmod = ? WHERE slug = ?'
    ).bind(new Date().toISOString(), slug).run();

    // 获取文章内容
    const article = await env.DB.prepare(
      'SELECT * FROM articles WHERE slug = ?'
    ).bind(slug).first();

    if (!article) {
      return new Response('Article not found', { status: 404 });
    }

    // 同步到其他平台
    await syncToPlatforms(article, env);

    return new Response(JSON.stringify(article), {
      headers: { 'Content-Type': 'application/json' }
    });
  } catch (error) {
    return new Response('Error publishing article: ' + error.message, { status: 500 });
  }
}

async function syncToPlatforms(article, env) {
  const config = await env.KV.get('platform_config', { type: 'json' });
  
  for (const platform of article.platforms) {
    switch (platform) {
      case 'wechat':
        await syncToWeChat(article, config.wechat);
        break;
      case 'zhihu':
        await syncToZhihu(article, config.zhihu);
        break;
      case 'xiaohongshu':
        await syncToXiaohongshu(article, config.xiaohongshu);
        break;
    }
  }
}

async function syncToWeChat(article, config) {
  // 实现微信公众号同步逻辑
  // 使用微信公众号API
}

async function syncToZhihu(article, config) {
  // 实现知乎同步逻辑
  // 使用知乎API
}

async function syncToXiaohongshu(article, config) {
  // 实现小红书同步逻辑
  // 使用小红书API
} 