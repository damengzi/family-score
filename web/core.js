const app = document.getElementById('app');
let state = { status: null, me: null, profile: null, children: [], childId: null, dashboard: null, childDashboards: {}, records: [], rewards: [], taskTemplates: [], exchangeOrders: [], wishes: [], appeals: [], users: [], guardianGroups: [], tab: 'overview' };
let appPrefs = loadPrefs();
let globalClockTimer = null;
let globalClockMode = 'digital';

const api = async (path, options = {}) => {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), options.timeout || 15000);
  try {
    const res = await fetch(path, {
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
      ...options,
      signal: controller.signal,
      body: options.body ? JSON.stringify(options.body) : undefined,
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.error || data.message || '请求失败');
    return data;
  } catch (err) {
    if (err.name === 'AbortError') throw new Error('请求超时，请确认服务是否正常运行');
    throw err;
  } finally {
    clearTimeout(timer);
  }
};

const h = (s) => String(s ?? '').replace(/[&<>"]/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]));
const toast = (msg) => {
  const el = document.createElement('div');
  el.className = 'toast';
  el.textContent = msg;
  document.body.appendChild(el);
  setTimeout(() => el.remove(), 3200);
};
const setError = (id, msg) => { const el = document.getElementById(id); if (el) el.textContent = msg || ''; };
const statusText = (s) => tr(`status.${s || 'GREEN'}`, ({GREEN:'绿色稳定',BLUE:'蓝色提醒',YELLOW:'黄色关注',ORANGE:'橙色预警',RED:'红色重点帮助',DEEP_REPAIR:'深度修复'}[s] || s || '绿色稳定'));
const statusClass = (s) => ({GREEN:'green',BLUE:'blue',YELLOW:'yellow',ORANGE:'orange',RED:'red',DEEP_REPAIR:'red'}[s] || 'green');
const recordTypeName = (s) => tr(`record.${s}`, ({ADD:'加分',DEDUCT:'扣分',REPAIR:'惩罚/修复',TEAM:'家庭小队分',STAR:'星星',EXCHANGE:'兑换'}[s] || s));
const isAdmin = () => state.me && state.me.role === 'ADMIN';
const canOperate = () => state.me && (state.me.role === 'ADMIN' || state.me.role === 'PARENT');
const roleName = (s) => tr(`role.${s}`, ({ADMIN:'管理员',PARENT:'家长',CHILD:'孩子'}[s] || s));
const roleKey = () => ({ADMIN:'admin',PARENT:'parent',CHILD:'child'}[state.me?.role] || 'guest');
const applyRoleTheme = () => { document.body.dataset.role = roleKey(); applyPreferences(); };
const roleIntro = () => tr(`roleIntro.${state.me?.role || 'guest'}`, ({ADMIN:'管理员模式：统筹家庭成员、账号和系统配置',PARENT:'家长模式：关注陪伴、审核和正向反馈',CHILD:'孩子模式：完成任务、积攒星星和兑换奖励'}[state.me?.role] || '家庭德育积分系统'));
const i18n = {
  zh: {profile:'个人主页',overview:'今日概览',auditCenter:'待审核中心',growthReport:'成长报告',guide:'分值说明',detail:'积分明细',score:'加分/扣分/惩罚',tasks:'今日任务',rewards:'奖励兑换'},
  en: {
    appTitle:'Family Moral Score', profile:'Profile', overview:'Overview', auditCenter:'Audit Center', growthReport:'Growth Report', guide:'Score Guide', detail:'Score Records', score:'Score', tasks:'Today Tasks', rewards:'Rewards', clockNow:'Current time', clockDate:'Date', clockToggleHint:'Click for clock face', clockToggleHintBack:'Click for digital time',
    childConfig:'Children', taskConfig:'Task Settings', rewardConfig:'Reward Settings', users:'Users', system:'Local Backup', logout:'Log out', currentUser:'Current user', age:' yrs', colon:': ',
    heroAdminEyebrow:'Family System Console', heroParentEyebrow:'Today’s Guidance & Review', heroChildEyebrow:'My Growth Base', heroGuestEyebrow:'Local Family Growth Space',
    heroAdminTitle:'Family Moral Score', heroParentTitle:'Turn rules into warm, steady guidance', heroChildTitle:'Light up one small progress today',
    heroAdminSubtitle:'Score records · Score changes · Repair tasks · Task settings · Reward exchange · Local backup', heroParentSubtitle:'Track status, review tasks, and record behaviors with concrete feedback.', heroChildSubtitle:'Complete tasks, collect bonus points and stars, and unlock rewards you like.', heroGuestSubtitle:'A local family score and growth tracking system.',
    metricBase:'Base moral score', metricBonus:'Bonus points', metricStars:'Stars', metricTeam:'Family team points', metricBonusHint:'For rewards and treats', metricStarsHint:'Long-term big goals', metricTeamHint:'Monthly family rating',
    'role.ADMIN':'Admin','role.PARENT':'Parent','role.CHILD':'Child','roleIntro.ADMIN':'Admin mode: manage family members, accounts, and system settings','roleIntro.PARENT':'Parent mode: focus on guidance, review, and positive feedback','roleIntro.CHILD':'Child mode: complete tasks, collect stars, and redeem rewards','roleIntro.guest':'Family Moral Score',
    'status.GREEN':'Stable green','status.BLUE':'Blue reminder','status.YELLOW':'Yellow attention','status.ORANGE':'Orange warning','status.RED':'Focused support','status.DEEP_REPAIR':'Deep repair',
    'record.ADD':'Add','record.DEDUCT':'Deduct','record.REPAIR':'Repair','record.TEAM':'Family team','record.STAR':'Star','record.EXCHANGE':'Exchange',
    'taskStatus.TODO':'To do','taskStatus.SUBMITTED':'Waiting review','taskStatus.APPROVED':'Approved','taskStatus.REJECTED':'Try again',
    'taskCategory.STUDY':'Study','taskCategory.SELF_CARE':'Self care','taskCategory.HOUSEWORK':'Housework','taskCategory.EMOTION':'Emotion','taskCategory.HEALTH':'Health','taskCategory.SAFETY':'Safety','taskCategory.GROWTH':'Growth',
    'rewardType.SNACK':'Snack','rewardType.DRINK':'Drink','rewardType.BOOK':'Book','rewardType.TOY':'Toy','rewardType.ACTIVITY':'Activity','rewardType.PRIVILEGE':'Privilege','rewardType.REWARD':'Reward',
    'risk.NONE':'Recommended','risk.LOW':'Low risk','risk.MEDIUM':'Moderate','risk.HIGH':'Caution','risk.UNKNOWN':'Unknown',
    toastPrefsSaved:'Preferences saved', toastPasswordMismatch:'The two new passwords do not match', toastPasswordChanged:'Password changed. Please remember the new password.', toastRecordSaved:'Record saved', toastChildAdded:'Child added', toastTaskTemplateAdded:'Task template added', toastRewardAdded:'Reward added', toastUserAdded:'User registered', toastUserSaved:'User saved', toastChildSaved:'Child profile saved', toastDeleted:'Deleted', toastTaskSubmitted:'Submitted. Waiting for parent review.', toastTaskApproved:'Task approved', toastTaskRejected:'Task rejected', toastExchangeSubmitted:'Exchange request submitted. Parent review required.', toastExchangeApproved:'Exchange approved', toastExchangeRejected:'Exchange rejected', toastChooseChild:'Please choose a child first', toastWishlistAdded:'Added to wishlist', toastWishlistRemoved:'Removed from wishlist', toastScorePreset:'Record preset filled. You can edit it.', toastTaskPreset:'Task preset filled. You can edit it.', toastRewardPreset:'Reward preset filled. You can edit it.',
    confirmDeleteChild:'Delete this child? Related score, task, exchange records and bound child account will also be removed.', confirmDeleteTaskTemplate:'Delete this task template?', confirmDeleteReward:'Delete this reward?', confirmDeleteUser:'Deactivate this user? The account will no longer be able to sign in.',
  },
};
const langKey = () => appPrefs.language === 'en' ? 'en' : 'zh';
const tr = (key, fallback) => i18n[langKey()]?.[key] || fallback || key;
const displayName = (name) => langKey() === 'en' && name === '管理员' ? 'Admin' : (name || '-');

const zhToEn = {
  '启动失败':'Startup failed','初始化家庭德育积分系统':'Initialize Family Moral Score','数据将保存在本机':'Data will be stored locally','完成初始化':'Finish setup','请输入你的登录名和密码。家长和孩子账号由管理员或家长创建。':'Enter your login name and password. Parent and child accounts are created by admins or parents.','登录名':'Login name','密码':'Password','请输入登录名':'Enter login name','请输入密码':'Enter password','登录':'Sign in','登录中...':'Signing in...','忘记密码':'Forgot password','重置密码':'Reset password','新密码':'New password','返回登录':'Back to sign in','图片验证码加载中...':'Loading image captcha...','请选择图片验证码':'Please choose the image captcha',
  '成长时间线':'Growth Timeline','每一次记录，都是一次可回看的家庭反馈。':'Every record is reviewable family feedback.','条记录':'records','无备注':'No note','还没有积分明细':'No score records yet','第一次记录可以从一个主动整理、一次认真阅读或一次修复任务开始。':'The first record can start with organizing, focused reading, or a repair task.',
  '今日任务':'Today Tasks','完成后提交给家长确认，通过后才会加分。':'Submit after completion; points are added after parent review.','共':'Total',' 个':'',' 分':' pts','今天还没有任务':'No tasks today','可以去任务自定义中添加一些适合今天的小目标。':'Add suitable small goals in task settings.','今日实践建议':'Today Practice Ideas','不知道做什么时，可以从这些低门槛任务开始。':'If unsure what to do, start with these low-barrier tasks.','家庭可用':'Family ready','我完成啦':'Done','确认完成':'Approve','再试一次':'Try again','等待家长确认':'Waiting for parent review','已点亮':'Approved','可以重新努力':'Try again','已处理':'Handled','任务自定义添加':'Add Task Template','任务名称':'Task name','任务类型':'Task type','每日':'Daily','临时':'Temporary','分类':'Category','目标账户':'Target account','自动':'Auto','基准分':'Base score','小队分':'Team points','说明':'Description','新增任务模板':'Add task template','任务建议库':'Task idea library','点击后填入上方表单，适合作为家庭常用任务模板。':'Click to fill the form above. Good for common family task templates.','任务模板':'Task templates','暂无模板':'No templates yet','添加一些每日可坚持的小任务。':'Add small daily tasks that can be kept consistently.',
  '奖励商店':'Reward Shop','零食、图书、活动和特权，都可以成为努力后的期待。':'Snacks, books, activities, and privileges can become goals after effort.','个奖励':'rewards','建议优先配置“图书、亲子活动、运动体验、家庭特权”等低风险奖励；零食饮品适合低频、限量、明确规则。':'Prefer low-risk rewards such as books, family activities, sports experiences, and family privileges. Snacks and drinks should be limited and clearly ruled.','奖励货架还是空的':'The reward shelf is empty','可以先配置一个小奖励，比如科普书、亲子活动或小零食。':'Add a small reward first, such as a science book, family activity, or small snack.','我的愿望单':'My Wishlist','先收藏想要的奖励，再通过任务一点点靠近它。':'Save desired rewards first, then get closer through tasks.','个愿望':'wishes','待审核兑换':'Pending Exchanges','确认前可以看看孩子当前分值和奖励健康风险。':'Before approving, review the child score and health risk.','个待处理':'pending','确认兑换':'Approve exchange','暂不兑换':'Not now','暂无待审核兑换':'No pending exchanges','孩子还在积攒兑换能量。':'The child is still collecting reward energy.','健康兑换原则':'Healthy Exchange Principles','让奖励成为期待，而不是新的拉扯。':'Make rewards a positive goal, not a new struggle.','家长参考':'Parent reference','低风险优先':'Low risk first','零食要限量':'Limit snacks','不兑换底线':'Do not exchange boundaries','兑换要复盘':'Review after exchange','申请这个奖励':'Request this reward','加入愿望单':'Add to wishlist','移出愿望单':'Remove from wishlist','想要':'Wanted','兑换分':'bonus pts',' 星':' stars','奖励/零食自定义添加':'Add Reward / Treat','奖励名称':'Reward name','健康风险':'Health risk','消耗兑换分':'Cost bonus points','消耗星星':'Cost stars','周限制':'Weekly limit','月限制':'Monthly limit','新增奖励':'Add reward','奖励配置建议':'Reward ideas','点击后填入上方表单，适合作为家庭奖励货架起点。':'Click to fill the form above. Good as a starting point for the reward shelf.','奖励配置':'Reward settings','暂无奖励':'No rewards yet','添加一个孩子愿意努力争取的小目标。':'Add a small goal the child wants to work for.',
  '加分 / 扣分 / 惩罚修复':'Add / Deduct / Repair','建议先描述事实，再记录分值；分数是反馈工具，不是情绪出口。':'Describe facts before recording points; scores are feedback tools, not emotional outlets.','行为记录助手':'Behavior helper','类型':'Type','加分':'Add','扣分':'Deduct','惩罚/修复加回基准分':'Repair base score','家庭小队分':'Family team points','星星':'Stars','项目名称':'Item name','如：主动整理书包 / 作业拖延 / 整理书桌修复':'e.g. organized schoolbag / homework delay / desk repair','分值':'Points','原因':'Reason','简要说明':'Short note','记录这次成长':'Record this growth','使用提醒':'Tips','扣分只扣基准分；修复任务只恢复基准分；基准分满 100 后，加分进入超额兑换分。':'Deductions affect base score only; repair restores base score only; when base reaches 100, added points go to bonus points.','常用场景模板':'Common Scenarios','点击模板会填入上方表单，提交前仍可修改。':'Click a template to fill the form. You can edit before submitting.','快捷填入':'Quick fill',
  '无权限':'No permission','只有家长或管理员可以添加孩子。':'Only parents or admins can add children.','孩子管理':'Children','添加孩子':'Add child','孩子姓名':'Child name','年龄':'Age','性别':'Gender','男孩':'Boy','女孩':'Girl','归属家长':'Parent owner','请选择家长':'Choose parent','孩子登录名':'Child login name','孩子登录密码':'Child login password','新增孩子':'Add child','孩子列表':'Children list','姓名':'Name','操作':'Actions','保存':'Save','删除':'Delete','未分配':'Unassigned','当前家长':'Current parent',
  '用户管理':'Users','只有管理员可以管理用户。':'Only admins can manage users.','注册用户':'Register user','显示名称':'Display name','角色':'Role','家长':'Parent','孩子':'Child','新增用户':'Add user','用户列表':'Users','绑定孩子':'Bound child','状态':'Status','启用':'Enabled','已注销':'Deactivated','注销':'Deactivate','留空不改':'Leave blank to keep',
  '本机数据':'Local data','数据目录':'Data dir','数据库':'Database','立即备份':'Backup now','运行提示':'Runtime tips','请定期备份 SQLite 数据库。当前版本默认仅监听本机地址，不会上传云端。':'Please back up the SQLite database regularly. This version listens locally by default and does not upload to the cloud.',
  '家庭共享访问':'Family Sharing','局域网信息加载中...':'Loading network info...','当前仅本机可访问。若要让手机、平板等其他设备使用，请退出后以家庭共享模式启动（设置环境变量':'Currently accessible from this computer only. To use phones, tablets or other devices, quit and start in family sharing mode (set env var','，桌面安装版默认开启），首次启动请在系统防火墙提示中允许访问。':'; the desktop installer enables it by default). Allow access when the system firewall prompts on first launch.','家庭共享已开启。同一 Wi-Fi 下的手机、平板或其他电脑，用浏览器打开下面任一地址即可使用，登录各自账号，数据实时同步。':'Family sharing is on. Phones, tablets or other computers on the same Wi-Fi can open any address below in a browser and sign in with their own accounts. Data stays in sync.','手机扫码直接打开':'Scan with your phone to open',
};

function localizeHtml(html) {
  if (langKey() !== 'en' || !html) return html;
  return Object.entries(zhToEn).sort((a, b) => b[0].length - a[0].length).reduce((out, [zh, en]) => out.split(zh).join(en), html);
}

function loadPrefs() {
  try {
    return { theme: 'system', language: 'zh', ...(JSON.parse(localStorage.getItem('fs_profile_prefs') || '{}')) };
  } catch {
    return { theme: 'system', language: 'zh' };
  }
}

function savePrefs(next) {
  appPrefs = { ...appPrefs, ...next };
  localStorage.setItem('fs_profile_prefs', JSON.stringify(appPrefs));
  applyPreferences();
}

function effectiveTheme() {
  if (appPrefs.theme === 'dark' || appPrefs.theme === 'light') return appPrefs.theme;
  return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function applyPreferences() {
  document.body.dataset.theme = effectiveTheme();
  document.documentElement.lang = appPrefs.language === 'en' ? 'en' : 'zh-CN';
  document.title = tr('appTitle', '家庭德育积分系统');
}

function renderRealtimeClock() {
  return `<button type="button" class="realtime-clock" data-clock-toggle data-clock-mode="${globalClockMode}" aria-label="${h(tr('clockNow','当前时间'))}">
    <div class="clock-digital" data-clock-panel="digital"><span>${h(tr('clockNow','当前时间'))}</span><strong data-clock-digital>--</strong><small>${h(tr('clockToggleHint','点击切换表盘'))}</small></div>
    <div class="clock-analog-panel" data-clock-panel="analog"><div class="clock-date"><span>${h(tr('clockDate','日期'))}</span><strong data-clock-date>--</strong><small>${h(tr('clockToggleHintBack','点击切回数字'))}</small></div><div class="analog-clock" aria-hidden="true">
      ${clockTicksHTML()}${clockNumbersHTML()}
      <i class="clock-hand hour" data-clock-hour-hand></i><i class="clock-hand minute" data-clock-minute-hand></i><i class="clock-hand second" data-clock-second-hand></i><i class="clock-pin"></i>
    </div></div>
  </button>`;
}

function mountRealtimeClock() {
  document.querySelectorAll('[data-clock-toggle]').forEach(el => {
    el.dataset.clockMode = globalClockMode;
    el.onclick = () => {
      globalClockMode = globalClockMode === 'digital' ? 'analog' : 'digital';
      document.querySelectorAll('[data-clock-toggle]').forEach(clock => clock.dataset.clockMode = globalClockMode);
      updateGlobalClock();
    };
  });
  updateGlobalClock();
  if (globalClockTimer) return;
  globalClockTimer = setInterval(updateGlobalClock, 1000);
}

function clockTicksHTML() {
  return Array.from({length: 60}, (_, i) => `<i class="clock-tick ${i % 5 === 0 ? 'major' : ''}" style="--tick-angle:${i * 6}deg"></i>`).join('');
}

function clockNumbersHTML() {
  return Array.from({length: 12}, (_, i) => {
    const n = i + 1;
    const angle = (n * 30 - 90) * Math.PI / 180;
    const x = 50 + Math.cos(angle) * 38;
    const y = 50 + Math.sin(angle) * 38;
    return `<span class="clock-number" style="left:${x.toFixed(3)}%;top:${y.toFixed(3)}%">${n}</span>`;
  }).join('');
}

function updateGlobalClock() {
  const now = new Date();
  document.querySelectorAll('[data-clock-digital]').forEach(el => el.textContent = formatClockDateTime(now));
  document.querySelectorAll('[data-clock-date]').forEach(el => el.textContent = formatClockDate(now));
  const secondDeg = now.getSeconds() * 6;
  const minuteDeg = (now.getMinutes() + now.getSeconds() / 60) * 6;
  const hourDeg = ((now.getHours() % 12) + now.getMinutes() / 60 + now.getSeconds() / 3600) * 30;
  document.querySelectorAll('[data-clock-second-hand]').forEach(el => el.style.transform = `translateX(-50%) rotate(${secondDeg}deg)`);
  document.querySelectorAll('[data-clock-minute-hand]').forEach(el => el.style.transform = `translateX(-50%) rotate(${minuteDeg}deg)`);
  document.querySelectorAll('[data-clock-hour-hand]').forEach(el => el.style.transform = `translateX(-50%) rotate(${hourDeg}deg)`);
}

function formatClockDateTime(date) {
  const d = date instanceof Date ? date : new Date(date);
  if (Number.isNaN(d.getTime())) return '-';
  const pad = n => String(n).padStart(2, '0');
  if (langKey() === 'en') return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
  return `${d.getFullYear()}年${pad(d.getMonth() + 1)}月${pad(d.getDate())}日 ${pad(d.getHours())}时${pad(d.getMinutes())}分${pad(d.getSeconds())}秒`;
}

function formatClockDate(date) {
  const d = date instanceof Date ? date : new Date(date);
  if (Number.isNaN(d.getTime())) return '-';
  const pad = n => String(n).padStart(2, '0');
  if (langKey() === 'en') return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
  return `${d.getFullYear()}年${pad(d.getMonth() + 1)}月${pad(d.getDate())}日`;
}

async function boot() {
  applyPreferences();
  state.status = await api('/api/system/status');
  if (!state.status.setupCompleted) return renderSetup();
  try {
    const me = await api('/api/auth/me');
    state.me = me.user;
    await loadHome();
  } catch {
    renderLogin();
  }
}

async function loadHome() {
  const children = await api('/api/children');
  state.children = children.children || [];
  if (!state.children.some(c => c.id === state.childId)) state.childId = state.children[0]?.id || null;
  await loadAll();
}

async function loadAll() {
  const result = {};
  const childDashboards = {};
  const reqs = [
    api('/api/profile').then(data => result.profile = data.profile || null),
    api('/api/rewards').then(data => result.rewards = data.rewards || []),
    api('/api/exchange-orders').then(data => result.exchangeOrders = data.exchangeOrders || []),
    api('/api/wishes').then(data => result.wishes = data.wishes || []),
    api('/api/appeals').then(data => result.appeals = data.appeals || []),
  ];
  if (canOperate()) {
    reqs.push(api('/api/task-templates').then(data => result.taskTemplates = data.taskTemplates || []));
    reqs.push(api('/api/users').then(data => result.users = data.users || []));
    reqs.push(api('/api/guardian-groups').then(data => result.guardianGroups = data.guardianGroups || []));
    reqs.push(Promise.allSettled(state.children.map(c => api(`/api/children/${c.id}/dashboard`).then(data => { childDashboards[c.id] = data; }))));
  }
  if (state.childId) {
    reqs.push(api(`/api/children/${state.childId}/dashboard`).then(data => result.dashboard = data));
    reqs.push(api(`/api/children/${state.childId}/score-records`).then(data => result.records = data.records || []));
  }
  await Promise.all(reqs);
  state.rewards = result.rewards || [];
  state.profile = result.profile || null;
  state.exchangeOrders = result.exchangeOrders || [];
  state.wishes = result.wishes || [];
  state.appeals = result.appeals || [];
  state.taskTemplates = result.taskTemplates || [];
  state.users = result.users || [];
  state.guardianGroups = result.guardianGroups || [];
  state.dashboard = result.dashboard || null;
  state.childDashboards = childDashboards;
  if (state.childId && state.dashboard && !state.childDashboards[state.childId]) state.childDashboards[state.childId] = state.dashboard;
  state.records = result.records || [];
  renderApp();
}
