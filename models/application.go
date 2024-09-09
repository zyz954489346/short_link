package models

import (
	"gorm.io/gorm"
	"time"
)

// Application 授权应用信息
type Application struct {
	gorm.Model
	Name      string `gorm:"size:100;not null;default:''"`
	Key       string `gorm:"size:100;index:unique;not null;default:''"`
	Secret    string `gorm:"size:100;not null;default:''"`
	DeletedAt *time.Time
}

// TableName 指定表名
func (*Application) TableName() string {
	return "applications"
}

// GetSecretByKey 根据 key 查密钥
func (app *Application) GetSecretByKey(key string) (string, uint) {
	where := map[string]interface{}{
		"key": key,
	}

	Database().
		Select("id", "secret").
		Where(where).
		Find(app)

	return app.Secret, app.ID
}
