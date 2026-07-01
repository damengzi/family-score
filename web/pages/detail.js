function renderDetail() {
  return `<div class="card"><h2>积分明细</h2>${state.records.length ? `<table class="table"><thead><tr><th>时间</th><th>类型</th><th>项目</th><th>分值</th><th>变更</th><th>原因</th></tr></thead><tbody>${state.records.map(r => `<tr><td>${h(r.occurredAt)}</td><td>${recordTypeName(r.recordType)}</td><td>${h(r.itemName)}</td><td>${r.scoreChange > 0 ? '+' : ''}${h(r.scoreChange)}</td><td>${h(r.beforeValue)} → ${h(r.afterValue)} ${h(r.targetAccount)}</td><td>${h(r.reason)}</td></tr>`).join('')}</tbody></table>` : '<div class="empty">暂无积分明细</div>'}</div>`;
}
