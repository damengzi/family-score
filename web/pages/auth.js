function renderSetup() {
  app.innerHTML = `<div class="login"><div class="card">
    <h1>初始化家庭德育积分系统</h1>
    <p class="small">数据将保存在本机：${h(state.status.dataDir)}。初始化后管理员账号固定为 admin，密码固定为 654321。</p>
    <div id="setupError" class="small" style="color:#e05260;margin-bottom:10px"></div>
    <form class="form" id="setupForm">
      <div class="field"><label>孩子姓名</label><input name="childName" value="小朋友" required></div>
      <div class="field"><label>孩子年龄</label><input name="childAge" type="number" value="8" min="1" required></div>
      <button>完成初始化</button>
    </form>
  </div></div>`;
  document.getElementById('setupForm').onsubmit = async (e) => {
    e.preventDefault();
    setError('setupError', '');
    const body = Object.fromEntries(new FormData(e.target));
    body.childAge = Number(body.childAge);
    try {
      await api('/api/setup/init', { method: 'POST', body });
      toast('初始化完成，请登录');
      await boot();
    } catch (err) { setError('setupError', err.message); toast(err.message); }
  };
}

function renderLogin() {
  app.innerHTML = `<div class="login"><div class="card">
    <h1>家庭德育积分系统</h1>
    <p class="small">管理员固定账号：admin，固定密码：654321。其他家长/孩子账号由管理员在用户管理中创建。</p>
    <div id="loginError" class="small" style="color:#e05260;margin-bottom:10px"></div>
    <form class="form" id="loginForm">
      <div class="field"><label>登录名</label><input name="loginName" placeholder="如 parent 或你初始化时填写的账号" required autofocus></div>
      <div class="field"><label>密码</label><input name="password" type="password" placeholder="初始化时填写的密码" required></div>
      <button id="loginBtn">登录</button>
    </form>
  </div></div>`;
  document.getElementById('loginForm').onsubmit = async (e) => {
    e.preventDefault();
    setError('loginError', '');
    const btn = document.getElementById('loginBtn');
    btn.disabled = true; btn.textContent = '登录中...';
    try {
      const body = Object.fromEntries(new FormData(e.target));
      const data = await api('/api/auth/login', { method: 'POST', body });
      state.me = data.user;
      toast('登录成功');
      app.innerHTML = '<div class="login"><div class="card"><h1>登录成功</h1><p class="small">正在加载积分系统...</p></div></div>';
      await loadHome();
    } catch (err) {
      setError('loginError', err.message);
      toast(err.message);
    } finally {
      const nextBtn = document.getElementById('loginBtn');
      if (nextBtn) { nextBtn.disabled = false; nextBtn.textContent = '登录'; }
    }
  };
}
