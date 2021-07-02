package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"cloudiac/portal/libs/db"
	"cloudiac/utils"
	"github.com/jinzhu/gorm"
)

type Attrs map[string]interface{}

type Modeler interface {
	TableName() string
	Validate() error
	ValidateAttrs(attrs Attrs) error
	Migrate(*db.Session) error
	//AddUniqueIndex(*db.Session) error
}

type ModelIdGenerator interface {
	NewId() string
}

type Id string

func (i Id) Value() (driver.Value, error) {
	return string(i), nil
}

func (i *Id) Scan(value interface{}) error {
	*i = ""
	switch v := value.(type) {
	case []byte:
		*i = Id(v)
	case string:
		*i = Id(v)
	default:
		return fmt.Errorf("invalid type %T, value: %T", value, value)
	}
	return nil
}

func NewId(prefix string) Id {
	return Id(utils.GenGuid(prefix))
}

type BaseModel struct {
	Id Id `gorm:"size:32;primary_key" json:"id"`
}

func (base *BaseModel) BeforeCreate(scope *gorm.Scope) error {
	// 为设置 Id 值的情况下默认生成一个无前缀的 id，如果对前缀有要求主动设置一个 Id 值,
	// 或者在 Model 层定义自己的 BeforeCreate() 方法
	if base.Id == "" {
		base.Id = NewId("")
	}
	return nil
}

func (BaseModel) Migrate(*db.Session) error {
	return nil
}

func (BaseModel) Validate() error {
	return nil
}

func (BaseModel) ValidateAttrs(attrs Attrs) error {
	return nil
}

func (BaseModel) AddUniqueIndex(sess *db.Session, index string, cols ...string) error {
	return sess.AddUniqueIndex(index, cols...)
}

type TimedModel struct {
	BaseModel

	CreatedAt time.Time `json:"createdAt" csv:"-" tsdb:"-"`
	UpdatedAt time.Time `json:"updatedAt" csv:"-" tsdb:"-"`
}

type SoftDeleteModel struct {
	TimedModel
	DeletedAt *time.Time `json:"-" csv:"-" sql:"index"`
	// 因为 deleted_at 字段的默认值为 NULL(gorm 也会依赖这个值做软删除)，会导致唯一约束与软删除冲突,
	// 所以我们增加 deleted_at_t 字段来避免这个情况。
	// 如果 model 需要同时支持软删除和唯一约束就需要在唯一约束索引中增加该字段
	// (使用 SoftDeleteModel.AddUniqueIndex() 方法添加索引时会自动加上该字段)。
	DeletedAtT int64 `json:"-" csv:"-" gorm:"default:0"`
}

func (SoftDeleteModel) AfterDelete(db *gorm.DB) error {
	return db.Unscoped().UpdateColumn("deleted_at_t", time.Now().Unix()).Error
}

func (m SoftDeleteModel) AddUniqueIndex(sess *db.Session, index string, cols ...string) error {
	cols = append(cols, "deleted_at_t")
	return m.TimedModel.AddUniqueIndex(sess, index, cols...)
}
