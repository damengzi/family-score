function renderUsers() {
  if (!isAdmin()) return '<div class="card"><h2>无权限</h2><div class="notice">只有管理员可以管理用户。</div></div>';
  const childOptions = state.children.map(c => `<option value="${c.id}">${h(c.name)} ${c.age}岁</option>`).join('');
  return `<div class="split">
    <div class="card"><h2>注册用户</h2><form class="form" id="userForm">
      <div class="field"><label>显示名称</label><input name="displayName" placeholder="如：爸爸 / 妈妈 / 孩子" required></div>
      <div class="field"><label>登录名</label><input name="loginName" placeholder="用于登录" required></div>
      <div class="field"><label>密码</label><input name="password" type="password" value="123456" required></div>
      <div class="field"><label>角色</label><select name="role" id="userRole"><option value="PARENT">家长</option><option value="CHILD">孩子</option></select><div class="small">管理员账号固定为 admin / 654321，其他账号只分家长和孩子。</div></div>
      <div class="field" id="bindChildField" style="display:none"><label>绑定孩子档案</label><select name="childId"><option value="0">请选择</option>${childOptions}</select><div class="small">孩子账号只能查看并操作绑定孩子的日常任务和兑换申请。</div></div>
      <button>新增用户</button>
    </form></div>
    <div class="card"><h2>用户列表</h2>${state.users.length ? `<table class="table"><thead><tr><th>名称</th><th>登录名</th><th>角色</th><th>绑定孩子</th><th>状态</th><th>操作</th></tr></thead><tbody>${state.users.map(u => `<tr><td>${h(u.name)}</td><td>${h(u.loginName)}</td><td>${roleName(u.role)}</td><td>${h(childNameById(u.childId))}</td><td>${u.enabled ? '启用' : '已注销'}</td><td>${u.enabled && u.id !== state.me.userId ? `<button class="danger" data-del-user="${u.id}">注销</button>` : '<span class="small">-</span>'}</td></tr>`).join('')}</tbody></table>` : '<div class="empty">暂无用户</div>'}</div>
  </div>`;
}

function childNameById(childId) {
  const child = state.children.find(c => c.id === childId);
  return child ? child.name : '-';
}
