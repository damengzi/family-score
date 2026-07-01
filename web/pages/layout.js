function renderApp() {
  if (!canOperate() && ['score', 'childConfig', 'taskConfig', 'rewardConfig', 'users', 'system'].includes(state.tab)) state.tab = 'detail';
  if (!isAdmin() && state.tab === 'users') state.tab = 'detail';
  const acc = state.dashboard?.account || {};
  const childOptions = state.children.map(c => `<option value="${c.id}" ${c.id === state.childId ? 'selected' : ''}>${h(c.name)} ${c.age}岁</option>`).join('');
  const operateTabs = canOperate() ? `${tabBtn('childConfig','孩子管理')}${tabBtn('taskConfig','任务自定义')}${tabBtn('rewardConfig','奖励/零食自定义')}${tabBtn('system','本机备份')}` : '';
  const adminTabs = isAdmin() ? `${tabBtn('users','用户管理')}` : '';
  app.innerHTML = `<div class="layout">
    <div class="hero">
      <div><h1>家庭德育积分系统</h1><p>积分明细 · 加扣分 · 惩罚修复 · 任务配置 · 奖励兑换 · 本机备份</p></div>
      <div class="row"><span class="tag ${isAdmin() ? 'green' : 'blue'}">${roleName(state.me?.role)}</span><span class="tag">当前用户：${h(state.me?.name || '-')}</span><select id="childSelect">${childOptions}</select><button class="secondary" id="logoutBtn">退出</button></div>
    </div>
    <div class="grid">
      ${metric('基准德育分', acc.baseScore ?? 100, `<span class="tag ${statusClass(acc.statusLevel)}">${statusText(acc.statusLevel)}</span>`)}
      ${metric('超额兑换分', acc.bonusScore ?? 0, '<span class="small">用于兑换零食/奖励</span>')}
      ${metric('星星', acc.starCount ?? 0, '<span class="small">长期大奖</span>')}
      ${metric('家庭小队分', acc.teamScore ?? 0, '<span class="small">月度评级</span>')}
    </div>
    <div class="tabs">
      ${tabBtn('guide','分值说明')}${tabBtn('detail','积分明细')}${canOperate() ? tabBtn('score','加分/扣分/惩罚') : ''}${tabBtn('tasks','今日任务')}${tabBtn('rewards','奖励兑换')}${operateTabs}${adminTabs}
    </div>
    ${renderTab()}
  </div>`;
  document.getElementById('childSelect').onchange = async (e) => { state.childId = Number(e.target.value); await loadAll(); };
  document.getElementById('logoutBtn').onclick = async () => { await api('/api/auth/logout', {method:'POST', body:{}}); state.me=null; renderLogin(); };
  bindEvents();
}

function metric(label, value, extra) { return `<div class="card metric"><div class="label">${label}</div><div class="value">${h(value)}</div><div>${extra}</div></div>`; }
function tabBtn(key, text) { return `<button class="tab ${state.tab===key?'active':''}" data-tab="${key}">${text}</button>`; }
function renderTab() {
  if (state.tab === 'guide') return renderScoreGuide();
  if (state.tab === 'score') return renderScoreForm();
  if (state.tab === 'tasks') return renderTasks();
  if (state.tab === 'childConfig') return renderChildren();
  if (state.tab === 'taskConfig') return renderTaskConfig();
  if (state.tab === 'rewards') return renderRewards();
  if (state.tab === 'rewardConfig') return renderRewardConfig();
  if (state.tab === 'users') return renderUsers();
  if (state.tab === 'system') return renderSystem();
  return renderDetail();
}

