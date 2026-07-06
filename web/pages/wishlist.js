function wishMarker(rewardId) {
  return `奖励ID:${Number(rewardId)}`;
}

function rewardWish(childId, rewardId) {
  return (state.wishes || []).find(w => Number(w.childId) === Number(childId) && String(w.reason || '').includes(wishMarker(rewardId)) && w.status !== 'REJECTED');
}

function isWishlisted(childId, rewardId) {
  return Boolean(rewardWish(childId, rewardId));
}

function wishlistRewards(childId = state.childId) {
  if (!childId) return [];
  return state.rewards.filter(r => isWishlisted(childId, r.id));
}

function wishlistCount(childId = state.childId) {
  return wishlistRewards(childId).length;
}

function renderWishlistPanel(childId = state.childId) {
  const wishes = (state.wishes || []).filter(w => Number(w.childId) === Number(childId));
  if (!childId) return '<div class="empty-state"><div>🌟</div><b>先选择一个孩子</b><p>选择孩子后，就可以提交 TA 的愿望。</p></div>';
  if (!wishes.length) return '<div class="empty-state"><div>🌟</div><b>愿望还是空的</b><p>看到喜欢的奖励，可以提交给家长审批。</p></div>';
  return `<div class="wishlist-list">${wishes.map(w => `<div class="wishlist-item">
    <div class="task-main"><div class="task-icon">🌟</div><div><b>${h(w.wishName)}</b><div class="small">${h(wishStatusText(w.status))} · ${w.expectedScore || 0}分 / ${w.expectedStar || 0}星</div>${w.auditNote ? `<p class="small">家长意见：${h(w.auditNote)}</p>` : ''}</div></div>
  </div>`).join('')}</div>`;
}
