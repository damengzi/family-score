function renderProfile() {
  const profile = state.profile || {};
  const user = profile.user || {};
  const acc = state.dashboard?.account || {};
  const tasks = state.dashboard?.tasks || [];
  const pendingOrders = state.exchangeOrders.filter(o => o.status === 'PENDING' && (!state.childId || Number(o.childId) === Number(state.childId))).length;
  return `<div class="stack">
    <div class="card overview-hero-card profile-hero">
      <div><div class="eyebrow">${h(profileText('个人主页','Profile'))}</div><h2>${h(profileText(`${user.name || state.me?.name || '我的账号'}，欢迎回来`, `Welcome, ${user.name || state.me?.name || 'User'}`))}</h2><p>${h(profileText('这里可以查看账号信息、设置主题语言、修改密码，并快速了解自己的家庭积分使用状态。','View account information, preferences, password security, and your family score activity.'))}</p></div>
      <div class="overview-actions"><button class="secondary" data-tab="overview">${tr('overview','今日概览')}</button><button class="secondary" data-tab="growthReport">${tr('growthReport','成长报告')}</button></div>
    </div>

    <div class="mini-stat-grid">
      ${miniStat(profileText('当前角色','Role'), roleName(user.role || state.me?.role), profileText('账号权限边界','Permission scope'), '🪪')}
      ${miniStat(profileText('可见孩子','Children'), profile.accessibleChilds ?? state.children.length, profileText('当前账号可访问','Accessible profiles'), '👧')}
      ${miniStat(profileText('愿望单','Wishlist'), wishlistCount(state.childId), profileText('当前孩子收藏','Saved rewards'), '🌟')}
      ${miniStat(profileText('待处理','Pending'), pendingOrders + tasks.filter(t => t.status === 'SUBMITTED').length, profileText('任务/兑换提醒','Tasks / rewards'), '⏳')}
    </div>

    <div class="split">
      <div class="card"><div class="section-title"><div><h2>${h(profileText('基本信息','Basic info'))}</h2><p class="small">${h(profileText('账号资料由管理员或家长创建，个人主页只展示当前登录用户。','This page only shows the current signed-in user.'))}</p></div><span class="tag">${h(roleName(user.role || state.me?.role))}</span></div>${profileInfo(user, profile)}</div>
      <div class="card"><div class="section-title"><div><h2>${h(profileText('偏好设置','Preferences'))}</h2><p class="small">${h(profileText('主题和语言保存在当前浏览器，不影响其他设备。','Theme and language are saved in this browser only.'))}</p></div><span class="tag">localStorage</span></div>${profilePrefsForm()}</div>
    </div>

    <div class="split">
      <div class="card"><div class="section-title"><div><h2>${h(profileText('修改密码','Change password'))}</h2><p class="small">${h(profileText('需要输入旧密码；修改成功后请使用新密码登录。','Old password is required. Use the new password next time.'))}</p></div><span class="tag red">${h(profileText('安全','Security'))}</span></div>${passwordForm()}</div>
      <div class="card"><div class="section-title"><div><h2>${h(profileText('我的使用小结','Activity summary'))}</h2><p class="small">${h(profileText('结合当前孩子和账号可见数据生成。','Generated from currently visible data.'))}</p></div><span class="tag ${statusClass(acc.statusLevel)}">${statusText(acc.statusLevel)}</span></div>${profileSummary(acc, tasks)}</div>
    </div>

    <div class="grid two">
      <div class="card"><h2>${h(profileText('安全建议','Security tips'))}</h2><div class="practice-list"><div><b>${h(profileText('定期更换密码','Rotate passwords'))}</b><p>${h(profileText('家长和管理员账号建议使用不容易被孩子猜到的密码。','Parent and admin accounts should use passwords children cannot easily guess.'))}</p></div><div><b>${h(profileText('退出公共设备','Sign out on shared devices'))}</b><p>${h(profileText('本系统是本机部署，若多人共用电脑，操作后建议退出登录。','If sharing this computer, sign out after sensitive actions.'))}</p></div><div><b>${h(profileText('先备份再大改','Backup before big changes'))}</b><p>${h(profileText('批量调整孩子、任务、奖励前，管理员可先创建本机备份。','Admins can create a local backup before major changes.'))}</p></div></div></div>
      <div class="card"><h2>${h(profileText('我的家庭身份卡','Family identity card'))}</h2><div class="speech-grid profile-badges"><div class="speech-card"><span>${h(profileText('身份','Role'))}</span><p>${h(profileIdentity())}</p></div><div class="speech-card"><span>${h(profileText('当前关注','Focus'))}</span><p>${h(profileFocus())}</p></div></div></div>
    </div>
  </div>`;
}

