function renderTasks() {
  const tasks = state.dashboard?.tasks || [];
  return `<div class="card"><h2>今日任务</h2>${tasks.length ? `<table class="table"><thead><tr><th>任务</th><th>类型</th><th>分值</th><th>状态</th><th>操作</th></tr></thead><tbody>${tasks.map(t => `<tr><td>${h(t.taskName)}<div class="small">${h(t.category)}</div></td><td>${h(t.taskType)}</td><td>+${h(t.scoreValue)}</td><td><span class="tag">${h(t.status)}</span></td><td>${taskActions(t)}</td></tr>`).join('')}</tbody></table>` : '<div class="empty">暂无任务</div>'}</div>`;
}

function taskActions(t) {
  if (t.status === 'TODO') return `<button class="secondary" data-submit-task="${t.id}">完成任务</button>`;
  if (t.status === 'SUBMITTED' && canOperate()) return `<button data-audit-task="${t.id}">通过</button> <button class="secondary" data-reject-task="${t.id}">驳回</button>`;
  if (t.status === 'SUBMITTED') return '<span class="small">等待家长审核</span>';
  return '<span class="small">已处理</span>';
}

function renderTaskConfig() {
  return `<div class="split">
    <div class="card"><h2>任务自定义添加</h2><form class="form" id="taskTemplateForm">
      <div class="field"><label>任务名称</label><input name="taskName" placeholder="如：洗袜子 / 阅读 20 分钟" required></div>
      <div class="form two"><div class="field"><label>任务类型</label><select name="taskType"><option value="DAILY">每日</option><option value="REPAIR">惩罚/修复</option><option value="TEAM">家庭小队</option><option value="TEMP">临时</option></select></div><div class="field"><label>分类</label><select name="category"><option value="STUDY">学习</option><option value="SELF_CARE">自理</option><option value="HOUSEWORK">家务</option><option value="EMOTION">情绪</option><option value="HEALTH">健康</option><option value="SAFETY">安全</option></select></div></div>
      <div class="form two"><div class="field"><label>分值</label><input name="scoreValue" type="number" min="1" value="1"></div><div class="field"><label>目标账户</label><select name="targetAccount"><option value="AUTO">自动</option><option value="BASE">基准分</option><option value="TEAM">小队分</option></select></div></div>
      <div class="field"><label>说明</label><textarea name="description"></textarea></div><button>新增任务模板</button>
    </form></div>
    <div class="card"><h2>任务模板</h2>${state.taskTemplates.length ? `<table class="table"><tbody>${state.taskTemplates.map(t => `<tr><td><b>${h(t.taskName)}</b><div class="small">${h(t.taskType)} · ${h(t.category)} · +${h(t.scoreValue)} · ${h(t.description)}</div></td><td><button class="danger" data-del-task-template="${t.id}">删除</button></td></tr>`).join('')}</tbody></table>` : '<div class="empty">暂无模板</div>'}</div>
  </div>`;
}
