package tests_test

import (
	"testing"
	"time"

	"gorm.io/gorm"
	. "gorm.io/gorm/utils/tests"
)

func TestCustomizeColumn(t *testing.T) {
	type CustomizeColumn struct {
		ID   int64      `gorm:"column:mapped_id; primary_key:yes"`
		Name string     `gorm:"column:mapped_name"`
		Date *time.Time `gorm:"column:mapped_time"`
	}

	DB.Migrator().DropTable(&CustomizeColumn{})
	DB.AutoMigrate(&CustomizeColumn{})

	expected := "foo"
	now := time.Now()
	cc := CustomizeColumn{ID: 666, Name: expected, Date: &now}

	if count := DB.Create(&cc).RowsAffected; count != 1 {
		t.Error("There should be one record be affected when create record")
	}

	var cc1 CustomizeColumn
	DB.First(&cc1, "mapped_name = ?", "foo")

	if cc1.Name != expected {
		t.Errorf("Failed to query CustomizeColumn")
	}

	cc.Name = "bar"
	DB.Save(&cc)

	var cc2 CustomizeColumn
	DB.First(&cc2, "mapped_id = ?", 666)
	if cc2.Name != "bar" {
		t.Errorf("Failed to query CustomizeColumn")
	}
}

func TestCustomColumnAndIgnoredFieldClash(t *testing.T) {
	// Make sure an ignored field does not interfere with another field's custom
	// column name that matches the ignored field.
	type CustomColumnAndIgnoredFieldClash struct {
		Body    string `gorm:"-"`
		RawBody string `gorm:"column:body"`
	}

	DB.Migrator().DropTable(&CustomColumnAndIgnoredFieldClash{})

	if err := DB.AutoMigrate(&CustomColumnAndIgnoredFieldClash{}); err != nil {
		t.Errorf("Should not raise error: %v", err)
	}
}

func TestCustomizeField(t *testing.T) {
	type CustomizeFieldStruct struct {
		gorm.Model
		Name                   string
		FieldAllowCreate       string `gorm:"<-:create"`
		FieldAllowUpdate       string `gorm:"<-:update"`
		FieldAllowSave         string `gorm:"<-"`
		FieldAllowSave2        string `gorm:"<-:create,update"`
		FieldAllowSave3        string `gorm:"->:false;<-:create"`
		FieldReadonly          string `gorm:"->"`
		FieldIgnore            string `gorm:"-"`
		AutoUnixCreateTime     int64  `gorm:"autocreatetime"`
		AutoUnixNanoCreateTime int64  `gorm:"autocreatetime:nano"`
		AutoUnixUpdateTime     int64  `gorm:"autoupdatetime"`
		AutoUnixNanoUpdateTime int64  `gorm:"autoupdatetime:nano"`
	}

	DB.Migrator().DropTable(&CustomizeFieldStruct{})

	if err := DB.AutoMigrate(&CustomizeFieldStruct{}); err != nil {
		t.Errorf("Failed to migrate, got error: %v", err)
	}

	if DB.Migrator().HasColumn(&CustomizeFieldStruct{}, "FieldIgnore") {
		t.Errorf("FieldIgnore should not be created")
	}

	if DB.Migrator().HasColumn(&CustomizeFieldStruct{}, "field_ignore") {
		t.Errorf("FieldIgnore should not be created")
	}

	generateStruct := func(name string) *CustomizeFieldStruct {
		return &CustomizeFieldStruct{
			Name:             name,
			FieldAllowCreate: name + "_allow_create",
			FieldAllowUpdate: name + "_allow_update",
			FieldAllowSave:   name + "_allow_save",
			FieldAllowSave2:  name + "_allow_save2",
			FieldAllowSave3:  name + "_allow_save3",
			FieldReadonly:    name + "_allow_readonly",
			FieldIgnore:      name + "_allow_ignore",
		}
	}

	create := generateStruct("create")
	DB.Create(&create)

	var result CustomizeFieldStruct
	DB.Find(&result, "name = ?", "create")

	AssertObjEqual(t, result, create, "Name", "FieldAllowCreate", "FieldAllowSave", "FieldAllowSave2")

	if result.FieldAllowUpdate != "" || result.FieldReadonly != "" || result.FieldIgnore != "" || result.FieldAllowSave3 != "" {
		t.Fatalf("invalid result: %#v", result)
	}

	if result.AutoUnixCreateTime != result.AutoUnixUpdateTime || result.AutoUnixCreateTime == 0 {
		t.Fatalf("invalid create/update unix time: %#v", result)
	}

	if result.AutoUnixNanoCreateTime != result.AutoUnixNanoUpdateTime || result.AutoUnixNanoCreateTime == 0 || result.AutoUnixNanoCreateTime/result.AutoUnixCreateTime < 1e6 {
		t.Fatalf("invalid create/update unix nano time: %#v", result)
	}

	result.FieldAllowUpdate = "field_allow_update_updated"
	result.FieldReadonly = "field_readonly_updated"
	result.FieldIgnore = "field_ignore_updated"
	DB.Save(&result)

	var result2 CustomizeFieldStruct
	DB.Find(&result2, "name = ?", "create")

	if result2.FieldAllowUpdate != result.FieldAllowUpdate || result2.FieldReadonly != "" || result2.FieldIgnore != "" {
		t.Fatalf("invalid updated result: %#v", result2)
	}

	if err := DB.Table("customize_field_structs").Where("1 = 1").UpdateColumn("field_readonly", "readonly").Error; err != nil {
		t.Fatalf("failed to update field_readonly column")
	}

	var result3 CustomizeFieldStruct
	DB.Find(&result3, "name = ?", "create")

	if result3.FieldReadonly != "readonly" {
		t.Fatalf("invalid updated result: %#v", result3)
	}

	var result4 CustomizeFieldStruct
	if err := DB.First(&result4, "field_allow_save3 = ?", create.FieldAllowSave3).Error; err != nil {
		t.Fatalf("failed to query with inserted field, got error %v", err)
	}

	AssertEqual(t, result3, result4)

	createWithDefaultTime := generateStruct("create_with_default_time")
	createWithDefaultTime.AutoUnixCreateTime = 100
	createWithDefaultTime.AutoUnixUpdateTime = 100
	createWithDefaultTime.AutoUnixNanoCreateTime = 100
	createWithDefaultTime.AutoUnixNanoUpdateTime = 100
	DB.Create(&createWithDefaultTime)

	var createWithDefaultTimeResult CustomizeFieldStruct
	DB.Find(&createWithDefaultTimeResult, "name = ?", createWithDefaultTime.Name)

	if createWithDefaultTimeResult.AutoUnixCreateTime != createWithDefaultTimeResult.AutoUnixUpdateTime || createWithDefaultTimeResult.AutoUnixCreateTime != 100 {
		t.Fatalf("invalid create/update unix time: %#v", createWithDefaultTimeResult)
	}

	if createWithDefaultTimeResult.AutoUnixNanoCreateTime != createWithDefaultTimeResult.AutoUnixNanoUpdateTime || createWithDefaultTimeResult.AutoUnixNanoCreateTime != 100 {
		t.Fatalf("invalid create/update unix nano time: %#v", createWithDefaultTimeResult)
	}
}