function renderAuditCenter() {
  if (!canOperate()) return '<div class="card"><h2>无权限</h2><div class="notice">只有家长或管理员可以查看待审核中心。</div></div>';
  const tasks = pendingAuditTasks();
  const orders = state.exchangeOrders.filter(o => o.status === 'PENDING');
  return `<div class="stack">
    <div class="card overview-hero-card audit-hero">
      <div><div class="eyebrow">审核工作台</div><h2>今天有 ${tasks.length + orders.length} 个事项等待确认</h2><p>把任务确认、兑换审核集中处理，减少在多个页面来回切换。</p></div>
      <div class="overview-actions"><button data-tab="tasks" class="secondary">看今日任务</button><button data-tab="rewards" class="secondary">看奖励兑换</button></div>
    </div>

    <div class="mini-stat-grid">
      ${miniStat('待确认任务', tasks.length, '孩子已提交完成', '✅')}
      ${miniStat('待审核兑换', orders.length, '奖励申请待处理', '🎁')}
      ${miniStat('涉及孩子', auditChildCount(tasks, orders), '需要关注的孩子', '👧')}
      ${miniStat('今日总待办', tasks.length + orders.length, '集中确认事项', '📌')}
    </div>

    <div class="split">
      <div class="card">
        <div class="section-title"><div><h2>待确认任务</h2><p class="small">建议确认前先看孩子提交的说明，并给一句具体鼓励。</p></div><span class="tag">${tasks.length} 个</span></div>
        ${tasks.length ? `<div class="task-card-list">${tasks.map(t => `<div class="task-card pending">
          <div class="task-main"><div class="task-icon">${taskIcon(t.category)}</div><div><b>${h(t.taskName)}</b><div class="small">${h(auditChildName(t.childId))} · ${h(taskCategoryName(t.category))} · ${h(taskSubjectName(t.subject))} · +${h(t.scoreValue)} 分</div><p class="small">${h(t.submitNote || '孩子已提交完成')}</p></div></div>
          <div class="task-side"><button data-audit-task="${t.id}">确认完成</button><button class="secondary" data-reject-task="${t.id}">再试一次</button></div>
        </div>`).join('')}</div>` : '<div class="empty-state"><div>✅</div><b>暂无待确认任务</b><p>孩子提交任务后会集中出现在这里。</p></div>'}
      </div>

      <div class="card">
        <div class="section-title"><div><h2>待审核兑换</h2><p class="small">确认前可结合基准分状态、健康风险和兑换频率判断。</p></div><span class="tag">${orders.length} 个</span></div>
        ${orders.length ? `<div class="task-card-list">${orders.map(o => `<div class="task-card pending">
          <div class="task-main"><div class="task-icon">🎁</div><div><b>${h(o.rewardName)}</b><div class="small">${h(auditChildName(o.childId))} · ${h(o.appliedAt)} · ${o.costScore}分 / ${o.costStar}星</div><p class="small">${h(o.applyNote || '申请兑换')}</p></div></div>
          <div class="task-side"><button data-audit-order="${o.id}">确认兑换</button><button class="secondary" data-reject-order="${o.id}">暂不兑换</button></div>
        </div>`).join('')}</div>` : '<div class="empty-state"><div>🎁</div><b>暂无待审核兑换</b><p>孩子发起奖励申请后会集中出现在这里。</p></div>'}
      </div>
    </div>
  </div>`;
}

function pendingAuditTasks() {
  const items = [];
  Object.values(state.childDashboards || {}).forEach(d => (d.tasks || []).forEach(t => {
    if (t.status === 'SUBMITTED') items.push(t);
  }));
  return items.sort((a, b) => String(a.taskDate || '').localeCompare(String(b.taskDate || '')));
}

function auditChildName(childId) {
  const child = state.children.find(c => Number(c.id) === Number(childId));
  return child ? child.name : `孩子 ${childId || '-'}`;
}

function auditChildCount(tasks, orders) {
  return new Set([...tasks.map(t => Number(t.childId)), ...orders.map(o => Number(o.childId))].filter(Boolean)).size;
}
