export async function onRequest(context) {
  const { request, env } = context;
  const url = new URL(request.url);
  const path = url.pathname.replace('/api/articles', '');

  switch (request.method) {
    case 'GET':
      if (path === '') {
        return handleListArticles(context);
      }
      return handleGetArticle(context);
    case 'POST':
      return handleCreateArticle(context);
    case 'PUT':
      return handleUpdateArticle(context);
    case 'DELETE':
      return handleDeleteArticle(context);
    default:
      return new Response('Method not allowed', { status: 405 });
  }
}

async function handleListArticles({ env }) {
  const articles = await env.DB.prepare(
    'SELECT * FROM articles ORDER BY created_at DESC'
  ).all();
  return new Response(JSON.stringify(articles), {
    headers: { 'Content-Type': 'application/json' }
  });
}

async function handleGetArticle({ env, params }) {
  const article = await env.DB.prepare(
    'SELECT * FROM articles WHERE slug = ?'
  ).bind(params.slug).first();
  
  if (!article) {
    return new Response('Article not found', { status: 404 });
  }
  
  return new Response(JSON.stringify(article), {
    headers: { 'Content-Type': 'application/json' }
  });
}

async function handleCreateArticle({ request, env }) {
  const article = await request.json();
  
  try {
    await env.DB.prepare(
      'INSERT INTO articles (title, slug, content, category, draft) VALUES (?, ?, ?, ?, ?)'
    ).bind(
      article.title,
      article.slug,
      article.content,
      article.category,
      article.draft
    ).run();
    
    return new Response(JSON.stringify(article), {
      status: 201,
      headers: { 'Content-Type': 'application/json' }
    });
  } catch (error) {
    return new Response('Error creating article: ' + error.message, { status: 500 });
  }
}

async function handleUpdateArticle({ request, env, params }) {
  const article = await request.json();
  
  try {
    await env.DB.prepare(
      'UPDATE articles SET title = ?, content = ?, category = ?, draft = ? WHERE slug = ?'
    ).bind(
      article.title,
      article.content,
      article.category,
      article.draft,
      params.slug
    ).run();
    
    return new Response(JSON.stringify(article), {
      headers: { 'Content-Type': 'application/json' }
    });
  } catch (error) {
    return new Response('Error updating article: ' + error.message, { status: 500 });
  }
}

async function handleDeleteArticle({ env, params }) {
  try {
    await env.DB.prepare(
      'DELETE FROM articles WHERE slug = ?'
    ).bind(params.slug).run();
    
    return new Response(null, { status: 204 });
  } catch (error) {
    return new Response('Error deleting article: ' + error.message, { status: 500 });
  }
} 