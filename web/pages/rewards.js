function renderRewards() {
  const pending = state.exchangeOrders.filter(o => o.status === 'PENDING');
  const rewardCards = state.rewards.map(r => `<div class="reward-card">
    <div class="reward-emoji">${rewardIcon(r.rewardType)}</div>
    <h3>${h(r.rewardName)} ${isWishlisted(state.childId, r.id) ? '<span class="wish-badge">想要</span>' : ''}</h3>
    <p class="small">${h(r.description || '一个值得期待的小奖励')}</p>
    <div class="row"><span class="tag">${h(rewardTypeName(r.rewardType))}</span><span class="tag">${h(rewardCostText(r))}</span><span class="tag ${r.healthRisk==='NONE'?'green':r.healthRisk==='HIGH'?'red':'yellow'}">${h(riskName(r.healthRisk))}</span></div>
    <div class="row"><button data-reward="${r.id}">申请这个奖励</button><button class="secondary" data-wishlist="${r.id}">${isWishlisted(state.childId, r.id) ? '移出愿望单' : '加入愿望单'}</button></div>
  </div>`).join('');
  const auditPanel = canOperate() ? `<div class="card"><div class="section-title"><div><h2>待审核兑换</h2><p class="small">确认前可以看看孩子当前分值和奖励健康风险。</p></div><span class="tag">${pending.length} 个待处理</span></div>${pending.length ? `<div class="task-card-list">${pending.map(o => `<div class="task-card pending"><div class="task-main"><div class="task-icon">🎁</div><div><b>${h(o.rewardName)}</b><div class="small">${h(o.appliedAt)} · ${o.costScore}分 / ${o.costStar}星</div></div></div><div class="task-side"><button data-audit-order="${o.id}">确认兑换</button><button class="secondary" data-reject-order="${o.id}">暂不兑换</button></div></div>`).join('')}</div>` : '<div class="empty-state"><div>🎁</div><b>暂无待审核兑换</b><p>孩子还在积攒兑换能量。</p></div>'}</div>` : '';
  return `<div class="stack"><div class="card"><div class="section-title"><div><h2>奖励商店</h2><p class="small">零食、图书、活动和特权，都可以成为努力后的期待。</p></div><span class="tag">${state.rewards.length} 个奖励</span></div><div class="notice">建议优先配置“图书、亲子活动、运动体验、家庭特权”等低风险奖励；零食饮品适合低频、限量、明确规则。</div><br>${state.rewards.length ? `<div class="reward-shelf">${rewardCards}</div>` : '<div class="empty-state"><div>🎁</div><b>奖励货架还是空的</b><p>可以先配置一个小奖励，比如科普书、亲子活动或小零食。</p></div>'}</div><div class="card"><div class="section-title"><div><h2>我的愿望单</h2><p class="small">先收藏想要的奖励，再通过任务一点点靠近它。</p></div><span class="tag">${wishlistCount(state.childId)} 个愿望</span></div>${renderWishlistPanel(state.childId)}</div>${rewardPrinciples()}${auditPanel}</div>`;
}

function rewardCostText(r) {
  if (Number(r.costStar || 0) > 0) return `${r.costStar} 星`;
  return `${r.costScore || 0} 兑换分`;
}

function rewardTypeName(type) {
  return ({SNACK:'零食',DRINK:'饮品',BOOK:'图书',TOY:'玩具',ACTIVITY:'活动',PRIVILEGE:'特权'}[type] || type || '奖励');
}

function riskName(risk) {
  return ({NONE:'健康推荐',LOW:'低风险',MEDIUM:'适量',HIGH:'谨慎'}[risk] || risk || '未标记');
}

