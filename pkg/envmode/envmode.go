package envmode

// UseEtcd 是否使用 etcd。
// sandbox=true：开发/沙箱，不连接 etcd。
// sandbox=false 或配置中未出现该字段（解析为 false）：生产环境，连接 etcd。
func UseEtcd(sandbox bool) bool {
	return !sandbox
}

// SandboxLabel 日志用简短说明
func SandboxLabel(sandbox bool) string {
	if sandbox {
		return "sandbox=true（开发环境，不使用 etcd）"
	}
	return "sandbox=false 或未配置（生产环境，使用 etcd）"
}
