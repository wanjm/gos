package biz

// NoteStatus is the note lifecycle exported to Dart and TypeScript.
// displayWord overrides the line comment; a member without displayWord uses the line comment.
// @gos enum=NoteStatus
const (
	// @gos displayWord="草稿"
	NoteStatus_Draft     = 1 // 草稿
	NoteStatus_Published = 2 // 已发布
)

// NoteKind is a string enum. The line comment is the frontend display text.
// @gos enum=NoteKind
const (
	NoteKind_Text  = "text"  // 文本
	NoteKind_Image = "image" // 图片
)

// The enum name is inferred from the NoteFlag_ prefix when enum has no value.
// @gos enum
const (
	NoteFlag_On  = true  // 开
	NoteFlag_Off = false // 关
)
