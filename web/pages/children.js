function renderChildren() {
  if (!canOperate()) return '<div class="card"><h2>无权限</h2><div class="notice">只有家长或管理员可以添加孩子。</div></div>';
  const parentField = `<div class="field"><label>归属家长</label><select name="parentUserId"><option value="0">请选择家长</option>${parentOptionsHTML(0)}</select></div>`;
  return `<div class="split">
    <div class="card"><h2>添加孩子</h2><form class="form" id="childForm">
      <div class="field"><label>孩子姓名</label><input name="name" placeholder="如：小明" required></div>
      <div class="field"><label>年龄</label><input name="age" type="number" min="1" value="8" required></div>
      <div class="field"><label>性别</label><select name="gender"><option value="BOY">男孩</option><option value="GIRL">女孩</option></select></div>
      ${parentField}
      <div class="field"><label>孩子登录名</label><input name="childLoginName" placeholder="选填，如 xiaoming"><div class="small">填写后会同步创建孩子账号。</div></div>
      <div class="field"><label>孩子登录密码</label><input name="childPassword" type="password" placeholder="选填，至少 4 位"></div>
      <button>新增孩子</button>
    </form></div>
    <div class="card"><h2>孩子列表</h2>${state.children.length ? `<table class="table"><thead><tr><th>姓名</th><th>年龄</th><th>性别</th><th>归属家长</th><th>操作</th></tr></thead><tbody>${state.children.map(renderChildRow).join('')}</tbody></table>` : '<div class="empty">暂无孩子档案</div>'}</div>
  </div>`;
}

function renderChildRow(c) {
  return `<tr>
    <td><input data-child-name="${c.id}" value="${h(c.name)}"></td>
    <td><input data-child-age="${c.id}" type="number" min="1" value="${h(c.age)}"></td>
    <td><select data-child-gender="${c.id}"><option value="BOY" ${c.gender === 'BOY' ? 'selected' : ''}>男孩</option><option value="GIRL" ${c.gender === 'GIRL' ? 'selected' : ''}>女孩</option></select></td>
    <td><select data-child-parent="${c.id}">${parentOptionsHTML(c.parentUserId)}</select></td>
    <td class="row"><button data-save-child="${c.id}">保存</button><button class="danger" data-del-child="${c.id}">删除</button></td>
  </tr>`;
}

function parentOptionsHTML(selectedId) {
  const parents = state.users.filter(u => u.enabled && u.role === 'PARENT');
  const list = parents.length ? parents : [{id: state.me?.userId || 0, name: state.me?.name || '当前家长', loginName: ''}];
  return list.map(u => `<option value="${u.id}" ${u.id === selectedId ? 'selected' : ''}>${h(u.name)}${u.loginName ? `（${h(u.loginName)}）` : ''}</option>`).join('');
}

function parentNameById(parentUserId) {
  if (!parentUserId) return isAdmin() ? '未分配' : state.me?.name || '-';
  const parent = state.users.find(u => u.id === parentUserId);
  return parent ? parent.name : '-';
}