function bindEvents() {
  document.querySelectorAll('[data-tab]').forEach(b => b.onclick = async () => { state.tab = b.dataset.tab; renderApp(); });
  const scoreForm = document.getElementById('scoreForm');
  if (scoreForm) scoreForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); body.childId = state.childId; body.scoreChange = Number(body.scoreChange); body.targetAccount = body.recordType === 'TEAM' ? 'TEAM' : body.recordType === 'STAR' ? 'STAR' : 'AUTO'; try { await api('/api/score-records', {method:'POST', body}); toast('记录成功'); await loadAll(); } catch(err) { toast(err.message); } };
  const childForm = document.getElementById('childForm');
  if (childForm) childForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); body.age = Number(body.age); body.parentUserId = Number(body.parentUserId || 0); try { await api('/api/children', {method:'POST', body}); toast('孩子已新增'); await loadHome(); } catch(err) { toast(err.message); } };
  const taskTemplateForm = document.getElementById('taskTemplateForm');
  if (taskTemplateForm) taskTemplateForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); body.scoreValue = Number(body.scoreValue); try { await api('/api/task-templates', {method:'POST', body}); toast('任务模板已新增'); await loadAll(); } catch(err) { toast(err.message); } };
  const rewardForm = document.getElementById('rewardForm');
  if (rewardForm) rewardForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); ['costScore','costStar','weeklyLimit','monthlyLimit'].forEach(k => body[k] = Number(body[k])); try { await api('/api/rewards', {method:'POST', body}); toast('奖励已新增'); await loadAll(); } catch(err) { toast(err.message); } };
  const userRole = document.getElementById('userRole');
  if (userRole) userRole.onchange = () => { const field = document.getElementById('bindChildField'); if (field) field.style.display = userRole.value === 'CHILD' ? 'block' : 'none'; };
  const userForm = document.getElementById('userForm');
  if (userForm) userForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); body.childId = Number(body.childId || 0); try { await api('/api/users', {method:'POST', body}); toast('用户已注册'); await loadAll(); } catch(err) { toast(err.message); } };
  document.querySelectorAll('[data-save-user]').forEach(b => b.onclick = async () => { const id = Number(b.dataset.saveUser); const body = {displayName: document.querySelector(`[data-user-name="${id}"]`)?.value || '', childId: Number(document.querySelector(`[data-user-child="${id}"]`)?.value || 0), password: document.querySelector(`[data-user-pass="${id}"]`)?.value || ''}; try { const data = await api(`/api/users/${id}`, {method:'PATCH', body}); if (data.user?.id === state.me.userId) state.me.name = data.user.name; toast('用户已保存'); await loadAll(); } catch(err) { toast(err.message); } });
  document.querySelectorAll('[data-save-child]').forEach(b => b.onclick = async () => { const id = Number(b.dataset.saveChild); const body = {name: document.querySelector(`[data-child-name="${id}"]`)?.value || '', age: Number(document.querySelector(`[data-child-age="${id}"]`)?.value || 0), gender: document.querySelector(`[data-child-gender="${id}"]`)?.value || 'BOY', parentUserId: Number(document.querySelector(`[data-child-parent="${id}"]`)?.value || 0)}; try { await api(`/api/children/${id}`, {method:'PATCH', body}); toast('孩子档案已保存'); await loadHome(); } catch(err) { toast(err.message); } });
  document.querySelectorAll('[data-del-child]').forEach(b => b.onclick = async () => { if (!confirm('确认删除该孩子？相关积分、任务、兑换记录也会删除，绑定的孩子账号会被注销。')) return; try { await api(`/api/children/${b.dataset.delChild}`, {method:'DELETE'}); toast('孩子已删除'); if (state.childId === Number(b.dataset.delChild)) state.childId = null; await loadHome(); } catch(err) { toast(err.message); } });
  document.querySelectorAll('[data-del-task-template]').forEach(b => b.onclick = async () => { if (!confirm('确认删除该任务模板？')) return; try { await api(`/api/task-templates/${b.dataset.delTaskTemplate}`, {method:'DELETE'}); toast('已删除'); await loadAll(); } catch(err) { toast(err.message); } });
  document.querySelectorAll('[data-del-reward]').forEach(b => b.onclick = async () => { if (!confirm('确认删除该奖励？')) return; try { await api(`/api/rewards/${b.dataset.delReward}`, {method:'DELETE'}); toast('已删除'); await loadAll(); } catch(err) { toast(err.message); } });
  document.querySelectorAll('[data-del-user]').forEach(b => b.onclick = async () => { if (!confirm('确认注销该用户？注销后该账号不能登录。')) return; try { await api(`/api/users/${b.dataset.delUser}`, {method:'DELETE'}); toast('用户已注销'); await loadAll(); } catch(err) { toast(err.message); } });
  document.querySelectorAll('[data-submit-task]').forEach(b => b.onclick = async () => { try { await api(`/api/tasks/${b.dataset.submitTask}/submit`, {method:'POST', body:{submitNote:'已完成'}}); toast('已提交，等待家长确认'); await loadAll(); } catch(err) { toast(err.message); } });
  document.querySelectorAll('[data-audit-task]').forEach(b => b.onclick = async () => { try { await api(`/api/tasks/${b.dataset.auditTask}/audit`, {method:'POST', body:{result:'APPROVED', auditNote:'通过'}}); toast('任务已通过'); await loadAll(); } catch(err) { toast(err.message); } });
  document.querySelectorAll('[data-reject-task]').forEach(b => b.onclick = async () => { try { await api(`/api/tasks/${b.dataset.rejectTask}/audit`, {method:'POST', body:{result:'REJECTED', auditNote:'暂不通过'}}); toast('任务已驳回'); await loadAll(); } catch(err) { toast(err.message); } });
  document.querySelectorAll('[data-reward]').forEach(b => b.onclick = async () => { try { await api('/api/exchange-orders', {method:'POST', body:{childId:state.childId, rewardId:Number(b.dataset.reward), note:'申请兑换'}}); toast('已提交兑换申请，需家长审核'); await loadAll(); } catch(err) { toast(err.message); } });
  document.querySelectorAll('[data-audit-order]').forEach(b => b.onclick = async () => { try { await api(`/api/exchange-orders/${b.dataset.auditOrder}/audit`, {method:'POST', body:{result:'APPROVED', auditNote:'通过'}}); toast('兑换已通过'); await loadAll(); } catch(err) { toast(err.message); } });
  document.querySelectorAll('[data-reject-order]').forEach(b => b.onclick = async () => { try { await api(`/api/exchange-orders/${b.dataset.rejectOrder}/audit`, {method:'POST', body:{result:'REJECTED', auditNote:'驳回'}}); toast('兑换已驳回'); await loadAll(); } catch(err) { toast(err.message); } });
  const backupBtn = document.getElementById('backupBtn');
  if (backupBtn) backupBtn.onclick = async () => { try { const data = await api('/api/system/backup', {method:'POST', body:{}}); toast(`备份完成：${data.filePath}`); } catch(err) { toast(err.message); } };
}
