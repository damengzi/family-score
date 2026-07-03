function renderScoreGuide() {
  return `<div class="stack">
    <div class="card"><h2>分值说明</h2><div class="notice">本系统采用“基准德育分 + 兑换分 + 星星 + 家庭小队分”四轨账户。基准分反映行为状态，兑换分用于兑换奖励，星星用于长期大奖，家庭小队分用于家庭集体福利。</div></div>
    <div class="grid two">
      <div class="card"><h3>基础分</h3><p>每个孩子每月默认 <b>100 分</b>。基础分用于衡量当月行为状态，不直接兑换物品。</p><ul><li>扣分只扣基础分。</li><li>修复任务优先恢复基础分。</li><li>低于 90 分暂停高价值兑换。</li><li>低于 80 分暂停全部兑换。</li></ul></div>
      <div class="card"><h3>兑换分</h3><p>基础分满 100 后，优秀表现产生的加分进入兑换分。</p><ul><li>兑换分可兑换零食、文具、活动和特权。</li><li>月末可按 <b>10 兑换分 = 1 星</b> 转换。</li><li>兑换前需要满足基础分状态要求。</li></ul></div>
      <div class="card"><h3>星星</h3><p>星星用于长期奖励，不建议频繁兑换普通零食。</p><ul><li>月末兑换分可转星星。</li><li>重大进步、连续习惯稳定可额外给星星。</li><li>星星上限为 <b>20 颗</b>。</li></ul></div>
      <div class="card"><h3>家庭小队分</h3><p>家庭小队分用于鼓励合作，不扣个人分，不直接消耗。</p><ul><li>家庭阅读、运动、家务协作可获得。</li><li>月度按小队分评等级。</li><li>可解锁家庭活动类福利。</li></ul></div>
    </div>
    <div class="card"><h2>加减分规则</h2><table class="table"><thead><tr><th>类型</th><th>规则</th><th>说明</th></tr></thead><tbody>
      <tr><td>加分</td><td>基础分未满 100 时先恢复基础分；满 100 后进入兑换分。</td><td>鼓励主动行为。</td></tr>
      <tr><td>扣分</td><td>只扣基础分。</td><td>扣分需说明原因，不做情绪化扣分。</td></tr>
      <tr><td>惩罚/修复</td><td>修复任务只恢复基础分，不进入兑换分。</td><td>用于补救错误行为。</td></tr>
      <tr><td>任务加分</td><td>孩子完成任务后需家长审核，通过后才加分。</td><td>同一孩子当天任务审核加分总额不超过 <b>15 分</b>。</td></tr>
    </tbody></table></div>
    <div class="card"><h2>兑换物品规则</h2><ul><li>只能使用兑换分或星星兑换。</li><li>基础分低于 90 分暂停高价值兑换，低于 80 分暂停全部兑换。</li><li>奖励可设置周限制、月限制和健康风险。</li><li>高咖啡因饮料、无限量零食、延迟睡觉、免作业类奖励不建议配置。</li><li>孩子账号可自主申请兑换，家长/管理员审核后生效。</li></ul></div>
    <div class="grid two">
      <div class="card"><h2>家庭场景案例</h2><div class="practice-list">
        <div><b>学习自驱</b><p>主动开始作业、阅读后分享、错题订正完整，可给 1-3 分；重点记录“主动”和“坚持”。</p></div>
        <div><b>生活自理</b><p>整理书包、按时洗漱、自己准备第二天衣物，可给 1 分；连续稳定再考虑星星。</p></div>
        <div><b>家务协作</b><p>饭后收拾、扫地、晾衣服、照顾宠物，优先计入家庭小队分，强调“我是家庭成员”。</p></div>
        <div><b>情绪修复</b><p>顶撞、拖延、摔东西后，不只扣分；要设计道歉、复盘、整理现场等修复任务。</p></div>
      </div></div>
      <div class="card"><h2>修复闭环</h2><ol class="principle-list"><li><b>描述事实：</b>今天发生了什么，不贴标签。</li><li><b>确认影响：</b>这件事影响了谁、影响了什么。</li><li><b>选择修复：</b>让孩子从 2-3 个可完成任务里选一个。</li><li><b>完成确认：</b>家长确认后恢复基准分，不翻旧账。</li></ol><div class="notice">修复任务建议控制在 5-15 分钟，目标是恢复秩序和责任感，不是制造额外惩罚。</div></div>
    </div>
    <div class="card"><h2>亲子沟通话术</h2><div class="speech-grid">
      <div class="speech-card"><span>加分时</span><p>“我注意到你没有提醒就开始整理书包，这说明你在练习对自己的事情负责。”</p></div>
      <div class="speech-card"><span>扣分时</span><p>“这次扣的是拖延这个行为，不是说你不好。我们一起看下次怎么提前开始。”</p></div>
      <div class="speech-card"><span>修复时</span><p>“事情已经发生了，重要的是怎么补救。你想选择整理现场，还是写下下次做法？”</p></div>
      <div class="speech-card"><span>兑换时</span><p>“你通过努力攒到了兑换分。我们也看看健康风险和本周次数，再决定什么时候兑换。”</p></div>
    </div></div>
  </div>`;
}
