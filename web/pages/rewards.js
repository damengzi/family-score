function renderRewards() {
  const pending = state.exchangeOrders.filter(o => o.status === 'PENDING');
  const rewardCards = state.rewards.map(r => `<div class="card"><h3>${h(r.rewardName)}</h3><p class="small">${h(r.description)}</p><div class="row"><span class="tag">${r.costScore ? `${r.costScore} 兑换分` : `${r.costStar} 星`}</span><span class="tag ${r.healthRisk==='NONE'?'green':'yellow'}">${h(r.healthRisk)}</span></div><br><button data-reward="${r.id}">申请兑换</button></div>`).join('');
  const auditPanel = canOperate() ? `<br><div class="card"><h2>待审核兑换</h2>${pending.length ? `<table class="table"><tbody>${pending.map(o => `<tr><td>${h(o.rewardName)}<div class="small">${h(o.appliedAt)} · ${o.costScore}分 / ${o.costStar}星</div></td><td><button data-audit-order="${o.id}">通过</button><button class="secondary" data-reject-order="${o.id}">驳回</button></td></tr>`).join('')}</tbody></table>` : '<div class="empty">暂无待审核兑换</div>'}</div>` : '';
  return `<div class="grid two">${rewardCards}</div>${auditPanel}`;
}

function renderRewardConfig() {
  return `<div class="split">
    <div class="card"><h2>奖励/零食自定义添加</h2><form class="form" id="rewardForm">
      <div class="field"><label>奖励名称</label><input name="rewardName" placeholder="如：薯片一包 / 奶茶小杯 / 科普书" required></div>
      <div class="form two"><div class="field"><label>类型</label><select name="rewardType"><option value="SNACK">零食</option><option value="DRINK">饮品</option><option value="BOOK">图书</option><option value="TOY">玩具</option><option value="ACTIVITY">活动</option><option value="PRIVILEGE">特权</option></select></div><div class="field"><label>健康风险</label><select name="healthRisk"><option value="NONE">无</option><option value="LOW">低</option><option value="MEDIUM">中</option><option value="HIGH">高/禁用</option></select></div></div>
      <div class="form two"><div class="field"><label>消耗兑换分</label><input name="costScore" type="number" min="0" value="5"></div><div class="field"><label>消耗星星</label><input name="costStar" type="number" min="0" value="0"></div></div>
      <div class="form two"><div class="field"><label>周限制</label><input name="weeklyLimit" type="number" min="1" value="1"></div><div class="field"><label>月限制</label><input name="monthlyLimit" type="number" min="1" value="4"></div></div>
      <div class="field"><label>说明</label><textarea name="description"></textarea></div><button>新增奖励</button>
    </form></div>
    <div class="card"><h2>奖励配置</h2>${state.rewards.length ? `<table class="table"><tbody>${state.rewards.map(r => `<tr><td><b>${h(r.rewardName)}</b><div class="small">${h(r.rewardType)} · ${r.costScore}分 · ${r.costStar}星 · ${h(r.healthRisk)} · ${h(r.description)}</div></td><td><button class="danger" data-del-reward="${r.id}">删除</button></td></tr>`).join('')}</tbody></table>` : '<div class="empty">暂无奖励</div>'}</div>
  </div>`;
}
