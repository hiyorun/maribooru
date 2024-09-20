package structs

import "gorm.io/gorm"

type (
	AppSettings struct {
		gorm.Model
		Key          string
		ValueBool    bool    `gorm:"default:false"`
		ValueInteger int     `gorm:"default:0"`
		ValueFloat   float64 `gorm:"default:0.0"`
		ValueString  string  `gorm:"default:''"`
	}

	AppSettingsSlice []AppSettings

	ResponseAppSettings struct {
		AdminCreated bool `json:"admin_created"`
	}
)

func (a AppSettingsSlice) ToResponse() ResponseAppSettings {
	response := ResponseAppSettings{}
	for _, setting := range a {
		if setting.Key == "ADMIN_CREATED" {
			response.AdminCreated = setting.ValueBool
		}
	}
	return response
}