function profileText(zh, en) {
  return langKey() === 'en' ? en : zh;
}

function profileInfo(user, profile) {
  const rows = [
    [profileText('显示名称','Display name'), user.name || state.me?.name || '-'],
    [profileText('登录名','Login name'), user.loginName || '-'],
    [profileText('角色','Role'), roleName(user.role || state.me?.role)],
    [profileText('绑定孩子','Bound child'), profile.boundChildName || (user.childId ? `ID ${user.childId}` : profileText('未绑定','Not bound'))],
    [profileText('创建时间','Created at'), user.createdAt || '-'],
    [profileText('登录有效期','Session expires'), profile.sessionExpiresAt || '-'],
  ];
  return `<div class="report-list">${rows.map(r => `<div class="report-row"><div><b>${h(r[0])}</b></div><strong>${h(r[1])}</strong></div>`).join('')}</div>`;
}

function profilePrefsForm() {
  return `<form class="form" id="profilePrefsForm">
    <div class="field"><label>${h(profileText('主题','Theme'))}</label><select name="theme"><option value="system" ${appPrefs.theme === 'system' ? 'selected' : ''}>${h(profileText('以系统为准','Follow system'))}</option><option value="light" ${appPrefs.theme === 'light' ? 'selected' : ''}>${h(profileText('浅色','Light'))}</option><option value="dark" ${appPrefs.theme === 'dark' ? 'selected' : ''}>${h(profileText('深色','Dark'))}</option></select></div>
    <div class="field"><label>${h(profileText('语言','Language'))}</label><select name="language"><option value="zh" ${appPrefs.language !== 'en' ? 'selected' : ''}>简体中文</option><option value="en" ${appPrefs.language === 'en' ? 'selected' : ''}>English</option></select><div class="small">${h(profileText('当前版本优先覆盖个人主页和导航，业务表单文案后续可继续国际化。','This version covers profile and navigation first; business copy can be expanded later.'))}</div></div>
    <button>${h(profileText('保存偏好','Save preferences'))}</button>
  </form><div class="theme-preview"><div></div><div></div><div></div></div>`;
}

function passwordForm() {
  return `<form class="form" id="profilePasswordForm">
    <div class="field"><label>${h(profileText('旧密码','Old password'))}</label><input name="oldPassword" type="password" required></div>
    <div class="field"><label>${h(profileText('新密码','New password'))}</label><input name="newPassword" type="password" minlength="4" required></div>
    <div class="field"><label>${h(profileText('确认新密码','Confirm new password'))}</label><input name="confirmPassword" type="password" minlength="4" required></div>
    <button>${h(profileText('修改密码','Change password'))}</button>
  </form>`;
}

function profileSummary(acc, tasks) {
  return `<div class="report-list">
    ${reportRow(profileText('基准德育分','Base score'), acc.baseScore ?? 100, profileText('当前孩子状态账户','Current child status'))}
    ${reportRow(profileText('兑换分 / 星星','Bonus / stars'), `${acc.bonusScore ?? 0} / ${acc.starCount ?? 0}`, profileText('奖励兑换资源','Reward resources'))}
    ${reportRow(profileText('今日任务','Today tasks'), `${tasks.filter(t => t.status === 'APPROVED').length}/${tasks.length}`, profileText('已点亮 / 总任务','Approved / total'))}
    ${reportRow(profileText('家庭小队分','Team score'), acc.teamScore ?? 0, profileText('家庭协作能量','Family teamwork'))}
  </div>`;
}

function profileIdentity() {
  if (state.me?.role === 'ADMIN') return profileText('你是家庭系统管理员，负责账号、备份、规则和全局配置。','You manage accounts, backups, rules, and global settings.');
  if (state.me?.role === 'PARENT') return profileText('你是陪伴者和审核者，重点是稳定反馈、具体鼓励和必要修复。','You review, encourage, and guide repair with steady feedback.');
  return profileText('你是成长主角，可以完成任务、查看分值、申请奖励和管理愿望单。','You complete tasks, view scores, request rewards, and manage a wishlist.');
}

function profileFocus() {
  if (state.me?.role === 'ADMIN') return profileText('建议每周检查一次备份和用户状态。','Review backups and user status weekly.');
  if (state.me?.role === 'PARENT') return profileText('建议每天先处理待审核中心，再记录关键行为。','Review pending items first, then record key behaviors.');
  return profileText('建议先完成今日任务，再看看愿望单还差多少。','Complete today’s tasks, then check wishlist progress.');
}
