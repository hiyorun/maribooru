package setting

import "gorm.io/gorm"

type (
	AppSetting struct {
		gorm.Model
		Key          string
		ValueBool    bool    `gorm:"default:false"`
		ValueInteger int     `gorm:"default:0"`
		ValueFloat   float64 `gorm:"default:0.0"`
		ValueString  string  `gorm:"default:''"`
	}

	AppSettingSlice []AppSetting

	SettingModel struct {
		db *gorm.DB
	}
)

func NewSettingModel(db *gorm.DB) *SettingModel {
	return &SettingModel{
		db: db,
	}
}

func (s *SettingModel) Get() (AppSettingSlice, error) {
	settings := []AppSetting{}
	err := s.db.Find(&settings).Error
	return settings, err
}

func (s *SettingModel) GetByKey(key string) (AppSetting, error) {
	settings := AppSetting{}
	err := s.db.Where("key = ?", key).First(&settings).Error
	return settings, err
}

func (s *SettingModel) Update(settings AppSetting) error {
	res := s.db.Model(&AppSetting{}).Where("key = ?", settings.Key).Updates(&settings)
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
