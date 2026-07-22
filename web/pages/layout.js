function renderApp() {
  applyRoleTheme();
  if (!canOperate() && ['score', 'childConfig', 'taskConfig', 'rewardConfig', 'auditCenter'].includes(state.tab)) state.tab = 'overview';
  if (!isAdmin() && ['users', 'system'].includes(state.tab)) state.tab = 'overview';
  const acc = state.dashboard?.account || {};
  const childOptions = state.children.map(c => `<option value="${c.id}" ${c.id === state.childId ? 'selected' : ''}>${h(c.name)} ${c.age}${tr('age', '岁')}</option>`).join('');
  const operateTabs = canOperate() ? `${tabBtn('childConfig', tr('childConfig', '孩子管理'))}${tabBtn('taskConfig', tr('taskConfig', '任务自定义'))}${tabBtn('rewardConfig', tr('rewardConfig', '奖励/零食自定义'))}` : '';
  const adminTabs = isAdmin() ? `${tabBtn('users', tr('users', '用户管理'))}${tabBtn('familyGroups', '家庭组')}${tabBtn('system', tr('system', '本机备份'))}` : '';
  app.innerHTML = `<div class="layout">
    <div class="hero">
      <div><div class="eyebrow">${h(heroEyebrow())}</div><h1>${h(heroTitle())}</h1><p>${h(heroSubtitle())}</p><span class="role-note">${h(roleIntro())}</span></div>
      <div class="hero-panel"><div class="row"><span class="tag ${isAdmin() ? 'green' : 'blue'}">${roleName(state.me?.role)}</span><button class="profile-entry" data-tab="profile">${tr('profile', '个人主页')} · ${tr('currentUser', '当前用户')}${tr('colon', '：')}${h(displayName(state.me?.name))}</button></div>${renderRealtimeClock()}<div class="row"><select id="childSelect">${childOptions}</select><button class="secondary" id="logoutBtn">${tr('logout', '退出')}</button></div></div>
    </div>
    <div class="grid">
      ${metric(tr('metricBase', '基准德育分'), acc.baseScore ?? 100, `<span class="tag ${statusClass(acc.statusLevel)}">${statusText(acc.statusLevel)}</span>`, '🌱')}
      ${metric(tr('metricBonus', '超额兑换分'), acc.bonusScore ?? 0, `<span class="small">${tr('metricBonusHint', '用于兑换零食/奖励')}</span>`, '🪙')}
      ${metric(tr('metricStars', '星星'), acc.starCount ?? 0, `<span class="small">${tr('metricStarsHint', '长期大奖')}</span>`, '⭐')}
      ${metric(tr('metricTeam', '家庭小队分'), acc.teamScore ?? 0, `<span class="small">${tr('metricTeamHint', '月度评级')}</span>`, '🏠')}
    </div>
    <div class="tabs">
      ${tabBtn('overview', tr('overview', '今日概览'))}${canOperate() ? tabBtn('auditCenter', tr('auditCenter', '待审核中心')) : ''}${tabBtn('growthReport', tr('growthReport', '成长报告'))}${tabBtn('guide', tr('guide', '分值说明'))}${tabBtn('detail', tr('detail', '积分明细'))}${canOperate() ? tabBtn('score', tr('score', '加分/扣分/惩罚')) : ''}${tabBtn('tasks', tr('tasks', '今日任务'))}${tabBtn('rewards', tr('rewards', '奖励兑换'))}${operateTabs}${adminTabs}
    </div>
    ${localizeHtml(renderTab())}
  </div>`;
  document.getElementById('childSelect').onchange = async (e) => { state.childId = Number(e.target.value); await loadAll(); };
  document.getElementById('logoutBtn').onclick = async () => { await api('/api/auth/logout', { method: 'POST', body: {} }); state.me = null; renderLogin(); };
  bindEvents();
  mountRealtimeClock();
  if (typeof showTaskDeadlineReminder === 'function') showTaskDeadlineReminder();
}

