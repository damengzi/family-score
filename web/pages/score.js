function renderScoreForm() {
  return `<div class="stack">
    <div class="split">
    <div class="card"><div class="section-title"><div><h2>加分 / 扣分 / 惩罚修复</h2><p class="small">建议先描述事实，再记录分值；分数是反馈工具，不是情绪出口。</p></div><span class="tag">行为记录助手</span></div><form class="form" id="scoreForm">
      <div class="field"><label>类型</label><select name="recordType"><option value="ADD">加分</option><option value="DEDUCT">扣分</option><option value="REPAIR">惩罚/修复加回基准分</option><option value="TEAM">家庭小队分</option><option value="STAR">星星</option></select></div>
      <div class="field"><label>项目名称</label><input name="itemName" placeholder="如：主动整理书包 / 作业拖延 / 整理书桌修复" required></div>
      <div class="field"><label>分值</label><input name="scoreChange" type="number" value="1" min="1" required></div>
      <div class="field"><label>原因</label><textarea name="reason" placeholder="简要说明"></textarea></div>
      <button>记录这次成长</button>
    </form></div>
    <div class="card"><h2>使用提醒</h2><div class="notice">扣分只扣基准分；修复任务只恢复基准分；基准分满 100 后，加分进入超额兑换分。</div><div class="practice-list"><div><b>扣分前先问</b><p class="small">这次记录的是具体行为，还是我的情绪？孩子知道下次怎么做吗？</p></div><div><b>修复要可完成</b><p class="small">优先设计 5-15 分钟能完成的小任务，让孩子看见“补救路径”。</p></div><div><b>加分要具体</b><p class="small">不要只说“你很棒”，尽量说清楚棒在哪里。</p></div></div></div>
    </div>
    <div class="card"><div class="section-title"><div><h2>常用场景模板</h2><p class="small">点击模板会填入上方表单，提交前仍可修改。</p></div><span class="tag">快捷填入</span></div>${scorePresetGrid()}</div>
  </div>`;
}

function scorePresetGrid() {
  const presets = [
    ['ADD','主动整理书包',1,'放学后没有提醒，主动把书包、作业和文具整理好。','🎒'],
    ['ADD','认真阅读 20 分钟',2,'阅读过程中专注安静，并能简单分享一个收获。','📚'],
    ['DEDUCT','作业明显拖延',1,'已经约定开始时间，但多次提醒后仍未进入学习状态。','⏰'],
    ['DEDUCT','情绪表达不当',1,'出现大声顶撞或摔东西，需要复盘更合适的表达方式。','💬'],
    ['REPAIR','整理书桌修复',1,'完成书桌整理，并说明下次如何保持。','🧹'],
    ['REPAIR','道歉和补救',1,'向被影响的家人表达歉意，并完成一个具体补救行动。','🤝'],
    ['TEAM','家庭阅读共读',3,'全家共同阅读或分享，计入家庭小队分。','🏠'],
    ['STAR','连续一周稳定完成',1,'连续保持一项好习惯，可奖励一颗星星。','⭐'],
  ];
  return `<div class="practice-grid">${presets.map(p => `<button type="button" class="practice-card" data-score-type="${p[0]}" data-score-item="${h(p[1])}" data-score-value="${p[2]}" data-score-reason="${h(p[3])}"><span>${p[4]}</span><b>${h(p[1])}</b><small>${h(recordTypeName(p[0]))} · ${p[2]} 分</small></button>`).join('')}</div>`;
}

