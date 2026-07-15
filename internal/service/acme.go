package service

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"endfield-music/internal/db"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns"
	"github.com/go-acme/lego/v4/registration"
)

type ACMEConfig struct {
	Provider string            `json:"provider"`
	Email    string            `json:"email"`
	Domain   string            `json:"domain"`
	Fields   map[string]string `json:"fields"`
}

type ACMEUser struct {
	Email        string
	Registration *registration.Resource
	Key          crypto.PrivateKey
}

func (u *ACMEUser) GetEmail() string              { return u.Email }
func (u *ACMEUser) GetRegistration() *registration.Resource { return u.Registration }
func (u *ACMEUser) GetPrivateKey() crypto.PrivateKey { return u.Key }

const certDir = "/data/ssl/acme"

func init() {
	os.MkdirAll(certDir, 0755)
}

func RunACME(config ACMEConfig) (certPath, keyPath string, err error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("生成私钥失败: %w", err)
	}

	user := &ACMEUser{
		Email: config.Email,
		Key:   privateKey,
	}

	c := lego.NewConfig(user)
	c.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(c)
	if err != nil {
		return "", "", fmt.Errorf("创建 ACME 客户端失败: %w", err)
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return "", "", fmt.Errorf("注册 ACME 账户失败: %w", err)
	}
	user.Registration = reg

	provider, err := createDNSProvider(config)
	if err != nil {
		return "", "", fmt.Errorf("创建 DNS 提供商失败: %w", err)
	}

	delay := 30
	plugin := GetPluginByID(config.Provider)
	if plugin != nil {
		delay = plugin.Delay
	}

	err = client.Challenge.SetDNS01Provider(provider,
		dns01.AddDNSTimeout(time.Duration(delay)*time.Second),
	)
	if err != nil {
		return "", "", fmt.Errorf("设置 DNS 挑战失败: %w", err)
	}

	request := certificate.ObtainRequest{
		Domains: []string{config.Domain},
		Bundle:  true,
	}

	certResource, err := client.Certificate.Obtain(request)
	if err != nil {
		return "", "", fmt.Errorf("获取证书失败: %w", err)
	}

	domainDir := filepath.Join(certDir, sanitizeDomain(config.Domain))
	os.MkdirAll(domainDir, 0755)

	certPath = filepath.Join(domainDir, "cert.pem")
	keyPath = filepath.Join(domainDir, "privkey.pem")

	if err := os.WriteFile(certPath, certResource.Certificate, 0644); err != nil {
		return "", "", fmt.Errorf("保存证书失败: %w", err)
	}
	if err := os.WriteFile(keyPath, certResource.PrivateKey, 0600); err != nil {
		return "", "", fmt.Errorf("保存私钥失败: %w", err)
	}

	return certPath, keyPath, nil
}

func createDNSProvider(config ACMEConfig) (challenge.Provider, error) {
	plugin := GetPluginByID(config.Provider)
	if plugin == nil {
		return nil, fmt.Errorf("不支持的 DNS 提供商: %s", config.Provider)
	}

	// 先清除该插件所有相关环境变量，避免残留值干扰
	for _, envName := range plugin.EnvMap {
		os.Unsetenv(envName)
	}

	// Cloudflare 特殊处理：API Token 和 Global API Key 互斥
	// 如果设置了有效的 API Token，就不设置 Email 和 API Key
	if config.Provider == "cloudflare" {
		token := config.Fields["CF_Token"]
		// 检测是否为 placeholder 文本（用户误填提示文字）
		isPlaceholder := token == "" || strings.Contains(token, "Token") || strings.Contains(token, "token") || strings.Contains(token, "全局") || strings.Contains(token, "编辑")
		if !isPlaceholder {
			os.Setenv("CLOUDFLARE_DNS_API_TOKEN", token)
			// 不设置 CLOUDFLARE_EMAIL 和 CLOUDFLARE_API_KEY，避免冲突
		} else {
			// 使用 Global API Key 模式
			if email, ok := config.Fields["CF_Email"]; ok && email != "" {
				os.Setenv("CLOUDFLARE_EMAIL", email)
			}
			if key, ok := config.Fields["CF_Key"]; ok && key != "" {
				os.Setenv("CLOUDFLARE_API_KEY", key)
			}
		}
		// Account ID 和 Zone ID 可选
		if accountID, ok := config.Fields["CF_Account_ID"]; ok && accountID != "" {
			os.Setenv("CLOUDFLARE_ACCOUNT_ID", accountID)
		}
		if zoneID, ok := config.Fields["CF_Zone_ID"]; ok && zoneID != "" {
			os.Setenv("CLOUDFLARE_ZONE_ID", zoneID)
		}
	} else {
		// 其他插件正常设置
		for key, envName := range plugin.EnvMap {
			if val, ok := config.Fields[key]; ok && val != "" {
				os.Setenv(envName, val)
			}
		}
	}

	return dns.NewDNSChallengeProviderByName(config.Provider)
}

func sanitizeDomain(domain string) string {
	return strings.ReplaceAll(domain, "*", "_wildcard")
}

func ParseCertificate(certPath string) (*x509.Certificate, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("无法解析证书")
	}
	return x509.ParseCertificate(block.Bytes)
}

func GetCertExpiry(certPath string) (time.Time, error) {
	cert, err := ParseCertificate(certPath)
	if err != nil {
		return time.Time{}, err
	}
	return cert.NotAfter, nil
}

func LoadACMEConfig(jsonStr string) ACMEConfig {
	var cfg ACMEConfig
	json.Unmarshal([]byte(jsonStr), &cfg)
	return cfg
}

// CheckAndRenewCert 检查证书是否即将过期，如果是则自动续期
// 返回是否需要续期、续期是否成功、错误信息
func CheckAndRenewCert() (bool, bool, error) {
	sslMode, _ := db.GetSetting("ssl_mode")
	if sslMode != "acme" {
		return false, false, nil
	}

	acmeConfigJSON, _ := db.GetSetting("acme_config")
	if acmeConfigJSON == "" {
		return false, false, nil
	}

	config := LoadACMEConfig(acmeConfigJSON)

	certPath, _ := db.GetSetting("ssl_cert_path")
	if certPath == "" {
		return false, false, nil
	}

	expiry, err := GetCertExpiry(certPath)
	if err != nil {
		return false, false, fmt.Errorf("读取证书过期时间失败: %w", err)
	}

	// 距离过期不到 30 天时自动续期
	threshold := time.Now().Add(30 * 24 * time.Hour)
	if expiry.After(threshold) {
		return false, false, nil
	}

	log.Printf("证书将于 %s 过期，开始自动续期", expiry.Format("2006-01-02"))

	newCertPath, newKeyPath, err := RunACME(config)
	if err != nil {
		return true, false, fmt.Errorf("自动续期失败: %w", err)
	}

	db.SetSetting("ssl_cert_path", newCertPath)
	db.SetSetting("ssl_key_path", newKeyPath)

	log.Printf("ACME 证书自动续期成功，新证书路径: %s", newCertPath)
	return true, true, nil
}
