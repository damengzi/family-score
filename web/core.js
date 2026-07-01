const app = document.getElementById('app');
let state = { status: null, me: null, children: [], childId: null, dashboard: null, records: [], rewards: [], taskTemplates: [], exchangeOrders: [], users: [], tab: 'detail' };

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
const statusText = (s) => ({GREEN:'绿色稳定',BLUE:'蓝色提醒',YELLOW:'黄色关注',ORANGE:'橙色预警',RED:'红色重点帮助',DEEP_REPAIR:'深度修复'}[s] || s || '绿色稳定');
const statusClass = (s) => ({GREEN:'green',BLUE:'blue',YELLOW:'yellow',ORANGE:'orange',RED:'red',DEEP_REPAIR:'red'}[s] || 'green');
const recordTypeName = (s) => ({ADD:'加分',DEDUCT:'扣分',REPAIR:'惩罚/修复',TEAM:'家庭小队分',EXCHANGE:'兑换'}[s] || s);
const isAdmin = () => state.me && state.me.role === 'ADMIN';
const canOperate = () => state.me && (state.me.role === 'ADMIN' || state.me.role === 'PARENT');
const roleName = (s) => ({ADMIN:'管理员',PARENT:'家长',CHILD:'孩子'}[s] || s);

async function boot() {
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
  const reqs = [
    api('/api/rewards').then(data => result.rewards = data.rewards || []),
    api('/api/exchange-orders').then(data => result.exchangeOrders = data.exchangeOrders || []),
  ];
  if (canOperate()) {
    reqs.push(api('/api/task-templates').then(data => result.taskTemplates = data.taskTemplates || []));
    reqs.push(api('/api/users').then(data => result.users = data.users || []));
  }
  if (state.childId) {
    reqs.push(api(`/api/children/${state.childId}/dashboard`).then(data => result.dashboard = data));
    reqs.push(api(`/api/children/${state.childId}/score-records`).then(data => result.records = data.records || []));
  }
  await Promise.all(reqs);
  state.rewards = result.rewards || [];
  state.exchangeOrders = result.exchangeOrders || [];
  state.taskTemplates = result.taskTemplates || [];
  state.users = result.users || [];
  state.dashboard = result.dashboard || null;
  state.records = result.records || [];
  renderApp();
}