function metric(label, value, extra, icon) { return `<div class="card metric"><div class="metric-icon">${icon || '•'}</div><div class="label">${label}</div><div class="value">${h(value)}</div><div>${extra}</div></div>`; }
function tabBtn(key, text) { return `<button class="tab ${state.tab === key ? 'active' : ''}" data-tab="${key}">${text}</button>`; }
function renderTab() {
  if (state.tab === 'overview') return renderOverview();
  if (state.tab === 'profile') return renderProfile();
  if (state.tab === 'auditCenter') return renderAuditCenter();
  if (state.tab === 'growthReport') return renderGrowthReport();
  if (state.tab === 'guide') return renderScoreGuide();
  if (state.tab === 'score') return renderScoreForm();
  if (state.tab === 'tasks') return renderTasks();
  if (state.tab === 'childConfig') return renderChildren();
  if (state.tab === 'familyGroups') return renderFamilyGroups();
  if (state.tab === 'taskConfig') return renderTaskConfig();
  if (state.tab === 'rewards') return renderRewards();
  if (state.tab === 'rewardConfig') return renderRewardConfig();
  if (state.tab === 'users') return renderUsers();
  if (state.tab === 'system') return renderSystem();
  return renderDetail();
}

function heroEyebrow() {
  if (state.me?.role === 'ADMIN') return tr('heroAdminEyebrow', '家庭系统总控台');
  if (state.me?.role === 'PARENT') return tr('heroParentEyebrow', '今日陪伴与审核');
  if (state.me?.role === 'CHILD') return tr('heroChildEyebrow', '我的成长小基地');
  return tr('heroGuestEyebrow', '本机家庭成长空间');
}

function heroTitle() {
  return dailyRoleTip().title;
}

function heroSubtitle() {
  return dailyRoleTip().subtitle;
}

function dailyRoleTip() {
  const role = state.me?.role || 'GUEST';
  const age = Number(selectedChild()?.age || 0);
  const band = age && age <= 7 ? 'young' : age && age >= 11 ? 'older' : 'middle';
  const tips = {
    ADMIN: [
      ['家庭德育积分系统', '今天可以先检查成员绑定、任务配置和本机备份，让家庭规则更稳。'],
      ['把家庭成长数据整理清楚', '清晰的账号、孩子档案和任务科目，会让后续记录更少混乱。'],
      ['让系统成为家庭协作台', '先把结构搭好，再把成长交给每天的小行动。'],
    ],
    PARENT: {
      young: [['多给一句具体鼓励', '低龄孩子更需要看见“我哪里做对了”，今天试着表扬一个具体动作。'], ['规则短一点，拥抱多一点', '把任务拆成一步一步的小挑战，孩子更容易坚持。']],
      middle: [['把规则变成温和、稳定的陪伴', '小学阶段适合用清晰目标和及时反馈，帮助孩子建立节奏感。'], ['今天先抓一个小进步', '与其追求全做对，不如确认一个可重复的好习惯。']],
      older: [['尊重感会让规则更有力量', '高年级孩子需要参与制定目标，今天可以一起复盘一次选择。'], ['少一点催促，多一点共识', '把任务、科目和奖励讲清楚，让孩子参与自己的成长计划。']],
    },
    CHILD: {
      young: [['今天也要点亮一个小进步', '从整理书包、认真阅读或完成一道题开始，小小行动也会发光。'], ['先完成一个容易的小目标', '做完记得提交，家长确认后就能看到自己的成长。']],
      middle: [['今天挑战一个更专注的自己', '阅读、数学或英语任务任选一个，完成后写下你的收获。'], ['把努力变成看得见的星星', '先完成任务，再慢慢靠近愿望单里的奖励。']],
      older: [['为自己的计划负责', '选择一个科目任务认真完成，提交时写清楚过程和答案。'], ['稳定坚持比一次冲刺更厉害', '今天认真完成一个任务，就是给明天的自己铺路。']],
    },
    GUEST: [['本机家庭成长空间', '适合家庭本机部署的德育积分和成长记录系统。']],
  };
  const pool = Array.isArray(tips[role]) ? tips[role] : (tips[role]?.[band] || tips.GUEST);
  const index = Math.abs(hashText(`${role}-${band}-${todayKey()}-${state.childId || 0}`)) % pool.length;
  return { title: pool[index][0], subtitle: pool[index][1] };
}

