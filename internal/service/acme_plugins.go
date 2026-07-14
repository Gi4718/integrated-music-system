package service

// ACMEPlugin 定义一个 DNS 验证插件
type ACMEPlugin struct {
	ID          string            `json:"id"`
	Label       string            `json:"label"`
	Delay       int               `json:"delay"` // 验证延迟（秒）
	Fields      []PluginField     `json:"fields"`
	EnvMap      map[string]string `json:"-"` // 字段 key -> 环境变量名
}

// PluginField 定义插件的一个配置字段
type PluginField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"` // text, password
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder"`
	Hint        string `json:"hint"`
}

// GetPlugins 返回所有可用的 DNS 验证插件
func GetPlugins() []ACMEPlugin {
	return []ACMEPlugin{
		{
			ID:    "cloudflare",
			Label: "Cloudflare Managed DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "CF_Token", Label: "API Token", Type: "password", Required: false, Placeholder: "全局 API Token 或 DNS 编辑 Token", Hint: "需要 Zone:DNS:Edit 权限"},
				{Key: "CF_Email", Label: "Email", Type: "text", Required: false, Placeholder: "Cloudflare 账号邮箱", Hint: "使用 API Key 模式时必填"},
				{Key: "CF_Key", Label: "Global API Key", Type: "password", Required: false, Placeholder: "Global API Key", Hint: "与 API Token 二选一"},
				{Key: "CF_Account_ID", Label: "Account ID", Type: "text", Required: false, Placeholder: "可选"},
				{Key: "CF_Zone_ID", Label: "Zone ID", Type: "text", Required: false, Placeholder: "可选"},
			},
			EnvMap: map[string]string{
				"CF_Token":      "CLOUDFLARE_DNS_API_TOKEN",
				"CF_Email":      "CLOUDFLARE_EMAIL",
				"CF_Key":        "CLOUDFLARE_API_KEY",
				"CF_Account_ID": "CLOUDFLARE_ACCOUNT_ID",
				"CF_Zone_ID":    "CLOUDFLARE_ZONE_ID",
			},
		},
		{
			ID:    "alidns",
			Label: "阿里云 DNS (AliDNS)",
			Delay: 30,
			Fields: []PluginField{
				{Key: "ALICLOUD_ACCESS_KEY", Label: "AccessKey ID", Type: "text", Required: true, Placeholder: "LTAI...", Hint: "RAM 子账号需 AliyunDNSFullAccess 权限"},
				{Key: "ALICLOUD_SECRET_KEY", Label: "AccessKey Secret", Type: "password", Required: true, Placeholder: "密钥"},
				{Key: "ALICLOUD_REGION_ID", Label: "Region ID", Type: "text", Required: false, Placeholder: "cn-hangzhou", Hint: "默认 cn-hangzhou"},
			},
			EnvMap: map[string]string{
				"ALICLOUD_ACCESS_KEY": "ALICLOUD_ACCESS_KEY",
				"ALICLOUD_SECRET_KEY": "ALICLOUD_SECRET_KEY",
				"ALICLOUD_REGION_ID":  "ALICLOUD_REGION_ID",
			},
		},
		{
			ID:    "tencentcloud",
			Label: "腾讯云 DNS (DNSPod)",
			Delay: 30,
			Fields: []PluginField{
				{Key: "TENCENTCLOUD_SECRET_ID", Label: "Secret ID", Type: "text", Required: true, Placeholder: "AKID...", Hint: "子账号需 DNSPod 管理权限"},
				{Key: "TENCENTCLOUD_SECRET_KEY", Label: "Secret Key", Type: "password", Required: true, Placeholder: "密钥"},
			},
			EnvMap: map[string]string{
				"TENCENTCLOUD_SECRET_ID":  "TENCENTCLOUD_SECRET_ID",
				"TENCENTCLOUD_SECRET_KEY": "TENCENTCLOUD_SECRET_KEY",
			},
		},
		{
			ID:    "dnspod",
			Label: "DNSPod (独立)",
			Delay: 30,
			Fields: []PluginField{
				{Key: "DNSPOD_API_KEY", Label: "API Key", Type: "password", Required: true, Placeholder: "ID,Token", Hint: "格式：数字ID,Token（如 12345,abc...）"},
			},
			EnvMap: map[string]string{
				"DNSPOD_API_KEY": "DNSPOD_API_KEY",
			},
		},
		{
			ID:    "huaweicloud",
			Label: "华为云 DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "HUAWEICLOUD_DOMAIN_NAME", Label: "Domain Name", Type: "text", Required: false, Placeholder: "域名", Hint: "可选，自动从域名推断"},
				{Key: "HUAWEICLOUD_ACCESS_KEY_ID", Label: "Access Key ID", Type: "text", Required: true, Placeholder: "AK..."},
				{Key: "HUAWEICLOUD_SECRET_ACCESS_KEY", Label: "Secret Access Key", Type: "password", Required: true, Placeholder: "SK..."},
			},
			EnvMap: map[string]string{
				"HUAWEICLOUD_DOMAIN_NAME":       "HUAWEICLOUD_DOMAIN_NAME",
				"HUAWEICLOUD_ACCESS_KEY_ID":     "HUAWEICLOUD_ACCESS_KEY_ID",
				"HUAWEICLOUD_SECRET_ACCESS_KEY": "HUAWEICLOUD_SECRET_ACCESS_KEY",
			},
		},
		{
			ID:    "baiducloud",
			Label: "百度云 DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "BAIDUCLOUD_ACCESS_KEY_ID", Label: "Access Key ID", Type: "text", Required: true, Placeholder: "AK..."},
				{Key: "BAIDUCLOUD_SECRET_ACCESS_KEY", Label: "Secret Access Key", Type: "password", Required: true, Placeholder: "SK..."},
			},
			EnvMap: map[string]string{
				"BAIDUCLOUD_ACCESS_KEY_ID":     "BAIDUCLOUD_ACCESS_KEY_ID",
				"BAIDUCLOUD_SECRET_ACCESS_KEY": "BAIDUCLOUD_SECRET_ACCESS_KEY",
			},
		},
		{
			ID:    "route53",
			Label: "Amazon Route53 (AWS)",
			Delay: 30,
			Fields: []PluginField{
				{Key: "AWS_ACCESS_KEY_ID", Label: "Access Key ID", Type: "text", Required: true, Placeholder: "AKIA..."},
				{Key: "AWS_SECRET_ACCESS_KEY", Label: "Secret Access Key", Type: "password", Required: true, Placeholder: "密钥"},
				{Key: "AWS_HOSTED_ZONE_ID", Label: "Hosted Zone ID", Type: "text", Required: false, Placeholder: "Z0123456...", Hint: "可选，自动查找"},
				{Key: "AWS_REGION", Label: "Region", Type: "text", Required: false, Placeholder: "us-east-1", Hint: "默认 us-east-1"},
			},
			EnvMap: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AWS_ACCESS_KEY_ID",
				"AWS_SECRET_ACCESS_KEY": "AWS_SECRET_ACCESS_KEY",
				"AWS_HOSTED_ZONE_ID":    "AWS_HOSTED_ZONE_ID",
				"AWS_REGION":            "AWS_REGION",
			},
		},
		{
			ID:    "azure",
			Label: "Azure DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "AZURE_CLIENT_ID", Label: "Client ID", Type: "text", Required: true, Placeholder: "应用 ID"},
				{Key: "AZURE_CLIENT_SECRET", Label: "Client Secret", Type: "password", Required: true, Placeholder: "密钥"},
				{Key: "AZURE_TENANT_ID", Label: "Tenant ID", Type: "text", Required: true, Placeholder: "目录 ID"},
				{Key: "AZURE_SUBSCRIPTION_ID", Label: "Subscription ID", Type: "text", Required: true, Placeholder: "订阅 ID"},
				{Key: "AZURE_RESOURCE_GROUP", Label: "Resource Group", Type: "text", Required: true, Placeholder: "资源组名"},
			},
			EnvMap: map[string]string{
				"AZURE_CLIENT_ID":       "AZURE_CLIENT_ID",
				"AZURE_CLIENT_SECRET":   "AZURE_CLIENT_SECRET",
				"AZURE_TENANT_ID":       "AZURE_TENANT_ID",
				"AZURE_SUBSCRIPTION_ID": "AZURE_SUBSCRIPTION_ID",
				"AZURE_RESOURCE_GROUP":  "AZURE_RESOURCE_GROUP",
			},
		},
		{
			ID:    "gcloud",
			Label: "Google Cloud DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "GCE_PROJECT", Label: "Project ID", Type: "text", Required: true, Placeholder: "项目 ID"},
				{Key: "GCE_SERVICE_ACCOUNT_FILE", Label: "Service Account JSON", Type: "text", Required: true, Placeholder: "/path/to/key.json", Hint: "容器内绝对路径"},
			},
			EnvMap: map[string]string{
				"GCE_PROJECT":                "GCE_PROJECT",
				"GCE_SERVICE_ACCOUNT_FILE":   "GCE_SERVICE_ACCOUNT_FILE",
			},
		},
		{
			ID:    "ovh",
			Label: "OVH DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "OVH_ENDPOINT", Label: "Endpoint", Type: "text", Required: true, Placeholder: "ovh-eu / ovh-ca / soyoustart-eu", Hint: "API 端点"},
				{Key: "OVH_APPLICATION_KEY", Label: "Application Key", Type: "text", Required: true, Placeholder: "应用密钥"},
				{Key: "OVH_APPLICATION_SECRET", Label: "Application Secret", Type: "password", Required: true, Placeholder: "应用密文"},
				{Key: "OVH_CONSUMER_KEY", Label: "Consumer Key", Type: "password", Required: true, Placeholder: "消费者密钥"},
			},
			EnvMap: map[string]string{
				"OVH_ENDPOINT":          "OVH_ENDPOINT",
				"OVH_APPLICATION_KEY":   "OVH_APPLICATION_KEY",
				"OVH_APPLICATION_SECRET": "OVH_APPLICATION_SECRET",
				"OVH_CONSUMER_KEY":      "OVH_CONSUMER_KEY",
			},
		},
		{
			ID:    "digitalocean",
			Label: "DigitalOcean DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "DO_AUTH_TOKEN", Label: "API Token", Type: "password", Required: true, Placeholder: "Personal Access Token", Hint: "需要 write 权限"},
			},
			EnvMap: map[string]string{
				"DO_AUTH_TOKEN": "DO_AUTH_TOKEN",
			},
		},
		{
			ID:    "godaddy",
			Label: "GoDaddy DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "GODADDY_API_KEY", Label: "API Key", Type: "text", Required: true, Placeholder: "密钥"},
				{Key: "GODADDY_API_SECRET", Label: "API Secret", Type: "password", Required: true, Placeholder: "密文"},
			},
			EnvMap: map[string]string{
				"GODADDY_API_KEY":    "GODADDY_API_KEY",
				"GODADDY_API_SECRET": "GODADDY_API_SECRET",
			},
		},
		{
			ID:    "namecheap",
			Label: "Namecheap DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "NAMECHEAP_API_KEY", Label: "API Key", Type: "password", Required: true, Placeholder: "Namecheap API Key"},
				{Key: "NAMECHEAP_API_USER", Label: "API User", Type: "text", Required: true, Placeholder: "用户名"},
			},
			EnvMap: map[string]string{
				"NAMECHEAP_API_KEY":  "NAMECHEAP_API_KEY",
				"NAMECHEAP_API_USER": "NAMECHEAP_API_USER",
			},
		},
		{
			ID:    "linode",
			Label: "Linode DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "LINODE_TOKEN", Label: "API Token", Type: "password", Required: true, Placeholder: "Personal Access Token", Hint: "需要 Domains:Read/Write 权限"},
			},
			EnvMap: map[string]string{
				"LINODE_TOKEN": "LINODE_TOKEN",
			},
		},
		{
			ID:    "vultr",
			Label: "Vultr DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "VULTR_API_KEY", Label: "API Key", Type: "password", Required: true, Placeholder: "Vultr API Key"},
			},
			EnvMap: map[string]string{
				"VULTR_API_KEY": "VULTR_API_KEY",
			},
		},
		{
			ID:    "hetzner",
			Label: "Hetzner DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "HETZNER_API_KEY", Label: "API Token", Type: "password", Required: true, Placeholder: "Hetzner DNS API Token"},
			},
			EnvMap: map[string]string{
				"HETZNER_API_KEY": "HETZNER_API_KEY",
			},
		},
		{
			ID:    "porkbun",
			Label: "Porkbun DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "PORKBUN_API_KEY", Label: "API Key", Type: "text", Required: true, Placeholder: "API Key"},
				{Key: "PORKBUN_SECRET_KEY", Label: "Secret Key", Type: "password", Required: true, Placeholder: "Secret Key"},
			},
			EnvMap: map[string]string{
				"PORKBUN_API_KEY":    "PORKBUN_API_KEY",
				"PORKBUN_SECRET_KEY": "PORKBUN_SECRET_KEY",
			},
		},
		{
			ID:    "cloudxns",
			Label: "CloudXNS DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "CLOUDXNS_API_KEY", Label: "API Key", Type: "text", Required: true, Placeholder: "API Key"},
				{Key: "CLOUDXNS_SECRET_KEY", Label: "Secret Key", Type: "password", Required: true, Placeholder: "Secret Key"},
			},
			EnvMap: map[string]string{
				"CLOUDXNS_API_KEY":    "CLOUDXNS_API_KEY",
				"CLOUDXNS_SECRET_KEY": "CLOUDXNS_SECRET_KEY",
			},
		},
		{
			ID:    "west",
			Label: "西部数码 DNS",
			Delay: 30,
			Fields: []PluginField{
				{Key: "WEST_API_KEY", Label: "API Key", Type: "text", Required: true, Placeholder: "API Key"},
				{Key: "WEST_API_USER", Label: "API User", Type: "text", Required: true, Placeholder: "用户名"},
			},
			EnvMap: map[string]string{
				"WEST_API_KEY":  "WEST_API_KEY",
				"WEST_API_USER": "WEST_API_USER",
			},
		},
	}
}

// GetPluginByID 根据 ID 查找插件
func GetPluginByID(id string) *ACMEPlugin {
	plugins := GetPlugins()
	for i := range plugins {
		if plugins[i].ID == id {
			return &plugins[i]
		}
	}
	return nil
}
