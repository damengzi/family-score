function renderDetail() {
  const records = state.records || [];
  return `<div class="card">
    <div class="section-title">
      <div>
        <h2>成长时间线</h2>
        <p class="small">每一次记录，都是一次可回看的家庭反馈。</p>
      </div>
      <span class="tag">${records.length} 条记录</span>
    </div>
    ${records.length ? `<div class="timeline">${records.map(r => `<div class="timeline-item ${recordTone(r.recordType)}">
      <div class="timeline-dot"></div>
      <div class="timeline-content">
        <div class="row"><b>${h(r.itemName || recordTypeName(r.recordType))}</b><span class="tag">${h(recordTypeName(r.recordType))}</span><span class="tag">${scoreDeltaText(r.scoreChange)}</span></div>
        <div class="small">${h(r.occurredAt)} · ${h(r.beforeValue)} → ${h(r.afterValue)} ${h(r.targetAccount)}</div>
        <p>${h(r.reason || '无备注')}</p>
      </div>
    </div>`).join('')}</div>` : '<div class="empty-state"><div>📒</div><b>还没有积分明细</b><p>第一次记录可以从一个主动整理、一次认真阅读或一次修复任务开始。</p></div>'}
  </div>`;
}
