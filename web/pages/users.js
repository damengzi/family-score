function renderUsers() {
  if (!isAdmin()) return '<div class="card"><h2>无权限</h2><div class="notice">只有管理员可以管理用户。</div></div>';
  return `<div class="split">
    <div class="card"><h2>注册用户</h2><form class="form" id="userForm">
      <div class="field"><label>显示名称</label><input name="displayName" placeholder="如：爸爸 / 妈妈 / 小明" required></div>
      <div class="field"><label>登录名</label><input name="loginName" placeholder="用于登录" required></div>
      <div class="field"><label>密码</label><input name="password" type="password" value="123456" required></div>
      <div class="field"><label>角色</label><select name="role" id="userRole"><option value="PARENT">家长</option><option value="CHILD">孩子</option></select><div class="small">管理员账号固定创建；家长可选择爸爸、妈妈、爷爷、奶奶、姥姥、姥爷，并按同一个监护组绑定孩子。</div></div>
      <div id="parentBindField">
        <div class="form two"><div class="field"><label>家长称谓</label><select name="parentTitle">${parentTitleOptionsHTML('爸爸')}</select></div><div class="field"><label>监护组</label><input name="parentGroup" value="默认监护组" placeholder="如：小明监护组"><div class="small">同组内爸爸、妈妈、爷爷等都可访问绑定到该组的孩子。</div></div></div>
      </div>
      <div class="field" id="bindChildField" style="display:none"><label>绑定孩子档案</label><select name="childId"><option value="0">请选择</option>${childOptionsHTML(0)}</select><div class="small">孩子用这里设置的登录名和密码登录；密码创建后无法查看明文，忘记时可在下方重置。</div></div>
      <button>新增用户</button>
    </form></div>
    <div class="card"><h2>用户列表</h2><div class="notice">家长通过“监护组”绑定孩子；孩子账号通过“绑定孩子档案”绑定。已有密码不会明文展示，可在“新密码”中填写后保存完成重置。</div>${state.users.length ? `<table class="table"><thead><tr><th>名称</th><th>登录名</th><th>角色</th><th>绑定关系</th><th>新密码</th><th>状态</th><th>操作</th></tr></thead><tbody>${state.users.map(renderUserRow).join('')}</tbody></table>` : '<div class="empty">暂无用户</div>'}</div>
  </div>`;
}

function renderUserRow(u) {
  const binding = userBindingHTML(u);
  const actions = u.enabled
    ? `<button data-save-user="${u.id}">保存</button>${u.id !== state.me.userId ? `<button class="danger" data-del-user="${u.id}">注销</button>` : ''}`
    : '<span class="small">已注销</span>';
  return `<tr>
    <td><input data-user-name="${u.id}" value="${h(u.name)}" ${u.enabled ? '' : 'disabled'}></td>
    <td>${h(u.loginName)}</td>
    <td>${roleName(u.role)}${u.role === 'PARENT' ? ` · ${h(u.parentTitle || '家长')}` : ''}</td>
    <td>${binding}</td>
    <td><input data-user-pass="${u.id}" type="password" placeholder="留空不改" ${u.enabled ? '' : 'disabled'}></td>
    <td>${u.enabled ? '启用' : '已注销'}</td>
    <td class="row">${actions}</td>
  </tr>`;
}

function userBindingHTML(u) {
  if (u.role === 'CHILD') return `<select data-user-child="${u.id}"><option value="0">请选择</option>${childOptionsHTML(u.childId)}</select><div class="small">${u.childId ? `已绑定：${h(childNameById(u.childId))}` : '未绑定孩子'}</div>`;
  if (u.role === 'PARENT') return `<div class="form two compact"><div class="field"><select data-user-parent-title="${u.id}">${parentTitleOptionsHTML(u.parentTitle || '爸爸')}</select></div><div class="field"><input data-user-parent-group="${u.id}" value="${h(u.parentGroup || '默认监护组')}" placeholder="监护组"></div></div><div class="small">${h(u.parentGroup || '未绑定监护组')}</div>`;
  return '<span class="small">-</span>';
}

function childOptionsHTML(selectedId) {
  return state.children.map(c => `<option value="${c.id}" ${c.id === selectedId ? 'selected' : ''}>${h(c.name)} ${c.age}岁</option>`).join('');
}

function parentTitleOptionsHTML(selected) {
  return ['爸爸','妈妈','爷爷','奶奶','姥姥','姥爷'].map(x => `<option value="${x}" ${x === selected ? 'selected' : ''}>${x}</option>`).join('');
}

function childNameById(childId) {
  const child = state.children.find(c => c.id === childId);
  return child ? child.name : '-';
}