function renderRewardConfig() {
  return `<div class="split">
    <div class="card"><h2>奖励/零食自定义添加</h2><form class="form" id="rewardForm">
      <div class="field"><label>奖励名称</label><input name="rewardName" placeholder="如：薯片一包 / 奶茶小杯 / 科普书" required></div>
      <div class="form two"><div class="field"><label>类型</label><select name="rewardType"><option value="SNACK">零食</option><option value="DRINK">饮品</option><option value="BOOK">图书</option><option value="TOY">玩具</option><option value="ACTIVITY">活动</option><option value="PRIVILEGE">特权</option></select></div><div class="field"><label>健康风险</label><select name="healthRisk"><option value="NONE">无</option><option value="LOW">低</option><option value="MEDIUM">中</option><option value="HIGH">高/禁用</option></select></div></div>
      <div class="form two"><div class="field"><label>消耗兑换分</label><input name="costScore" type="number" min="0" value="5"></div><div class="field"><label>消耗星星</label><input name="costStar" type="number" min="0" value="0"></div></div>
      <div class="form two"><div class="field"><label>周限制</label><input name="weeklyLimit" type="number" min="1" value="1"></div><div class="field"><label>月限制</label><input name="monthlyLimit" type="number" min="1" value="4"></div></div>
      <div class="field"><label>说明</label><textarea name="description"></textarea></div><button>新增奖励</button>
    </form><div class="preset-block"><h3>奖励配置建议</h3><p class="small">点击后填入上方表单，适合作为家庭奖励货架起点。</p>${rewardPresetGrid()}</div></div>
    <div class="card"><h2>奖励配置</h2>${state.rewards.length ? `<table class="table"><tbody>${state.rewards.map(r => `<tr><td><b>${h(r.rewardName)}</b><div class="small">${h(r.rewardType)} · ${r.costScore}分 · ${r.costStar}星 · ${h(r.healthRisk)} · ${h(r.description)}</div></td><td><button class="danger" data-del-reward="${r.id}">删除</button></td></tr>`).join('')}</tbody></table>` : '<div class="empty-state"><div>🎁</div><b>暂无奖励</b><p>添加一个孩子愿意努力争取的小目标。</p></div>'}</div>
  </div>`;
}

function rewardPrinciples() {
  return `<div class="card"><div class="section-title"><div><h2>健康兑换原则</h2><p class="small">让奖励成为期待，而不是新的拉扯。</p></div><span class="tag">家长参考</span></div><div class="speech-grid">
    <div class="speech-card"><span>低风险优先</span><p>图书、文具、亲子活动、运动体验、家庭游戏时间，适合作为常规奖励。</p></div>
    <div class="speech-card"><span>零食要限量</span><p>零食饮品建议设置周限制和月限制，并避免高咖啡因、高糖、无限量。</p></div>
    <div class="speech-card"><span>不兑换底线</span><p>不建议兑换免作业、延迟睡觉、无限屏幕时间、取消必要规则等项目。</p></div>
    <div class="speech-card"><span>兑换要复盘</span><p>兑换时可以回看孩子是通过哪些努力获得的，让奖励连接到具体行为。</p></div>
  </div></div>`;
}

function rewardPresetGrid() {
  const presets = [
    ['科普书一本','BOOK','NONE',12,0,1,2,'适合作为长期阅读兴趣奖励。'],
    ['周末亲子公园','ACTIVITY','NONE',18,0,1,2,'用兑换分解锁一次家庭户外活动。'],
    ['家庭电影夜','ACTIVITY','LOW',15,0,1,2,'周末固定时段，全家一起观看并分享感受。'],
    ['选择晚餐一道菜','PRIVILEGE','NONE',8,0,1,4,'在健康范围内选择一道家庭晚餐菜品。'],
    ['小零食一份','SNACK','MEDIUM',6,0,1,3,'限定份量，建议饭后或周末兑换。'],
    ['长期大奖：博物馆日','ACTIVITY','NONE',0,5,1,1,'使用星星兑换一次更有仪式感的家庭活动。'],
  ];
  return `<div class="practice-grid compact">${presets.map(p => `<button type="button" class="practice-card" data-reward-name="${h(p[0])}" data-reward-type="${p[1]}" data-reward-risk="${p[2]}" data-reward-score="${p[3]}" data-reward-star="${p[4]}" data-reward-weekly="${p[5]}" data-reward-monthly="${p[6]}" data-reward-description="${h(p[7])}"><span>${rewardIcon(p[1])}</span><b>${h(p[0])}</b><small>${h(rewardTypeName(p[1]))} · ${p[3] ? `${p[3]}分` : `${p[4]}星`}</small></button>`).join('')}</div>`;
}

