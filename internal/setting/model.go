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

	Model struct {
		db *gorm.DB
	}
)

func NewModel(db *gorm.DB) *Model {
	return &Model{
		db: db,
	}
}

func (s *Model) Get() (AppSettingSlice, error) {
	settings := []AppSetting{}
	err := s.db.Find(&settings).Error
	return settings, err
}

func (s *Model) GetByKey(key string) (AppSetting, error) {
	settings := AppSetting{}
	err := s.db.Where("key = ?", key).First(&settings).Error
	return settings, err
}

func (s *Model) Update(settings AppSetting) error {
	res := s.db.Model(&AppSetting{}).Where("key = ?", settings.Key).Updates(&settings)
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
