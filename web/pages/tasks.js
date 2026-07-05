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
        <div><b>${h(t.taskName)}</b><div class="small">${h(taskCategoryName(t.category))} · ${h(taskSubjectName(t.subject))} · ${h(taskQuestionTypeName(t.questionType))} · +${h(t.scoreValue)} 分</div>${taskContentHTML(t)}</div>
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
  return tr(`taskCategory.${category || 'GROWTH'}`, ({ACTION:'行动',READING:'阅读',MATH:'数学题',STUDY:'学习',SELF_CARE:'自理',HOUSEWORK:'家务',EMOTION:'情绪',HEALTH:'健康',SAFETY:'安全',GROWTH:'成长'}[category] || category || '成长'));
}

function taskSubjectName(subject) {
  return ({GENERAL:'通用',CHINESE:'语文',MATH:'数学',ENGLISH:'英语',SCIENCE:'科学',MORAL:'道德与法治',PE:'体育',ART:'美术',MUSIC:'音乐',LABOR:'劳动',COMPUTER:'信息科技'}[subject] || subject || '通用');
}

function taskQuestionTypeName(type) {
  return ({NONE:'无题目',READING_TEXT:'阅读内容',CHOICE:'选择题',FILL:'填空题',CALCULATION:'计算题',SHORT_ANSWER:'简答题'}[type] || type || '无题目');
}

function taskContentHTML(t) {
  if (!t.content && !t.answer) return '';
  return `<p class="small">${t.content ? `内容：${h(t.content)}` : ''}${t.answer ? `<br>参考答案：${h(t.answer)}` : ''}</p>`;
}

function taskTone(status) {
  return ({TODO:'todo',SUBMITTED:'pending',APPROVED:'done',REJECTED:'rejected'}[status] || '');
}

function renderTaskConfig() {
  return `<div class="split">
    <div class="card"><h2>任务自定义添加</h2><form class="form" id="taskTemplateForm">
      <div class="field"><label>任务名称</label><input name="taskName" placeholder="如：阅读课文 / 数学口算 10 题" required></div>
      <div class="form two"><div class="field"><label>任务类型</label><select name="taskType"><option value="DAILY">每日</option><option value="REPAIR">惩罚/修复</option><option value="TEAM">家庭小队</option><option value="TEMP">临时</option></select></div><div class="field"><label>任务分类</label><select name="category"><option value="ACTION">行动类</option><option value="READING">阅读类</option><option value="MATH">数学题类</option><option value="STUDY">综合学习</option><option value="SELF_CARE">自理</option><option value="HOUSEWORK">家务</option><option value="EMOTION">情绪</option><option value="HEALTH">健康</option><option value="SAFETY">安全</option></select></div></div>
      <div class="form two"><div class="field"><label>科目</label><select name="subject">${taskSubjectOptionsHTML('GENERAL')}</select></div><div class="field"><label>题型</label><select name="questionType"><option value="NONE">无题目</option><option value="READING_TEXT">阅读内容</option><option value="CHOICE">选择题</option><option value="FILL">填空题</option><option value="CALCULATION">计算题</option><option value="SHORT_ANSWER">简答题</option></select></div></div>
      <div class="field"><label>阅读内容 / 题目</label><textarea name="content" placeholder="阅读类可填写要阅读的段落；数学题可填写选择题、填空题或计算题题目。"></textarea></div>
      <div class="field"><label>参考答案</label><input name="answer" placeholder="选填，如：B / 42 / 自主表述"></div>
      <div class="form two"><div class="field"><label>分值</label><input name="scoreValue" type="number" min="1" value="1"></div><div class="field"><label>目标账户</label><select name="targetAccount"><option value="AUTO">自动</option><option value="BASE">基准分</option><option value="TEAM">小队分</option></select></div></div>
      <div class="field"><label>说明</label><textarea name="description"></textarea></div><button>新增任务模板</button>
    </form><div class="preset-block"><h3>任务建议库</h3><p class="small">点击后填入上方表单，适合作为家庭常用任务模板。</p>${taskPresetGrid()}</div></div>
    <div class="card"><h2>任务模板</h2>${state.taskTemplates.length ? `<table class="table"><tbody>${state.taskTemplates.map(t => `<tr><td><b>${h(t.taskName)}</b><div class="small">${h(t.taskType)} · ${h(taskCategoryName(t.category))} · ${h(taskSubjectName(t.subject))} · ${h(taskQuestionTypeName(t.questionType))} · +${h(t.scoreValue)} · ${h(t.description)}</div>${taskContentHTML(t)}</td><td><button class="danger" data-del-task-template="${t.id}">删除</button></td></tr>`).join('')}</tbody></table>` : '<div class="empty-state"><div>🧩</div><b>暂无模板</b><p>添加一些每日可坚持的小任务。</p></div>'}</div>
  </div>`;
}

