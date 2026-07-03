const WISHLIST_PREFIX = 'fs_wishlist:';

function wishlistKey(childId = state.childId) {
  return `${WISHLIST_PREFIX}${childId || 'none'}`;
}

function getWishlist(childId = state.childId) {
  if (!childId) return [];
  try {
    const raw = localStorage.getItem(wishlistKey(childId));
    const list = JSON.parse(raw || '[]');
    return Array.isArray(list) ? list.map(Number).filter(Boolean) : [];
  } catch {
    return [];
  }
}

function setWishlist(list, childId = state.childId) {
  if (!childId) return;
  const normalized = [...new Set((list || []).map(Number).filter(Boolean))];
  try {
    localStorage.setItem(wishlistKey(childId), JSON.stringify(normalized));
  } catch {
    // 本地存储不可用时静默降级，不影响兑换主流程。
  }
}

function isWishlisted(childId, rewardId) {
  return getWishlist(childId).includes(Number(rewardId));
}

function toggleWishlist(childId, rewardId) {
  const id = Number(rewardId);
  const list = getWishlist(childId);
  const next = list.includes(id) ? list.filter(x => x !== id) : [...list, id];
  setWishlist(next, childId);
  return next.includes(id);
}

function wishlistRewards(childId = state.childId) {
  const ids = getWishlist(childId);
  const existing = state.rewards.filter(r => ids.includes(Number(r.id)));
  if (existing.length !== ids.length) setWishlist(existing.map(r => r.id), childId);
  return existing;
}

function wishlistCount(childId = state.childId) {
  return wishlistRewards(childId).length;
}

function renderWishlistPanel(childId = state.childId) {
  const wishes = wishlistRewards(childId);
  if (!childId) return '<div class="empty-state"><div>🌟</div><b>先选择一个孩子</b><p>选择孩子后，就可以记录 TA 想要的奖励。</p></div>';
  if (!wishes.length) return '<div class="empty-state"><div>🌟</div><b>愿望单还是空的</b><p>看到喜欢的奖励，可以先点“加入愿望单”。</p></div>';
  return `<div class="wishlist-list">${wishes.map(r => `<div class="wishlist-item">
    <div class="task-main"><div class="task-icon">${rewardIcon(r.rewardType)}</div><div><b>${h(r.rewardName)}</b><div class="small">${h(rewardCostText(r))} · ${h(riskName(r.healthRisk))}</div></div></div>
    <div class="task-side"><button data-reward="${r.id}">申请兑换</button><button class="secondary" data-wishlist="${r.id}">移出愿望单</button></div>
  </div>`).join('')}</div>`;
}
