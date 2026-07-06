function renderFamilyGroups() {
  if (!isAdmin()) return '<div class="card"><h2>无权限</h2><div class="notice">只有管理员可以管理家庭组。</div></div>';
  const groups = state.guardianGroups || [];
  const current = currentFamilyGroup();
  return `<div class="stack">
    <div class="card"><div class="section-title"><div><h2>家庭组管理</h2><p class="small">家庭组可以自定义名称，例如“小明小红家”“张家”。把爸爸、妈妈、爷爷、奶奶、姥姥、姥爷和孩子加入同一组后，可共同围绕孩子进行家庭协作。</p></div><span class="tag">${groups.length} 个家庭组</span></div>
      <form class="form" id="familyGroupForm"><div class="form two"><div class="field"><label>家庭组名称</label><input name="name" placeholder="如：小明小红家" required></div><div class="field"><label>说明</label><input name="description" placeholder="如：小明、小红的共同监护家庭组"></div></div><button>新增家庭组</button></form>
    </div>
    <div class="split">
      <div class="card"><h2>家庭组列表</h2>${groups.length ? `<div class="group-list">${groups.map(g => familyGroupListItem(g, current)).join('')}</div>` : '<div class="empty-state"><div>👪</div><b>暂无家庭组</b><p>先创建一个家庭组，如“小明小红家”。</p></div>'}</div>
      <div class="card"><h2>组成员</h2>${current ? familyGroupMembersPanel(current) : '<div class="empty-state"><div>👨‍👩‍👧‍👦</div><b>请选择家庭组</b><p>创建或选择家庭组后，可以添加和移除组成员。</p></div>'}</div>
    </div>
  </div>`;
}

function currentFamilyGroup() {
  const groups = state.guardianGroups || [];
  if (!groups.length) return null;
  const found = groups.find(g => g.name === state.familyGroupName);
  return found || groups[0];
}

function familyGroupListItem(g, current) {
  const active = current && current.id === g.id;
  return `<div class="group-list-item ${active ? 'active' : ''}">
    <button type="button" class="secondary" data-select-family-group="${h(g.name)}">${h(g.name)}</button>
    <div class="form two compact"><div class="field"><label>名称</label><input data-family-group-name="${g.id}" value="${h(g.name)}"></div><div class="field"><label>说明</label><input data-family-group-desc="${g.id}" value="${h(g.description || '')}"></div></div>
    <div class="row"><span class="tag">家长 ${h(g.parentCount || 0)}</span><span class="tag">孩子 ${h(g.childCount || 0)}</span><button data-save-family-group="${g.id}">保存</button><button class="danger" data-del-family-group="${g.id}">删除</button></div>
  </div>`;
}

function familyGroupMembersPanel(group) {
  const name = group.name;
  const parents = state.users.filter(u => u.enabled && u.role === 'PARENT');
  const inParents = parents.filter(u => u.parentGroup === name);
  const outParents = parents.filter(u => u.parentGroup !== name);
  const inChildren = state.children.filter(c => c.parentGroup === name);
  const outChildren = state.children.filter(c => c.parentGroup !== name);
  return `<div class="stack">
    <div class="notice">当前家庭组：<b>${h(name)}</b>。添加/移除成员会同步修改家长账号或孩子档案的家庭组绑定。</div>
    ${memberSection('家长成员', inParents, outParents, name, renderParentMember, renderParentCandidate)}
    ${memberSection('孩子成员', inChildren, outChildren, name, renderChildMember, renderChildCandidate)}
  </div>`;
}

function memberSection(title, inItems, outItems, groupName, renderIn, renderOut) {
  return `<div class="member-section"><div class="section-title"><h3>${title}</h3><span class="tag">${inItems.length} 个</span></div>
    <div class="member-grid"><div><b>已在组内</b>${inItems.length ? inItems.map(x => renderIn(x, groupName)).join('') : '<div class="empty">暂无成员</div>'}</div>
    <div><b>可添加</b>${outItems.length ? outItems.map(x => renderOut(x, groupName)).join('') : '<div class="empty">暂无可添加成员</div>'}</div></div>
  </div>`;
}

function renderParentMember(u, groupName) {
  return `<div class="member-card"><span>👤 ${h(u.name)}（${h(u.parentTitle || '家长')}）<small>${h(u.loginName)}</small></span><button class="secondary" data-remove-parent-group="${u.id}" data-group-name="${h(groupName)}">移出</button></div>`;
}

function renderParentCandidate(u, groupName) {
  return `<div class="member-card"><span>👤 ${h(u.name)}（${h(u.parentTitle || '家长')}）<small>${h(u.parentGroup || '未加入家庭组')}</small></span><button data-add-parent-group="${u.id}" data-group-name="${h(groupName)}">加入</button></div>`;
}

function renderChildMember(c, groupName) {
  return `<div class="member-card"><span>🧒 ${h(c.name)}<small>${h(c.age)}岁 · ${c.gender === 'GIRL' ? '女孩' : '男孩'}</small></span><button class="secondary" data-remove-child-group="${c.id}" data-group-name="${h(groupName)}">移出</button></div>`;
}

function renderChildCandidate(c, groupName) {
  return `<div class="member-card"><span>🧒 ${h(c.name)}<small>${h(c.parentGroup || '未加入家庭组')}</small></span><button data-add-child-group="${c.id}" data-group-name="${h(groupName)}">加入</button></div>`;
}

async function setParentFamilyGroup(userId, groupName) {
  const user = state.users.find(u => Number(u.id) === Number(userId));
  if (!user) throw new Error('家长不存在');
  await api(`/api/users/${user.id}`, {method:'PATCH', body:{displayName:user.name, parentTitle:user.parentTitle || '家长', parentGroup:groupName || '', password:''}});
}

async function setChildFamilyGroup(childId, groupName) {
  const child = state.children.find(c => Number(c.id) === Number(childId));
  if (!child) throw new Error('孩子不存在');
  await api(`/api/children/${child.id}`, {method:'PATCH', body:{name:child.name, age:Number(child.age), gender:child.gender || 'BOY', parentUserId: groupName ? Number(child.parentUserId || 0) : 0, parentGroup:groupName || ''}});
}
