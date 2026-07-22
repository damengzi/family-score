package service

import (
	"context"
	"encoding/base64"
	"errors"
	"net"
	"sort"

	"family-score/internal/protocol"

	qrcode "github.com/skip2/go-qrcode"
)

// NetworkInfo 返回本机局域网访问信息，仅管理员可见。
// 用于家庭共享场景：主机开启共享后，其他设备通过浏览器访问展示出来的地址。
func (s *Service) NetworkInfo(ctx context.Context, sess protocol.Session) (protocol.NetworkInfo, error) {
	if !IsAdmin(sess) {
		return protocol.NetworkInfo{}, errors.New("只有管理员可以查看网络信息")
	}
	addr := s.ListenAddr()
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return protocol.NetworkInfo{}, err
	}
	host, _, _ := net.SplitHostPort(addr)
	shareEnabled := host == "" || host == "0.0.0.0" || host == "::"
	info := protocol.NetworkInfo{
		ListenAddr:   addr,
		Port:         port,
		ShareEnabled: shareEnabled,
		URLs:         []string{},
	}
	if !shareEnabled {
		return info, nil
	}
	for _, ip := range lanIPv4s() {
		info.URLs = append(info.URLs, "http://"+ip+":"+port)
	}
	if len(info.URLs) > 0 {
		info.PrimaryURL = info.URLs[0]
		if png, qrErr := qrcode.Encode(info.PrimaryURL, qrcode.Medium, 256); qrErr == nil {
			info.QRPngBase64 = base64.StdEncoding.EncodeToString(png)
		}
	}
	return info, nil
}

// lanIPv4s 返回本机所有可用于局域网访问的 IPv4 地址。
func lanIPv4s() []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	ips := []string{}
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipNet.IP.To4()
		if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			continue
		}
		ips = append(ips, ip.String())
	}
	sort.Strings(ips)
	return ips
}
