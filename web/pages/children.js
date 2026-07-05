function renderChildren() {
  if (!canOperate()) return '<div class="card"><h2>无权限</h2><div class="notice">只有家长或管理员可以添加孩子。</div></div>';
  const parentField = `<div class="form two"><div class="field"><label>代表家长</label><select name="parentUserId"><option value="0">不指定</option>${parentOptionsHTML(0)}</select></div><div class="field"><label>监护组</label><input name="parentGroup" list="parentGroupList" value="${h(defaultParentGroup())}" placeholder="如：小明监护组"><div class="small">同一监护组内的爸爸、妈妈、爷爷、奶奶、姥姥、姥爷都可以绑定并访问同一个孩子。</div>${parentGroupDatalist()}</div></div>`;
  return `<div class="stack"><div class="split">
    <div class="card"><h2>添加孩子</h2><form class="form" id="childForm">
      <div class="field"><label>孩子姓名</label><input name="name" placeholder="如：小明" required></div>
      <div class="field"><label>年龄</label><input name="age" type="number" min="1" value="8" required></div>
      <div class="field"><label>性别</label><select name="gender"><option value="BOY">男孩</option><option value="GIRL">女孩</option></select></div>
      ${parentField}
      <div class="field"><label>孩子登录名</label><input name="childLoginName" placeholder="选填，如 xiaoming"><div class="small">填写后会同步创建孩子账号。</div></div>
      <div class="field"><label>孩子登录密码</label><input name="childPassword" type="password" placeholder="选填，至少 4 位"></div>
      <button>新增孩子</button>
    </form></div>
    <div class="card"><h2>孩子列表</h2><div class="notice">可在这里把孩子绑定到某个监护组；同组家长账号都能看到并操作该孩子。</div>${state.children.length ? `<table class="table"><thead><tr><th>姓名</th><th>年龄</th><th>性别</th><th>绑定状态</th><th>操作</th></tr></thead><tbody>${state.children.map(renderChildRow).join('')}</tbody></table>` : '<div class="empty">暂无孩子档案</div>'}</div>
  </div>${guardianGroupManager()}</div>`;
}

function renderChildRow(c) {
  return `<tr>
    <td><input data-child-name="${c.id}" value="${h(c.name)}"></td>
    <td><input data-child-age="${c.id}" type="number" min="1" value="${h(c.age)}"></td>
    <td><select data-child-gender="${c.id}"><option value="BOY" ${c.gender === 'BOY' ? 'selected' : ''}>男孩</option><option value="GIRL" ${c.gender === 'GIRL' ? 'selected' : ''}>女孩</option></select></td>
    <td><select data-child-parent="${c.id}"><option value="0">不指定</option>${parentOptionsHTML(c.parentUserId)}</select><input data-child-parent-group="${c.id}" list="parentGroupList" value="${h(c.parentGroup || '')}" placeholder="监护组"><div class="small">${childBindingText(c)}</div></td>
    <td class="row"><button data-save-child="${c.id}">保存</button><button class="danger" data-del-child="${c.id}">删除</button></td>
  </tr>`;
}

function guardianGroupManager() {
  const groups = state.guardianGroups || [];
  const form = isAdmin() ? `<form class="form" id="guardianGroupForm"><div class="form two"><div class="field"><label>监护组名称</label><input name="name" placeholder="如：小明监护组" required></div><div class="field"><label>说明</label><input name="description" placeholder="选填，如：小明的主要监护人"></div></div><button>新增监护组</button></form>` : '<div class="notice">家长可查看自己的监护组；新增、改名、删除由管理员在这里操作。</div>';
  const table = groups.length ? `<table class="table"><thead><tr><th>名称</th><th>说明</th><th>绑定家长</th><th>绑定孩子</th><th>操作</th></tr></thead><tbody>${groups.map(g => `<tr><td><input data-group-name="${g.id}" value="${h(g.name)}" ${isAdmin() ? '' : 'disabled'}></td><td><input data-group-desc="${g.id}" value="${h(g.description || '')}" ${isAdmin() ? '' : 'disabled'}></td><td>${h(g.parentCount || 0)}</td><td>${h(g.childCount || 0)}</td><td class="row">${isAdmin() ? `<button data-save-group="${g.id}">保存</button><button class="danger" data-del-group="${g.id}">删除</button>` : '<span class="small">只读</span>'}</td></tr>`).join('')}</tbody></table>` : '<div class="empty-state"><div>👪</div><b>暂无监护组</b><p>先创建一个监护组，再把爸爸、妈妈等家长和孩子绑定进来。</p></div>';
  return `<div class="card"><div class="section-title"><div><h2>监护组管理</h2><p class="small">监护组用于把爸爸、妈妈、爷爷、奶奶、姥姥、姥爷等账号归为同一组，并共同绑定孩子。</p></div><span class="tag">${groups.length} 个</span></div>${form}${table}</div>`;
}

function parentOptionsHTML(selectedId) {
  const parents = state.users.filter(u => u.enabled && u.role === 'PARENT');
  const list = parents.length ? parents : [{id: state.me?.userId || 0, name: state.me?.name || '当前家长', loginName: '', parentTitle: '', parentGroup: defaultParentGroup()}];
  return list.map(u => `<option value="${u.id}" ${u.id === selectedId ? 'selected' : ''}>${h(u.name)}${u.parentTitle ? `（${h(u.parentTitle)}）` : ''}${u.parentGroup ? ` · ${h(u.parentGroup)}` : ''}</option>`).join('');
}

function parentGroupDatalist() {
  return `<datalist id="parentGroupList">${parentGroupOptions().map(g => `<option value="${h(g)}"></option>`).join('')}</datalist>`;
}

function parentGroupOptions() {
  const groups = (state.guardianGroups || []).map(g => g.name);
  state.users.filter(u => u.enabled && u.role === 'PARENT' && u.parentGroup).forEach(u => groups.push(u.parentGroup));
  state.children.filter(c => c.parentGroup).forEach(c => groups.push(c.parentGroup));
  if (state.me?.parentGroup) groups.push(state.me.parentGroup);
  groups.push('默认监护组');
  return [...new Set(groups.filter(Boolean))];
}

function defaultParentGroup() {
  return state.me?.parentGroup || parentGroupOptions()[0] || '默认监护组';
}

function childBindingText(c) {
  const group = c.parentGroup || '';
  const parent = parentNameById(c.parentUserId);
  if (!group && !c.parentUserId) return '未绑定家长或监护组';
  return `已绑定：${group ? `监护组 ${group}` : ''}${parent && parent !== '-' ? ` / 代表家长 ${parent}` : ''}`;
}

function parentNameById(parentUserId) {
  if (!parentUserId) return '';
  const parent = state.users.find(u => u.id === parentUserId);
  return parent ? parent.name : '-';
}
