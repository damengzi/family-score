function renderGrowthReport() {
  const acc = state.dashboard?.account || {};
  const tasks = state.dashboard?.tasks || [];
  const records = state.records || [];
  const orders = state.exchangeOrders.filter(o => !state.childId || Number(o.childId) === Number(state.childId));
  const today = todayKey();
  const weekRecords = records.filter(r => isWithinDays(r.occurredAt, 7));
  const monthRecords = records.filter(r => sameMonth(r.occurredAt, new Date()));
  const todayRecords = records.filter(r => String(r.occurredAt || '').slice(0, 10) === today);
  const wishes = wishlistRewards(state.childId);

  return `<div class="stack">
    <div class="card overview-hero-card report-hero">
      <div><div class="eyebrow">成长报告</div><h2>${h(reportTitle(tasks, records))}</h2><p>${h(reportSubtitle(acc, weekRecords))}</p></div>
      <div class="overview-actions"><button class="secondary" data-tab="detail">看时间线</button><button class="secondary" data-tab="tasks">看任务</button></div>
    </div>

    <div class="mini-stat-grid">
      ${miniStat('今日净变化', sumScore(todayRecords), '今日积分记录汇总', '📈')}
      ${miniStat('近7天记录', weekRecords.length, '成长反馈次数', '📒')}
      ${miniStat('本月加分', sumByType(monthRecords, ['ADD','REPAIR','TEAM','STAR']), '正向与修复累计', '🌱')}
      ${miniStat('愿望单', wishes.length, '正在期待的奖励', '🌟')}
    </div>

    <div class="split">
      <div class="card">
        <div class="section-title"><div><h2>今日简报</h2><p class="small">今天的任务、记录和下一步建议。</p></div><span class="tag">${h(today)}</span></div>
        ${dailyReport(tasks, todayRecords)}
      </div>

      <div class="card">
        <div class="section-title"><div><h2>近 7 天趋势</h2><p class="small">看见趋势，比只看单次分数更重要。</p></div><span class="tag">7 天</span></div>
        ${trendReport(weekRecords, orders)}
      </div>
    </div>

    <div class="split">
      <div class="card">
        <div class="section-title"><div><h2>本月关键词</h2><p class="small">按记录类型自动整理当前月份的成长结构。</p></div><span class="tag">${h(currentMonthKey())}</span></div>
        ${monthlyReport(monthRecords)}
      </div>

      <div class="card">
        <div class="section-title"><div><h2>愿望单</h2><p class="small">把想要的奖励先放进愿望单，慢慢攒分。</p></div><button class="secondary" data-tab="rewards">管理愿望</button></div>
        ${renderWishlistPanel(state.childId)}
      </div>
    </div>
  </div>`;
}

function reportTitle(tasks, records) {
  if (!state.childId) return '先选择一个孩子查看成长报告';
  const done = tasks.filter(t => t.status === 'APPROVED').length;
  if (done) return `今天已经点亮 ${done} 个任务`;
  if (records.length) return '成长记录正在累积，继续保持节奏';
  return '从今天开始记录第一个小进步';
}

function reportSubtitle(acc, weekRecords) {
  return `当前基准分 ${acc.baseScore ?? 100}，状态为 ${statusText(acc.statusLevel)}；近 7 天共有 ${weekRecords.length} 条成长记录。`;
}

function dailyReport(tasks, records) {
  const todo = tasks.filter(t => t.status === 'TODO').length;
  const submitted = tasks.filter(t => t.status === 'SUBMITTED').length;
  const approved = tasks.filter(t => t.status === 'APPROVED').length;
  return `<div class="report-list">
    ${reportRow('待挑战任务', todo, '还可以完成的小目标')}
    ${reportRow('等待确认', submitted, canOperate() ? '可以到待审核中心处理' : '等待家长确认')}
    ${reportRow('已点亮任务', approved, '今天已经完成的任务')}
    ${reportRow('今日净变化', sumScore(records), '今日积分账户变化')}
  </div>`;
}

function trendReport(records, orders) {
  const applied = orders.filter(o => isWithinDays(o.appliedAt, 7)).length;
  const approved = orders.filter(o => o.status === 'APPROVED' && isWithinDays(o.appliedAt, 7)).length;
  return `<div class="report-list">
    ${reportRow('近7天净变化', sumScore(records), '含加分、扣分、修复、小队分和星星')}
    ${reportRow('正向记录', records.filter(r => Number(r.scoreChange) > 0).length, '加分、修复或奖励性记录')}
    ${reportRow('提醒记录', records.filter(r => Number(r.scoreChange) < 0).length, '需要复盘的扣分记录')}
    ${reportRow('兑换申请', `${applied}/${approved}`, '近7天申请 / 已通过')}
  </div>`;
}

function monthlyReport(records) {
  return `<div class="report-list">
    ${reportRow('加分表现', countByType(records, 'ADD'), `累计 ${sumByType(records, ['ADD'])} 分`)}
    ${reportRow('修复行动', countByType(records, 'REPAIR'), `累计修复 ${sumByType(records, ['REPAIR'])} 分`)}
    ${reportRow('家庭小队', countByType(records, 'TEAM'), `小队分变化 ${sumByType(records, ['TEAM'])}`)}
    ${reportRow('需要提醒', countByType(records, 'DEDUCT'), `扣分合计 ${sumByType(records, ['DEDUCT'])}`)}
  </div>`;
}

function reportRow(label, value, hint) {
  return `<div class="report-row"><div><b>${h(label)}</b><div class="small">${h(hint)}</div></div><strong>${h(value)}</strong></div>`;
}

function sumScore(records) {
  return records.reduce((sum, r) => sum + Number(r.scoreChange || 0), 0);
}

function countByType(records, type) {
  return records.filter(r => r.recordType === type).length;
}

function sumByType(records, types) {
  return records.filter(r => types.includes(r.recordType)).reduce((sum, r) => sum + Number(r.scoreChange || 0), 0);
}

function isWithinDays(value, days) {
  const d = parseLocalDate(value);
  if (!d) return false;
  const start = new Date();
  start.setHours(0, 0, 0, 0);
  start.setDate(start.getDate() - days + 1);
  return d >= start;
}

function sameMonth(value, now) {
  const d = parseLocalDate(value);
  return !!d && d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth();
}

function parseLocalDate(value) {
  if (!value) return null;
  const normalized = String(value).replace(' ', 'T');
  const d = new Date(normalized);
  return Number.isNaN(d.getTime()) ? null : d;
}

function currentMonthKey() {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
}