function taskSubjectOptionsHTML(selected) {
  const subjects = [['GENERAL','通用'],['CHINESE','语文'],['MATH','数学'],['ENGLISH','英语'],['SCIENCE','科学'],['MORAL','道德与法治'],['PE','体育'],['ART','美术'],['MUSIC','音乐'],['LABOR','劳动'],['COMPUTER','信息科技']];
  return subjects.map(([value, label]) => `<option value="${value}" ${value === selected ? 'selected' : ''}>${label}</option>`).join('');
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
    {name:'阅读课文 15 分钟', type:'DAILY', category:'READING', subject:'CHINESE', questionType:'READING_TEXT', value:2, target:'AUTO', content:'阅读今天的语文课文或一本课外书 15 分钟，并说出一个新词或一句喜欢的话。', answer:'能说出阅读内容即可', desc:'阅读后用 1-2 句话分享内容或收获。'},
    {name:'数学选择题 5 题', type:'DAILY', category:'MATH', subject:'MATH', questionType:'CHOICE', value:2, target:'AUTO', content:'例：36 + 27 = ?  A.53  B.63  C.73；可继续添加家庭自定义题目。', answer:'B', desc:'完成选择题并订正错题。'},
    {name:'数学填空题 5 题', type:'DAILY', category:'MATH', subject:'MATH', questionType:'FILL', value:2, target:'AUTO', content:'例：8 × 7 = __；120 - 45 = __。', answer:'56；75', desc:'完成填空题并写清步骤。'},
    {name:'英语单词朗读', type:'DAILY', category:'READING', subject:'ENGLISH', questionType:'READING_TEXT', value:1, target:'AUTO', content:'朗读今天学习的 5 个英语单词，并尝试各造一个短句。', answer:'读准并能说出中文意思', desc:'适合低门槛英语练习。'},
    {name:'睡前整理书包', type:'DAILY', category:'SELF_CARE', subject:'GENERAL', questionType:'NONE', value:1, target:'AUTO', content:'', answer:'', desc:'把作业、课本、文具和水杯准备好。'},
    {name:'饭后收拾餐桌', type:'DAILY', category:'HOUSEWORK', subject:'LABOR', questionType:'NONE', value:1, target:'TEAM', content:'', answer:'', desc:'收拾自己的餐具，并帮助擦桌子。'},
    {name:'情绪冷静复盘', type:'REPAIR', category:'EMOTION', subject:'MORAL', questionType:'SHORT_ANSWER', value:1, target:'BASE', content:'刚才我为什么生气？影响了谁？下次我准备怎么做？', answer:'能说出原因、影响和下一步做法', desc:'情绪爆发后说出原因、影响和下次做法。'},
    {name:'亲子运动 15 分钟', type:'TEAM', category:'HEALTH', subject:'PE', questionType:'NONE', value:3, target:'TEAM', content:'', answer:'', desc:'全家一起跳绳、散步、拉伸或球类运动。'},
  ];
  return `<div class="practice-grid compact">${presets.map(p => `<button type="button" class="practice-card" data-task-name="${h(p.name)}" data-task-type="${p.type}" data-task-category="${p.category}" data-task-subject="${p.subject}" data-task-question-type="${p.questionType}" data-task-content="${h(p.content)}" data-task-answer="${h(p.answer)}" data-task-value="${p.value}" data-task-target="${p.target}" data-task-description="${h(p.desc)}"><span>${taskIcon(p.category)}</span><b>${h(p.name)}</b><small>${h(taskSubjectName(p.subject))} · ${h(taskQuestionTypeName(p.questionType))} · +${p.value}</small></button>`).join('')}</div>`;
}

