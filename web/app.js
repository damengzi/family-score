boot().catch(err => {
  app.innerHTML = `<div class="login"><div class="card"><h1>启动失败</h1><p>${h(err.message)}</p></div></div>`;
});
