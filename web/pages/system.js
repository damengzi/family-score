function renderSystem() {
  return `<div class="grid two"><div class="card"><h2>本机数据</h2><p>数据目录：<code>${h(state.status.dataDir)}</code></p><p>数据库：<code>${h(state.status.dbPath)}</code></p><button id="backupBtn">立即备份</button></div><div class="card"><h2>家庭共享访问</h2><div id="networkInfoBox"><div class="notice">局域网信息加载中...</div></div></div><div class="card"><h2>运行提示</h2><div class="notice">请定期备份 SQLite 数据库。当前版本默认仅监听本机地址，不会上传云端。</div></div></div>`;
}

async function loadNetworkInfo(box) {
  try {
    const info = await api('/api/system/network');
    if (!document.body.contains(box)) return;
    if (!info.shareEnabled) {
      box.innerHTML = localizeHtml(`<div class="notice">当前仅本机可访问。若要让手机、平板等其他设备使用，请退出后以家庭共享模式启动（设置环境变量 <code>FAMILY_SCORE_LAN=1</code>，桌面安装版默认开启），首次启动请在系统防火墙提示中允许访问。</div>`);
      return;
    }
    const urls = (info.urls || []).map(u => `<li><code>${h(u)}</code></li>`).join('');
    const qr = info.qrPngBase64 ? `<div class="share-qr-row"><img class="share-qr" src="data:image/png;base64,${info.qrPngBase64}" alt="家庭访问二维码" /><p>手机扫码直接打开</p></div>` : '';
    box.innerHTML = localizeHtml(`<div class="notice">家庭共享已开启。同一 Wi-Fi 下的手机、平板或其他电脑，用浏览器打开下面任一地址即可使用，登录各自账号，数据实时同步。</div><ul class="share-urls">${urls}</ul>${qr}`);
  } catch (err) {
    if (document.body.contains(box)) box.innerHTML = `<div class="notice">${h(err.message)}</div>`;
  }
}
