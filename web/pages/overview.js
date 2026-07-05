function renderOverview() {
  const acc = state.dashboard?.account || {};
  const tasks = state.dashboard?.tasks || [];
  const records = state.records || [];
  const rewards = state.rewards || [];
  const orders = state.exchangeOrders || [];

  const todoTasks = tasks.filter(t => t.status === 'TODO');
  const pendingTasks = canOperate() ? pendingAuditTasks() : tasks.filter(t => t.status === 'SUBMITTED');
  const pendingOrders = orders.filter(o => o.status === 'PENDING');
  const todayRecords = records.filter(r => String(r.occurredAt || '').slice(0, 10) === todayKey());

  return `<div class="stack">
    <div class="card overview-hero-card">
      <div>
        <div class="eyebrow">${h(roleName(state.me?.role))} · 今日成长概览</div>
        <h2>${h(overviewTitle(acc, todoTasks, pendingTasks, pendingOrders))}</h2>
        <p>${h(overviewSubtitle(acc))}</p>
      </div>
      <div class="overview-actions">
        ${overviewPrimaryActions(todoTasks, pendingTasks, pendingOrders)}
      </div>
    </div>

    <div class="mini-stat-grid">
      ${miniStat('待完成任务', todoTasks.length, '今日还可以挑战', '🧩')}
      ${miniStat('等待确认', canOperate() ? pendingTasks.length + pendingOrders.length : pendingTasks.length, canOperate() ? '需要家长处理' : '等待家长审核', '⏳')}
      ${miniStat('今日记录', todayRecords.length, '成长日志条数', '📒')}
      ${miniStat('愿望单', wishlistCount(state.childId), '正在期待的奖励', '🌟')}
    </div>

    <div class="split">
      <div class="card">
        <div class="section-title">
          <div>
            <h2>今日任务</h2>
            <p class="small">把每天的小事变成看得见的成长。</p>
          </div>
          <button class="secondary" data-tab="tasks">查看全部</button>
        </div>
        ${overviewTasks(tasks)}
      </div>

      <div class="card">
        <div class="section-title">
          <div>
            <h2>分值状态</h2>
            <p class="small">当前状态和下一步建议。</p>
          </div>
          <span class="tag ${statusClass(acc.statusLevel)}">${statusText(acc.statusLevel)}</span>
        </div>
        ${scoreCompass(acc)}
      </div>
    </div>

    <div class="split">
      <div class="card">
        <div class="section-title">
          <div>
            <h2>奖励小货架</h2>
            <p class="small">把努力兑换成小小期待。</p>
          </div>
          <button class="secondary" data-tab="rewards">去兑换</button>
        </div>
        ${overviewRewards(rewards)}
      </div>

      <div class="card">
        <div class="section-title">
          <div>
            <h2>最新成长记录</h2>
            <p class="small">最近发生的加分、扣分、修复和兑换。</p>
          </div>
          <button class="secondary" data-tab="detail">看明细</button>
        </div>
        ${overviewRecords(records)}
      </div>
    </div>
  </div>`;
}

function overviewTitle(acc, todoTasks, pendingTasks, pendingOrders) {
  if (!state.childId) return '先添加一个孩子档案，开始记录家庭成长';
  if (state.me?.role === 'CHILD') {
    if (todoTasks.length) return `今天还有 ${todoTasks.length} 个成长任务可以挑战`;
    if (pendingTasks.length) return '任务已经提交，等待家长确认';
    return '今天表现不错，继续保持节奏';
  }
  if (canOperate()) {
    const count = pendingTasks.length + pendingOrders.length;
    if (count) return `今天有 ${count} 个事项需要确认`;
    return '今天暂时没有待审核事项';
  }
  return `当前基准德育分 ${acc.baseScore ?? 100}`;
}

function overviewSubtitle(acc) {
  const status = statusText(acc.statusLevel);
  if (state.me?.role === 'CHILD') return `当前状态：${status}。完成任务、积攒星星，慢慢解锁喜欢的奖励。`;
  if (state.me?.role === 'PARENT') return `当前状态：${status}。多一点确认、鼓励和修复建议，会让规则更容易被坚持。`;
  if (state.me?.role === 'ADMIN') return `当前状态：${status}。可以从这里统筹家庭成员、任务、奖励和本机备份。`;
  return '欢迎来到家庭德育积分系统。';
}

function overviewPrimaryActions(todoTasks, pendingTasks, pendingOrders) {
  if (!state.childId) return canOperate() ? '<button data-tab="childConfig">添加孩子</button>' : '<button data-tab="detail">查看明细</button>';
  if (state.me?.role === 'CHILD') {
    if (todoTasks.length) return '<button data-tab="tasks">去完成任务</button><button class="secondary" data-tab="rewards">看看奖励</button>';
    return '<button data-tab="rewards">看看能兑换什么</button>';
  }
  if (canOperate()) {
    if (pendingTasks.length || pendingOrders.length) return '<button data-tab="auditCenter">处理待审核</button><button class="secondary" data-tab="growthReport">看成长报告</button>';
    return '<button data-tab="score">记录一次成长</button>';
  }
  return '<button data-tab="detail">查看明细</button>';
}

function miniStat(label, value, hint, icon) {
  return `<div class="card mini-stat">
    <div class="mini-stat-icon">${icon}</div>
    <div>
      <div class="label">${h(label)}</div>
      <div class="value">${h(value)}</div>
      <div class="small">${h(hint)}</div>
    </div>
  </div>`;
}

