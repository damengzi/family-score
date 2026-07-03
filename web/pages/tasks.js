function renderTasks() {
  const tasks = state.dashboard?.tasks || [];
  return `<div class="stack"><div class="card">
    <div class="section-title">
      <div>
        <h2>今日任务</h2>
        <p class="small">完成后提交给家长确认，通过后才会加分。</p>
      </div>
      <span class="tag">共 ${tasks.length} 个</span>
    </div>
    ${tasks.length ? `<div class="task-card-list">${tasks.map(t => `<div class="task-card ${taskTone(t.status)}">
      <div class="task-main">
        <div class="task-icon">${taskIcon(t.category)}</div>
        <div><b>${h(t.taskName)}</b><div class="small">${h(taskCategoryName(t.category))} · ${h(t.taskType)} · +${h(t.scoreValue)} 分</div></div>
      </div>
      <div class="task-side"><span class="tag">${h(taskStatusText(t.status))}</span><div>${taskActions(t)}</div></div>
    </div>`).join('')}</div>` : '<div class="empty-state"><div>🧩</div><b>今天还没有任务</b><p>可以去任务自定义中添加一些适合今天的小目标。</p></div>'}
  </div><div class="card"><div class="section-title"><div><h2>今日实践建议</h2><p class="small">不知道做什么时，可以从这些低门槛任务开始。</p></div><span class="tag">家庭可用</span></div>${dailyPracticeIdeas()}</div></div>`;
}

function taskActions(t) {
  if (t.status === 'TODO') return `<button class="secondary" data-submit-task="${t.id}">我完成啦</button>`;
  if (t.status === 'SUBMITTED' && canOperate()) return `<button data-audit-task="${t.id}">确认完成</button> <button class="secondary" data-reject-task="${t.id}">再试一次</button>`;
  if (t.status === 'SUBMITTED') return '<span class="small">等待家长确认</span>';
  if (t.status === 'APPROVED') return '<span class="small">已点亮</span>';
  if (t.status === 'REJECTED') return '<span class="small">可以重新努力</span>';
  return '<span class="small">已处理</span>';
}

function taskStatusText(status) {
  return tr(`taskStatus.${status}`, ({TODO:'待挑战',SUBMITTED:'等待确认',APPROVED:'已点亮',REJECTED:'再试一次'}[status] || status));
}

function taskCategoryName(category) {
  return tr(`taskCategory.${category || 'GROWTH'}`, ({STUDY:'学习',SELF_CARE:'自理',HOUSEWORK:'家务',EMOTION:'情绪',HEALTH:'健康',SAFETY:'安全'}[category] || category || '成长'));
}

function taskTone(status) {
  return ({TODO:'todo',SUBMITTED:'pending',APPROVED:'done',REJECTED:'rejected'}[status] || '');
}

function renderTaskConfig() {
  return `<div class="split">
    <div class="card"><h2>任务自定义添加</h2><form class="form" id="taskTemplateForm">
      <div class="field"><label>任务名称</label><input name="taskName" placeholder="如：洗袜子 / 阅读 20 分钟" required></div>
      <div class="form two"><div class="field"><label>任务类型</label><select name="taskType"><option value="DAILY">每日</option><option value="REPAIR">惩罚/修复</option><option value="TEAM">家庭小队</option><option value="TEMP">临时</option></select></div><div class="field"><label>分类</label><select name="category"><option value="STUDY">学习</option><option value="SELF_CARE">自理</option><option value="HOUSEWORK">家务</option><option value="EMOTION">情绪</option><option value="HEALTH">健康</option><option value="SAFETY">安全</option></select></div></div>
      <div class="form two"><div class="field"><label>分值</label><input name="scoreValue" type="number" min="1" value="1"></div><div class="field"><label>目标账户</label><select name="targetAccount"><option value="AUTO">自动</option><option value="BASE">基准分</option><option value="TEAM">小队分</option></select></div></div>
      <div class="field"><label>说明</label><textarea name="description"></textarea></div><button>新增任务模板</button>
    </form><div class="preset-block"><h3>任务建议库</h3><p class="small">点击后填入上方表单，适合作为家庭常用任务模板。</p>${taskPresetGrid()}</div></div>
    <div class="card"><h2>任务模板</h2>${state.taskTemplates.length ? `<table class="table"><tbody>${state.taskTemplates.map(t => `<tr><td><b>${h(t.taskName)}</b><div class="small">${h(t.taskType)} · ${h(t.category)} · +${h(t.scoreValue)} · ${h(t.description)}</div></td><td><button class="danger" data-del-task-template="${t.id}">删除</button></td></tr>`).join('')}</tbody></table>` : '<div class="empty-state"><div>🧩</div><b>暂无模板</b><p>添加一些每日可坚持的小任务。</p></div>'}</div>
  </div>`;
}

function dailyPracticeIdeas() {
  const ideas = [
    ['📚','学习','阅读 20 分钟后说出一个新知识。'],
    ['🎒','自理','睡前整理书包和第二天衣物。'],
    ['🧹','家务','饭后收拾自己的餐具并擦桌子。'],
    ['💛','情绪','生气后用一句话说出“我需要什么”。'],
    ['🏃','健康','和家人一起运动或拉伸 15 分钟。'],
    ['🛡️','安全','复述一条今天用得上的安全规则。'],
  ];
  return `<div class="practice-grid">${ideas.map(i => `<div class="practice-card static"><span>${i[0]}</span><b>${h(i[1])}</b><small>${h(i[2])}</small></div>`).join('')}</div>`;
}

function taskPresetGrid() {
  const presets = [
    ['阅读 20 分钟','DAILY','STUDY',2,'AUTO','阅读后用 1-2 句话分享内容或收获。'],
    ['睡前整理书包','DAILY','SELF_CARE',1,'AUTO','把作业、课本、文具和水杯准备好。'],
    ['饭后收拾餐桌','DAILY','HOUSEWORK',1,'TEAM','收拾自己的餐具，并帮助擦桌子。'],
    ['情绪冷静复盘','REPAIR','EMOTION',1,'BASE','情绪爆发后说出原因、影响和下次做法。'],
    ['亲子运动 15 分钟','TEAM','HEALTH',3,'TEAM','全家一起跳绳、散步、拉伸或球类运动。'],
    ['安全规则复述','TEMP','SAFETY',1,'AUTO','说出一条今天要遵守的交通、用电或外出安全规则。'],
  ];
  return `<div class="practice-grid compact">${presets.map(p => `<button type="button" class="practice-card" data-task-name="${h(p[0])}" data-task-type="${p[1]}" data-task-category="${p[2]}" data-task-value="${p[3]}" data-task-target="${p[4]}" data-task-description="${h(p[5])}"><span>${taskIcon(p[2])}</span><b>${h(p[0])}</b><small>${h(taskCategoryName(p[2]))} · +${p[3]}</small></button>`).join('')}</div>`;
}

