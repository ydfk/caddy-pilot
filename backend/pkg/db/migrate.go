package db

import (
	"go-fiber-starter/internal/model/basicauth"
	"go-fiber-starter/internal/model/certificate"
	"go-fiber-starter/internal/model/configversion"
	"go-fiber-starter/internal/model/dnsprovider"
	"go-fiber-starter/internal/model/passkey"
	"go-fiber-starter/internal/model/proxysite"
	"go-fiber-starter/internal/model/systemsetting"
	"go-fiber-starter/internal/model/user"
)

// autoMigrate 自动迁移数据库表
func autoMigrate() error {
	if err := DB.AutoMigrate(
		&user.User{},
		&basicauth.Credential{},
		&proxysite.ProxySite{},
		&configversion.ConfigVersion{},
		&certificate.CertificateProfile{},
		&dnsprovider.DNSProvider{},
		&passkey.Credential{},
		&systemsetting.Setting{},
	); err != nil {
		return err
	}

	return nil
}
