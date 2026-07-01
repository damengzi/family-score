function renderSystem() {
  return `<div class="grid two"><div class="card"><h2>本机数据</h2><p>数据目录：<code>${h(state.status.dataDir)}</code></p><p>数据库：<code>${h(state.status.dbPath)}</code></p><button id="backupBtn">立即备份</button></div><div class="card"><h2>运行提示</h2><div class="notice">请定期备份 SQLite 数据库。当前版本默认仅监听本机地址，不会上传云端。</div></div></div>`;
}
