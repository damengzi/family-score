package local.familyscore;

import android.app.Activity;
import android.os.Bundle;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import java.io.File;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.Socket;

/**
 * 安卓平板壳：启动内置的 Go 本机服务（家庭共享模式），
 * 再用 WebView 打开本机页面。平板既可自己使用，
 * 也可作为家庭主机，让同一 Wi-Fi 下的其他设备访问。
 */
public class MainActivity extends Activity {

    private static final String LOCAL_URL = "http://127.0.0.1:8080";

    private Process serverProcess;
    private WebView webView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        startServer();

        webView = new WebView(this);
        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        webView.setWebViewClient(new WebViewClient());
        webView.setKeepScreenOn(true);
        setContentView(webView);

        loadWhenReady(0);
    }

    private void startServer() {
        try {
            String serverBin = getApplicationInfo().nativeLibraryDir + "/libfamilyscore.so";
            File logDir = new File(getFilesDir(), "logs");
            //noinspection ResultOfMethodCallIgnored
            logDir.mkdirs();
            ProcessBuilder pb = new ProcessBuilder(serverBin);
            // 数据保存在应用私有目录，卸载即清除。
            pb.environment().put("FAMILY_SCORE_DATA_DIR", getFilesDir().getAbsolutePath());
            // 家庭共享模式：允许同一局域网内的手机、电脑访问本机服务。
            pb.environment().put("FAMILY_SCORE_LAN", "1");
            pb.redirectErrorStream(true);
            pb.redirectOutput(new File(logDir, "server.log"));
            serverProcess = pb.start();
        } catch (IOException e) {
            throw new RuntimeException("内置服务启动失败", e);
        }
    }

    private void loadWhenReady(final int attempt) {
        webView.postDelayed(new Runnable() {
            @Override
            public void run() {
                if (isServerReady()) {
                    webView.loadUrl(LOCAL_URL);
                } else if (attempt < 50) {
                    loadWhenReady(attempt + 1);
                } else {
                    webView.loadData("<h3>服务启动失败，请重启应用</h3>", "text/html; charset=utf-8", "utf-8");
                }
            }
        }, attempt == 0 ? 300 : 200);
    }

    private boolean isServerReady() {
        try (Socket socket = new Socket()) {
            socket.connect(new InetSocketAddress("127.0.0.1", 8080), 200);
            return true;
        } catch (IOException e) {
            return false;
        }
    }

    @Override
    public void onBackPressed() {
        if (webView != null && webView.canGoBack()) {
            webView.goBack();
        } else {
            super.onBackPressed();
        }
    }

    @Override
    protected void onDestroy() {
        if (serverProcess != null) {
            serverProcess.destroy();
            serverProcess = null;
        }
        super.onDestroy();
    }
}