function selectedChild() {
  return state.children.find(c => Number(c.id) === Number(state.childId)) || null;
}

function hashText(text) {
  return String(text).split('').reduce((sum, ch) => ((sum * 31) + ch.charCodeAt(0)) | 0, 7);
}

function bindEvents() {
  document.querySelectorAll('[data-tab]').forEach(b => b.onclick = async () => { state.tab = b.dataset.tab; renderApp(); });
  document.querySelectorAll('[data-score-type]').forEach(b => b.onclick = () => fillScorePreset(b));
  document.querySelectorAll('[data-task-name]').forEach(b => b.onclick = () => fillTaskPreset(b));
  document.querySelectorAll('[data-reward-name]').forEach(b => b.onclick = () => fillRewardPreset(b));
  document.querySelectorAll('[data-add-question-row]').forEach(b => b.onclick = () => addQuestionRow(b.dataset.addQuestionRow));
  if (typeof bindQuestionRowEvents === 'function') bindQuestionRowEvents();
  const profilePrefsForm = document.getElementById('profilePrefsForm');
  if (profilePrefsForm) profilePrefsForm.onsubmit = (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); savePrefs({ theme: body.theme || 'system', language: body.language || 'zh' }); toast(tr('toastPrefsSaved', '偏好已保存')); renderApp(); };
  const profilePasswordForm = document.getElementById('profilePasswordForm');
  if (profilePasswordForm) profilePasswordForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); if (body.newPassword !== body.confirmPassword) { toast(tr('toastPasswordMismatch', '两次输入的新密码不一致')); return; } try { await api('/api/profile/password', { method: 'POST', body: { oldPassword: body.oldPassword, newPassword: body.newPassword } }); e.target.reset(); toast(tr('toastPasswordChanged', '密码已修改，请牢记新密码')); } catch (err) { toast(err.message); } };
  const scoreForm = document.getElementById('scoreForm');
  if (scoreForm) scoreForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); body.childId = state.childId; body.scoreChange = Number(body.scoreChange); body.targetAccount = body.recordType === 'TEAM' ? 'TEAM' : body.recordType === 'STAR' ? 'STAR' : 'AUTO'; try { await api('/api/score-records', { method: 'POST', body }); toast(tr('toastRecordSaved', '记录成功')); await loadAll(); } catch (err) { toast(err.message); } };
  const childForm = document.getElementById('childForm');
  if (childForm) childForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); body.age = Number(body.age); body.parentUserId = Number(body.parentUserId || 0); try { await api('/api/children', { method: 'POST', body }); toast(tr('toastChildAdded', '孩子已新增')); await loadHome(); } catch (err) { toast(err.message); } };
  const guardianGroupForm = document.getElementById('guardianGroupForm');
  if (guardianGroupForm) guardianGroupForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); try { await api('/api/guardian-groups', { method: 'POST', body }); toast('监护组已新增'); await loadAll(); } catch (err) { toast(err.message); } };
  document.querySelectorAll('[data-save-group]').forEach(b => b.onclick = async () => { const id = Number(b.dataset.saveGroup); const body = { name: document.querySelector(`[data-group-name="${id}"]`)?.value || '', description: document.querySelector(`[data-group-desc="${id}"]`)?.value || '' }; try { await api(`/api/guardian-groups/${id}`, { method: 'PATCH', body }); toast('监护组已保存'); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-del-group]').forEach(b => b.onclick = async () => { if (!confirm('确认删除该监护组？仍有家长或孩子绑定时不能删除。')) return; try { await api(`/api/guardian-groups/${b.dataset.delGroup}`, { method: 'DELETE' }); toast(tr('toastDeleted', '已删除')); await loadAll(); } catch (err) { toast(err.message); } });
  const familyGroupForm = document.getElementById('familyGroupForm');
  if (familyGroupForm) familyGroupForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); try { await api('/api/guardian-groups', { method: 'POST', body }); toast('家庭组已新增'); await loadAll(); } catch (err) { toast(err.message); } };
  document.querySelectorAll('[data-select-family-group]').forEach(b => b.onclick = () => { state.familyGroupName = b.dataset.selectFamilyGroup; renderApp(); });
  document.querySelectorAll('[data-save-family-group]').forEach(b => b.onclick = async () => { const id = Number(b.dataset.saveFamilyGroup); const body = { name: document.querySelector(`[data-family-group-name="${id}"]`)?.value || '', description: document.querySelector(`[data-family-group-desc="${id}"]`)?.value || '' }; try { const old = currentFamilyGroup()?.name; await api(`/api/guardian-groups/${id}`, { method: 'PATCH', body }); if (old === state.familyGroupName) state.familyGroupName = body.name; toast('家庭组已保存'); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-del-family-group]').forEach(b => b.onclick = async () => { if (!confirm('确认删除该家庭组？仍有成员时不能删除。')) return; try { await api(`/api/guardian-groups/${b.dataset.delFamilyGroup}`, { method: 'DELETE' }); state.familyGroupName = ''; toast(tr('toastDeleted', '已删除')); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-add-parent-group]').forEach(b => b.onclick = async () => { try { await setParentFamilyGroup(Number(b.dataset.addParentGroup), b.dataset.groupName || ''); toast('已添加家长成员'); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-remove-parent-group]').forEach(b => b.onclick = async () => { try { await setParentFamilyGroup(Number(b.dataset.removeParentGroup), ''); toast('已移出家长成员'); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-add-child-group]').forEach(b => b.onclick = async () => { try { await setChildFamilyGroup(Number(b.dataset.addChildGroup), b.dataset.groupName || ''); toast('已添加孩子成员'); await loadHome(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-remove-child-group]').forEach(b => b.onclick = async () => { try { await setChildFamilyGroup(Number(b.dataset.removeChildGroup), ''); toast('已移出孩子成员'); await loadHome(); } catch (err) { toast(err.message); } });
  const publishTaskForm = document.getElementById('publishTaskForm');
  if (publishTaskForm) publishTaskForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); body.childId = Number(body.childId || state.childId || 0); body.scoreValue = Number(body.scoreValue || 1); applyTaskQuestionsToBody(e.target, body); try { await api('/api/tasks/publish', { method: 'POST', body }); toast('任务已发布'); await loadAll(); } catch (err) { toast(err.message); } };
  const taskTemplateForm = document.getElementById('taskTemplateForm');
  if (taskTemplateForm) taskTemplateForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); body.scoreValue = Number(body.scoreValue); applyTaskQuestionsToBody(e.target, body); try { await api('/api/task-templates', { method: 'POST', body }); toast(tr('toastTaskTemplateAdded', '任务模板已新增')); await loadAll(); } catch (err) { toast(err.message); } };
  const rewardForm = document.getElementById('rewardForm');
  if (rewardForm) rewardForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target));['costScore', 'costStar', 'weeklyLimit', 'monthlyLimit'].forEach(k => body[k] = Number(body[k])); try { await api('/api/rewards', { method: 'POST', body }); toast(tr('toastRewardAdded', '奖励已新增')); await loadAll(); } catch (err) { toast(err.message); } };
  const wishForm = document.getElementById('wishForm');
  if (wishForm) wishForm.onsubmit = async (e) => { e.preventDefault(); if (!state.childId) { toast(tr('toastChooseChild', '请先选择孩子')); return; } const body = Object.fromEntries(new FormData(e.target)); body.childId = state.childId; body.expectedScore = Number(body.expectedScore || 0); body.expectedStar = Number(body.expectedStar || 0); try { await api('/api/wishes', { method: 'POST', body }); toast('愿望已提交，等待家长审批'); e.target.reset(); await loadAll(); } catch (err) { toast(err.message); } };
  const userRole = document.getElementById('userRole');
  if (userRole) userRole.onchange = () => { const childField = document.getElementById('bindChildField'); const parentField = document.getElementById('parentBindField'); if (childField) childField.style.display = userRole.value === 'CHILD' ? 'block' : 'none'; if (parentField) parentField.style.display = userRole.value === 'PARENT' ? 'block' : 'none'; };
  const userForm = document.getElementById('userForm');
  if (userForm) userForm.onsubmit = async (e) => { e.preventDefault(); const body = Object.fromEntries(new FormData(e.target)); body.childId = Number(body.childId || 0); try { await api('/api/users', { method: 'POST', body }); toast(tr('toastUserAdded', '用户已注册')); await loadAll(); } catch (err) { toast(err.message); } };
  document.querySelectorAll('[data-save-user]').forEach(b => b.onclick = async () => { const id = Number(b.dataset.saveUser); const body = { displayName: document.querySelector(`[data-user-name="${id}"]`)?.value || '', childId: Number(document.querySelector(`[data-user-child="${id}"]`)?.value || 0), parentTitle: document.querySelector(`[data-user-parent-title="${id}"]`)?.value || '', parentGroup: document.querySelector(`[data-user-parent-group="${id}"]`)?.value || '', password: document.querySelector(`[data-user-pass="${id}"]`)?.value || '' }; try { const data = await api(`/api/users/${id}`, { method: 'PATCH', body }); if (data.user?.id === state.me.userId) state.me.name = data.user.name; toast(tr('toastUserSaved', '用户已保存')); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-save-child]').forEach(b => b.onclick = async () => { const id = Number(b.dataset.saveChild); const body = { name: document.querySelector(`[data-child-name="${id}"]`)?.value || '', age: Number(document.querySelector(`[data-child-age="${id}"]`)?.value || 0), gender: document.querySelector(`[data-child-gender="${id}"]`)?.value || 'BOY', parentUserId: Number(document.querySelector(`[data-child-parent="${id}"]`)?.value || 0), parentGroup: document.querySelector(`[data-child-parent-group="${id}"]`)?.value || '' }; try { await api(`/api/children/${id}`, { method: 'PATCH', body }); toast(tr('toastChildSaved', '孩子档案已保存')); await loadHome(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-del-child]').forEach(b => b.onclick = async () => { if (!confirm(tr('confirmDeleteChild', '确认删除该孩子？相关积分、任务、兑换记录也会删除，绑定的孩子账号会被注销。'))) return; try { await api(`/api/children/${b.dataset.delChild}`, { method: 'DELETE' }); toast(tr('toastDeleted', '已删除')); if (state.childId === Number(b.dataset.delChild)) state.childId = null; await loadHome(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-del-task-template]').forEach(b => b.onclick = async () => { if (!confirm(tr('confirmDeleteTaskTemplate', '确认删除该任务模板？'))) return; try { await api(`/api/task-templates/${b.dataset.delTaskTemplate}`, { method: 'DELETE' }); toast(tr('toastDeleted', '已删除')); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-del-reward]').forEach(b => b.onclick = async () => { if (!confirm(tr('confirmDeleteReward', '确认删除该奖励？'))) return; try { await api(`/api/rewards/${b.dataset.delReward}`, { method: 'DELETE' }); toast(tr('toastDeleted', '已删除')); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-del-user]').forEach(b => b.onclick = async () => { if (!confirm(tr('confirmDeleteUser', '确认注销该用户？注销后该账号不能登录。'))) return; try { await api(`/api/users/${b.dataset.delUser}`, { method: 'DELETE' }); toast(tr('toastDeleted', '已删除')); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-submit-task]').forEach(b => b.onclick = async () => { const note = prompt('完成说明（可简单写：已完成）', '已完成') || '已完成'; try { await api(`/api/tasks/${b.dataset.submitTask}/submit`, { method: 'POST', body: { submitNote: note } }); toast(tr('toastTaskSubmitted', '已提交，等待家长确认')); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-appeal-task]').forEach(b => b.onclick = async () => { if (!state.childId) { toast(tr('toastChooseChild', '请先选择孩子')); return; } const reason = prompt('请写下任务不合理或无法完成的原因'); if (!reason) return; const expectedSolution = prompt('你希望家长怎么调整？', '请家长帮我调整任务或重新说明规则') || ''; try { await api('/api/appeals', { method: 'POST', body: { childId: state.childId, targetType: 'TASK', targetId: Number(b.dataset.appealTask), appealReason: reason, expectedSolution } }); toast('申诉已提交，等待家长处理'); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-audit-task]').forEach(b => b.onclick = async () => { try { await api(`/api/tasks/${b.dataset.auditTask}/audit`, { method: 'POST', body: { result: 'APPROVED', auditNote: '通过' } }); toast(tr('toastTaskApproved', '任务已通过')); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-reject-task]').forEach(b => b.onclick = async () => { try { await api(`/api/tasks/${b.dataset.rejectTask}/audit`, { method: 'POST', body: { result: 'REJECTED', auditNote: '暂不通过' } }); toast(tr('toastTaskRejected', '任务已驳回')); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-reward]').forEach(b => b.onclick = async () => { if (!state.childId) { toast(tr('toastChooseChild', '请先选择孩子')); return; } try { await api('/api/exchange-orders', { method: 'POST', body: { childId: state.childId, rewardId: Number(b.dataset.reward), note: '申请兑换' } }); toast(tr('toastExchangeSubmitted', '已提交兑换申请，需家长审核')); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-wishlist]').forEach(b => b.onclick = async () => { if (!state.childId) { toast(tr('toastChooseChild', '请先选择孩子')); return; } const reward = state.rewards.find(r => Number(r.id) === Number(b.dataset.wishlist)); if (!reward) return; if (isWishlisted(state.childId, reward.id)) { toast('该愿望已提交，等待家长查看或审批'); return; } try { await api('/api/wishes', { method: 'POST', body: { childId: state.childId, wishName: reward.rewardName, wishType: reward.rewardType, expectedScore: Number(reward.costScore || 0), expectedStar: Number(reward.costStar || 0), reason: `奖励ID:${reward.id}。我想把这个加入愿望。` } }); toast('愿望已提交，等待家长审批'); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-audit-wish]').forEach(b => b.onclick = async () => { const note = prompt('通过意见', '可以作为阶段目标，我们一起努力') || '通过'; try { await api(`/api/wishes/${b.dataset.auditWish}/audit`, { method: 'POST', body: { result: 'APPROVED', auditNote: note } }); toast('愿望已通过'); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-reject-wish]').forEach(b => b.onclick = async () => { const note = prompt('暂不通过原因', '我们可以换一个更合适的目标') || '暂不通过'; try { await api(`/api/wishes/${b.dataset.rejectWish}/audit`, { method: 'POST', body: { result: 'REJECTED', auditNote: note } }); toast('愿望已驳回'); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-handle-appeal]').forEach(b => b.onclick = async () => { const note = prompt('处理意见', '申诉通过，任务重新打开或规则调整') || '申诉通过'; try { await api(`/api/appeals/${b.dataset.handleAppeal}/handle`, { method: 'POST', body: { result: 'APPROVED', handleNote: note } }); toast('申诉已处理'); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-reject-appeal]').forEach(b => b.onclick = async () => { const note = prompt('驳回原因', '规则暂不调整，请继续尝试') || '申诉驳回'; try { await api(`/api/appeals/${b.dataset.rejectAppeal}/handle`, { method: 'POST', body: { result: 'REJECTED', handleNote: note } }); toast('申诉已驳回'); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-audit-order]').forEach(b => b.onclick = async () => { try { await api(`/api/exchange-orders/${b.dataset.auditOrder}/audit`, { method: 'POST', body: { result: 'APPROVED', auditNote: '通过' } }); toast(tr('toastExchangeApproved', '兑换已通过')); await loadAll(); } catch (err) { toast(err.message); } });
  document.querySelectorAll('[data-reject-order]').forEach(b => b.onclick = async () => { try { await api(`/api/exchange-orders/${b.dataset.rejectOrder}/audit`, { method: 'POST', body: { result: 'REJECTED', auditNote: '驳回' } }); toast(tr('toastExchangeRejected', '兑换已驳回')); await loadAll(); } catch (err) { toast(err.message); } });
  const backupBtn = document.getElementById('backupBtn');
  if (backupBtn) backupBtn.onclick = async () => { try { const data = await api('/api/system/backup', { method: 'POST', body: {} }); toast(`备份完成：${data.filePath}`); } catch (err) { toast(err.message); } };
  const networkInfoBox = document.getElementById('networkInfoBox');
  if (networkInfoBox) loadNetworkInfo(networkInfoBox);
}

function fillScorePreset(btn) {
  const form = document.getElementById('scoreForm');
  if (!form) return;
  form.elements.recordType.value = btn.dataset.scoreType || 'ADD';
  form.elements.itemName.value = btn.dataset.scoreItem || '';
  form.elements.scoreChange.value = btn.dataset.scoreValue || 1;
  form.elements.reason.value = btn.dataset.scoreReason || '';
  toast(tr('toastScorePreset', '已填入记录模板，可继续修改'));
}

function fillTaskPreset(btn) {
  const form = document.getElementById('taskTemplateForm');
  if (!form) return;
  form.elements.taskName.value = btn.dataset.taskName || '';
  form.elements.taskType.value = btn.dataset.taskType || 'DAILY';
  form.elements.category.value = btn.dataset.taskCategory || 'ACTION';
  form.elements.subject.value = btn.dataset.taskSubject || 'GENERAL';
  form.elements.questionType.value = btn.dataset.taskQuestionType || 'NONE';
  form.elements.content.value = btn.dataset.taskContent || '';
  form.elements.answer.value = btn.dataset.taskAnswer || '';
  form.elements.scoreValue.value = btn.dataset.taskValue || 1;
  form.elements.targetAccount.value = btn.dataset.taskTarget || 'AUTO';
  form.elements.description.value = btn.dataset.taskDescription || '';
  toast(tr('toastTaskPreset', '已填入任务模板，可继续修改'));
}

function fillRewardPreset(btn) {
  const form = document.getElementById('rewardForm');
  if (!form) return;
  form.elements.rewardName.value = btn.dataset.rewardName || '';
  form.elements.rewardType.value = btn.dataset.rewardType || 'SNACK';
  form.elements.healthRisk.value = btn.dataset.rewardRisk || 'NONE';
  form.elements.costScore.value = btn.dataset.rewardScore || 0;
  form.elements.costStar.value = btn.dataset.rewardStar || 0;
  form.elements.weeklyLimit.value = btn.dataset.rewardWeekly || 1;
  form.elements.monthlyLimit.value = btn.dataset.rewardMonthly || 1;
  form.elements.description.value = btn.dataset.rewardDescription || '';
  toast(tr('toastRewardPreset', '已填入奖励配置，可继续修改'));
}

