function renderScoreForm() {
  return `<div class="split">
    <div class="card"><h2>加分 / 扣分 / 惩罚修复</h2><form class="form" id="scoreForm">
      <div class="field"><label>类型</label><select name="recordType"><option value="ADD">加分</option><option value="DEDUCT">扣分</option><option value="REPAIR">惩罚/修复加回基准分</option><option value="TEAM">家庭小队分</option><option value="STAR">星星</option></select></div>
      <div class="field"><label>项目名称</label><input name="itemName" placeholder="如：主动整理书包 / 作业拖延 / 整理书桌修复" required></div>
      <div class="field"><label>分值</label><input name="scoreChange" type="number" value="1" min="1" required></div>
      <div class="field"><label>原因</label><textarea name="reason" placeholder="简要说明"></textarea></div>
      <button>提交记录</button>
    </form></div>
    <div class="card"><h2>说明</h2><div class="notice">扣分只扣基准分；修复任务只恢复基准分；基准分满 100 后，加分进入超额兑换分。</div></div>
  </div>`;
}