function overviewTasks(tasks) {
  if (!tasks.length) return '<div class="empty-state"><div>🧩</div><b>今天还没有任务</b><p>可以去任务自定义里添加一些成长任务。</p></div>';
  return `<div class="task-card-list">${tasks.slice(0, 4).map(t => `<div class="task-card ${taskTone(t.status)}">
    <div class="task-main">
      <div class="task-icon">${taskIcon(t.category)}</div>
      <div>
        <b>${h(t.taskName)}</b>
        <div class="small">${h(taskCategoryName(t.category))} · ${h(t.taskType)} · +${h(t.scoreValue)} 分</div>
      </div>
    </div>
    <div class="task-side">
      <span class="tag">${h(taskStatusText(t.status))}</span>
      <div>${taskActions(t)}</div>
    </div>
  </div>`).join('')}</div>`;
}

function scoreCompass(acc) {
  const base = Number(acc.baseScore ?? 100);
  const percent = Math.max(0, Math.min(100, Number.isFinite(base) ? base : 100));
  return `<div class="score-compass">
    <div class="progress-ring" style="background:conic-gradient(var(--role-accent) 0 ${percent}%, #edf0f7 ${percent}% 100%)">
      <div class="progress-ring-inner">
        <strong>${h(base)}</strong>
        <span>基准分</span>
      </div>
    </div>
    <div class="stack">
      <div class="progress-line"><span style="width:${percent}%"></span></div>
      <div class="notice">${h(scoreAdvice(acc.statusLevel, base))}</div>
      <div class="row">
        <span class="tag">兑换分 ${h(acc.bonusScore ?? 0)}</span>
        <span class="tag">星星 ${h(acc.starCount ?? 0)}</span>
        <span class="tag">小队分 ${h(acc.teamScore ?? 0)}</span>
      </div>
    </div>
  </div>`;
}

function scoreAdvice(status, base) {
  if (status === 'GREEN') return '状态稳定，可以继续通过任务和正向行为积累兑换分。';
  if (status === 'BLUE') return '有轻微提醒，建议今天选择一个容易完成的小任务找回节奏。';
  if (status === 'YELLOW') return '需要关注，家长可以一起制定一个明确、可完成的修复任务。';
  if (status === 'ORANGE') return '处于预警状态，建议减少兑换刺激，先恢复基础行为。';
  if (status === 'RED' || status === 'DEEP_REPAIR') return '建议进入陪伴修复模式，先完成小目标，不急着追求高分。';
  if (base < 90) return '基准分低于 90，建议优先完成修复任务。';
  return '当前状态不错，继续保持稳定节奏。';
}

function overviewRewards(rewards) {
  if (!rewards.length) return '<div class="empty-state"><div>🎁</div><b>奖励货架还是空的</b><p>可以添加图书、活动、特权或小零食。</p></div>';
  return `<div class="reward-shelf compact">${rewards.slice(0, 4).map(r => `<div class="reward-card">
    <div class="reward-emoji">${rewardIcon(r.rewardType)}</div>
    <h3>${h(r.rewardName)} ${isWishlisted(state.childId, r.id) ? '<span class="wish-badge">想要</span>' : ''}</h3>
    <p class="small">${h(r.description || '一个值得期待的小奖励')}</p>
    <div class="row">
      <span class="tag">${h(rewardCostText(r))}</span>
      <span class="tag ${r.healthRisk === 'NONE' ? 'green' : 'yellow'}">${h(riskName(r.healthRisk))}</span>
    </div>
    <div class="row"><button data-reward="${r.id}">申请兑换</button><button class="secondary" data-wishlist="${r.id}">${isWishlisted(state.childId, r.id) ? '移出愿望' : '加入愿望'}</button></div>
  </div>`).join('')}</div>`;
}

function overviewRecords(records) {
  if (!records.length) return '<div class="empty-state"><div>📒</div><b>还没有成长记录</b><p>第一次记录可以从一次小小的主动行为开始。</p></div>';
  return `<div class="timeline compact">${records.slice(0, 5).map(r => `<div class="timeline-item ${recordTone(r.recordType)}">
    <div class="timeline-dot"></div>
    <div>
      <b>${h(r.itemName || recordTypeName(r.recordType))}</b>
      <div class="small">${h(r.occurredAt)} · ${h(recordTypeName(r.recordType))} · ${scoreDeltaText(r.scoreChange)}</div>
      <p>${h(r.reason || '无备注')}</p>
    </div>
  </div>`).join('')}</div>`;
}

function todayKey() {
  const d = new Date();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${d.getFullYear()}-${m}-${day}`;
}

function scoreDeltaText(v) {
  const n = Number(v || 0);
  return `${n > 0 ? '+' : ''}${n}`;
}

function taskIcon(category) {
  return ({ACTION:'✅',READING:'📖',MATH:'🧮',STUDY:'📚',SELF_CARE:'🎒',HOUSEWORK:'🧹',EMOTION:'💛',HEALTH:'🏃',SAFETY:'🛡️'}[category] || '🧩');
}

function rewardIcon(type) {
  return ({SNACK:'🍪',DRINK:'🥤',BOOK:'📚',TOY:'🧸',ACTIVITY:'🎡',PRIVILEGE:'🎟️'}[type] || '🎁');
}

function recordTone(type) {
  return ({ADD:'positive',DEDUCT:'warning',REPAIR:'repair',TEAM:'team',STAR:'star',EXCHANGE:'exchange'}[type] || '');
}
