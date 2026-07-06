function renderSetup() {
  document.body.dataset.role = 'guest';
  app.innerHTML = `<div class="login"><div class="card setup-card">
    <h1>初始化家庭德育积分系统</h1>
    <p class="small">数据将保存在本机：${h(state.status.dataDir)}。可一次性创建家庭组、常见家长账号、小明和小红。</p>
    <div id="setupError" class="small" style="color:#e05260;margin-bottom:10px"></div>
    <form class="form" id="setupForm">
      <div class="form two"><div class="field"><label>家庭名称</label><input name="familyName" value="我的家庭" required></div><div class="field"><label>家庭组名称</label><input name="groupName" value="小明小红家" required></div></div>
      <div class="field"><label>管理员密码</label><input name="adminPassword" type="password" value="654321" minlength="4" required><div class="small">管理员登录名仍为 admin；登录页不会展示账号和密码。</div></div>
      <div class="notice">默认会创建爸爸、妈妈、爷爷、奶奶、姥姥、姥爷，以及小明、小红两个孩子账号。初始化后可在用户管理和家庭组中继续调整。</div>
      <label class="row small"><input type="checkbox" name="withDefaults" checked> 创建默认家庭成员、孩子账号、任务和奖励</label>
      <button>完成初始化</button>
    </form>
  </div></div>`;
  document.getElementById('setupForm').onsubmit = async (e) => {
    e.preventDefault();
    setError('setupError', '');
    const form = Object.fromEntries(new FormData(e.target));
    const withDefaults = Boolean(form.withDefaults);
    const groupName = form.groupName || '默认家庭组';
    const body = {familyName: form.familyName, adminPassword: form.adminPassword, groupName, importDefaults: true, parents: [], children: []};
    if (withDefaults) {
      body.parents = ['爸爸','妈妈','爷爷','奶奶','姥姥','姥爷'].map((title, i) => ({displayName:title, loginName:['baba','mama','yeye','nainai','laolao','laoye'][i], password:'123456', parentTitle:title}));
      body.children = [
        {name:'小明', age:8, gender:'BOY', childLoginName:'xiaoming', childPassword:'123456'},
        {name:'小红', age:8, gender:'GIRL', childLoginName:'xiaohong', childPassword:'123456'},
      ];
    }
    try {
      await api('/api/setup/init', { method: 'POST', body });
      toast('初始化完成，请登录');
      await boot();
    } catch (err) { setError('setupError', err.message); toast(err.message); }
  };
}

function renderLogin() {
  document.body.dataset.role = 'guest';
  app.innerHTML = `<div class="login"><div class="card">
    <h1>家庭德育积分系统</h1>
    <p class="small">请输入你的登录名和密码。家长和孩子账号由管理员或家长创建。</p>
    <div id="loginError" class="small" style="color:#e05260;margin-bottom:10px"></div>
    <form class="form" id="loginForm">
      <div class="field"><label>登录名</label><input name="loginName" placeholder="请输入登录名" required autofocus></div>
      <div class="field"><label>密码</label><input name="password" type="password" placeholder="请输入密码" required></div>
      <button id="loginBtn">登录</button>
      <button type="button" class="secondary" id="forgotBtn">忘记密码</button>
    </form>
  </div></div>`;
  document.getElementById('forgotBtn').onclick = () => renderForgotPassword();
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

async function renderForgotPassword() {
  document.body.dataset.role = 'guest';
  app.innerHTML = `<div class="login"><div class="card">
    <h1>重置密码</h1>
    <p class="small">输入登录名，选择正确的图片验证码后设置新密码。</p>
    <div id="resetError" class="small" style="color:#e05260;margin-bottom:10px"></div>
    <form class="form" id="resetForm">
      <div class="field"><label>登录名</label><input name="loginName" placeholder="请输入登录名" required autofocus></div>
      <div class="field"><label>新密码</label><input name="password" type="password" minlength="4" placeholder="至少 4 位" required></div>
      <div class="field"><label id="captchaPrompt">图片验证码加载中...</label><div class="row" id="captchaChoices"></div></div>
      <button id="resetBtn">重置密码</button>
      <button type="button" class="secondary" id="backLoginBtn">返回登录</button>
    </form>
  </div></div>`;
  document.getElementById('backLoginBtn').onclick = () => renderLogin();
  const captcha = await api('/api/auth/password-captcha');
  let selected = '';
  document.getElementById('captchaPrompt').textContent = captcha.prompt;
  document.getElementById('captchaChoices').innerHTML = (captcha.choices || []).map(c => `<button type="button" class="secondary captcha-choice" data-captcha="${h(c.key)}"><img src="${h(c.image)}" alt="验证码选项" width="96" height="64"></button>`).join('');
  document.querySelectorAll('[data-captcha]').forEach(btn => btn.onclick = () => {
    selected = btn.dataset.captcha;
    document.querySelectorAll('[data-captcha]').forEach(x => x.classList.remove('active'));
    btn.classList.add('active');
  });
  document.getElementById('resetForm').onsubmit = async (e) => {
    e.preventDefault();
    setError('resetError', '');
    if (!selected) { setError('resetError', '请选择图片验证码'); return; }
    const body = Object.fromEntries(new FormData(e.target));
    body.captchaToken = captcha.token;
    body.captchaAnswer = selected;
    try {
      await api('/api/auth/reset-password', { method: 'POST', body });
      toast('密码已重置，请重新登录');
      renderLogin();
    } catch (err) {
      setError('resetError', err.message);
      toast(err.message);
    }
  };
}
