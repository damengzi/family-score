#!/bin/bash
# 三平台桌面安装包构建脚本。
# 产物输出到 dist/：
#   dist/family-score-mac-intel.zip        Mac Intel 芯片（.app）
#   dist/family-score-mac-apple-silicon.zip Mac M 系列芯片（.app）
#   dist/family-score-windows-amd64.zip    Windows 64 位（.exe + 启动脚本）
set -euo pipefail

APP_NAME="family-score"
APP_DISPLAY_NAME="FamilyScore"
BUNDLE_ID="local.family-score"
GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DIST_DIR="$ROOT_DIR/dist"
BUILD_DIR="$DIST_DIR/build"

rm -rf "$DIST_DIR"
mkdir -p "$BUILD_DIR"

# Go 编译临时目录放在项目内，避免受限环境下无法写入系统临时目录。
export GOTMPDIR="$ROOT_DIR/.gotmp"
mkdir -p "$GOTMPDIR"
trap 'rm -rf "$GOTMPDIR"' EXIT

build_bin() {
  local goos="$1" goarch="$2" out="$3"
  echo ">> 构建 $goos/$goarch -> $out"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" GOPROXY="$GOPROXY" \
    go build -trimpath -ldflags="-s -w" -o "$out" "$ROOT_DIR"
}

# 生成 macOS .app：双击启动本地服务（默认开启家庭共享），并自动打开浏览器。
make_mac_app() {
  local bin="$1" app_dir="$2"
  mkdir -p "$app_dir/Contents/MacOS"
  cp "$bin" "$app_dir/Contents/MacOS/$APP_NAME-bin"
  cat > "$app_dir/Contents/MacOS/$APP_DISPLAY_NAME" <<'EOF'
#!/bin/bash
DIR="$(cd "$(dirname "$0")" && pwd)"
export FAMILY_SCORE_LAN=1
( sleep 2; open "http://127.0.0.1:8080" ) &
exec "$DIR/family-score-bin"
EOF
  chmod +x "$app_dir/Contents/MacOS/$APP_DISPLAY_NAME" "$app_dir/Contents/MacOS/$APP_NAME-bin"
  cat > "$app_dir/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key><string>$APP_DISPLAY_NAME</string>
  <key>CFBundleDisplayName</key><string>家庭德育积分系统</string>
  <key>CFBundleIdentifier</key><string>$BUNDLE_ID</string>
  <key>CFBundleExecutable</key><string>$APP_DISPLAY_NAME</string>
  <key>CFBundleVersion</key><string>1.0.0</string>
  <key>CFBundleShortVersionString</key><string>1.0.0</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>LSMinimumSystemVersion</key><string>11.0</string>
  <key>NSHighResolutionCapable</key><true/>
</dict>
</plist>
EOF
}

# ---------- Mac Intel ----------
build_bin darwin amd64 "$BUILD_DIR/$APP_NAME-darwin-amd64"
make_mac_app "$BUILD_DIR/$APP_NAME-darwin-amd64" "$BUILD_DIR/mac-intel/$APP_DISPLAY_NAME.app"
(cd "$BUILD_DIR/mac-intel" && zip -qry "$DIST_DIR/$APP_NAME-mac-intel.zip" "$APP_DISPLAY_NAME.app")

# ---------- Mac Apple Silicon ----------
build_bin darwin arm64 "$BUILD_DIR/$APP_NAME-darwin-arm64"
make_mac_app "$BUILD_DIR/$APP_NAME-darwin-arm64" "$BUILD_DIR/mac-apple-silicon/$APP_DISPLAY_NAME.app"
(cd "$BUILD_DIR/mac-apple-silicon" && zip -qry "$DIST_DIR/$APP_NAME-mac-apple-silicon.zip" "$APP_DISPLAY_NAME.app")

# ---------- Windows 64 位 ----------
# 注意：zip 内文件名使用 ASCII，避免 Windows 资源管理器按本地代码页解压时中文文件名乱码。
build_bin windows amd64 "$BUILD_DIR/$APP_NAME-windows-amd64.exe"
WIN_PKG="$BUILD_DIR/windows-amd64/$APP_NAME"
mkdir -p "$WIN_PKG"
cp "$BUILD_DIR/$APP_NAME-windows-amd64.exe" "$WIN_PKG/$APP_NAME.exe"
cat > "$WIN_PKG/start-family-sharing.bat" <<'EOF'
@echo off
set FAMILY_SCORE_LAN=1
start "" "http://127.0.0.1:8080"
"%~dp0family-score.exe"
EOF
# README 使用 UTF-8 BOM，保证 Windows 记事本正确显示中文。
{ printf '\xef\xbb\xbf'; cat <<'EOF'
家庭德育积分系统（Windows 64 位）

1. 双击 start-family-sharing.bat 启动系统，浏览器会自动打开使用页面。
2. 首次启动如弹出 Windows 防火墙提示，请选择“允许访问”，
   这样家里同一 Wi-Fi 下的手机、平板也能访问（地址见系统内“本机备份-家庭共享访问”）。
3. 直接双击 family-score.exe 则仅本机可用。
4. 数据保存在 用户目录\.family-score\，卸载前可在系统页面做本机备份。
EOF
} > "$WIN_PKG/README.txt"
(cd "$BUILD_DIR/windows-amd64" && zip -qry "$DIST_DIR/$APP_NAME-windows-amd64.zip" "$APP_NAME")

# ---------- Android 平板（arm64） ----------
# Go 服务交叉编译为 android/arm64 可执行文件，以 libfamilyscore.so 名义放入 jniLibs，
# 安装后由安卓壳 MainActivity 直接执行。
build_bin android arm64 "$BUILD_DIR/libfamilyscore.so"
JNI_DIR="$ROOT_DIR/android/app/src/main/jniLibs/arm64-v8a"
mkdir -p "$JNI_DIR"
cp "$BUILD_DIR/libfamilyscore.so" "$JNI_DIR/libfamilyscore.so"
SDK_DIR="${ANDROID_HOME:-$HOME/Library/Android/sdk}"
if [ -d "$SDK_DIR" ] && command -v gradle >/dev/null 2>&1; then
  echo ">> 检测到 Android SDK 与 gradle，开始构建 APK"
  (cd "$ROOT_DIR/android" && gradle --no-daemon assembleDebug)
  cp "$ROOT_DIR/android/app/build/outputs/apk/debug/app-debug.apk" "$DIST_DIR/$APP_NAME-android-arm64.apk"
else
  echo ">> 未检测到 Android SDK 或 gradle，跳过 APK 构建"
  echo "   原生库已生成：android/app/src/main/jniLibs/arm64-v8a/libfamilyscore.so"
  echo "   用 Android Studio 打开 android/ 目录即可一键构建安装包"
fi

rm -rf "$BUILD_DIR"
echo ">> 完成，产物："
ls -lh "$DIST_DIR"
