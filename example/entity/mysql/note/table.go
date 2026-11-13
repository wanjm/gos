package note

// NoteBase is embedded so a schema field such as Title can match a nested entity field.
type NoteBase struct {
	Title string `json:"title" gorm:"column:title"` // 标题
}

// Note is the MySQL note table.
// tblName and dbVariable select the table and the DAL database field; arrays and maps add list helpers.
// @gos tblName=note; dbVariable=db; arrays=id,owner_id; maps=id
type Note struct {
	ID      int32  `json:"id" gorm:"column:id"`            // 主键
	OwnerID int32  `json:"ownerId" gorm:"column:owner_id"` // 所属用户
	Name    string `json:"name" gorm:"column:name"`        // 名称
	Status  int32  `json:"status" gorm:"column:status"`    // 状态
	Kind    string `json:"kind" gorm:"column:kind"`        // 种类
	NoteBase
}

// TableName returns the physical MySQL table name.
// GORM uses it when the generated DAL builds a query.
func (Note) TableName() string {
	return "note"
}
