function renderChildren() {
  if (!canOperate()) return '<div class="card"><h2>无权限</h2><div class="notice">只有家长或管理员可以添加孩子。</div></div>';
  const parentOptions = state.users.filter(u => u.enabled && u.role === 'PARENT').map(u => `<option value="${u.id}">${h(u.name)}（${h(u.loginName)}）</option>`).join('');
  const parentField = isAdmin() ? `<div class="field"><label>归属家长</label><select name="parentUserId"><option value="0">请选择家长</option>${parentOptions}</select></div>` : '';
  return `<div class="split">
    <div class="card"><h2>添加孩子</h2><form class="form" id="childForm">
      <div class="field"><label>孩子姓名</label><input name="name" placeholder="如：小明" required></div>
      <div class="field"><label>年龄</label><input name="age" type="number" min="1" value="8" required></div>
      <div class="field"><label>性别</label><select name="gender"><option value="BOY">男孩</option><option value="GIRL">女孩</option></select></div>
      ${parentField}
      <button>新增孩子</button>
    </form></div>
    <div class="card"><h2>孩子列表</h2>${state.children.length ? `<table class="table"><thead><tr><th>姓名</th><th>年龄</th><th>性别</th><th>归属家长</th></tr></thead><tbody>${state.children.map(c => `<tr><td>${h(c.name)}</td><td>${h(c.age)}</td><td>${c.gender === 'GIRL' ? '女孩' : '男孩'}</td><td>${h(parentNameById(c.parentUserId))}</td></tr>`).join('')}</tbody></table>` : '<div class="empty">暂无孩子档案</div>'}</div>
  </div>`;
}

function parentNameById(parentUserId) {
  if (!parentUserId) return isAdmin() ? '未分配' : state.me?.name || '-';
  const parent = state.users.find(u => u.id === parentUserId);
  return parent ? parent.name : '-';
}
