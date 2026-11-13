package biz

// NoteInfo is the API shape copied from entity Note.
// FromEntity assigns same-named fields, including nested NoteBase.Title, then calls FormatEntity.
// @gos entity=Note
type NoteInfo struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title"`
	// @gos enum=NoteStatus
	Status int32 `json:"status"` // 笔记状态
	// @gos enum=NoteKind
	Kind  string `json:"kind"` // 笔记种类
	Extra string `json:"extra"`
}

// FormatEntity fills schema fields that have no same-named entity field.
// It copies Name into Extra after FromEntity has assigned Name.
func (s *NoteInfo) FormatEntity() {
	s.Extra = s.Name
}

// NoteInfoList is the list form of NoteInfo.
// gos generates FromEntitys for this defined slice type.
type NoteInfoList []*NoteInfo
