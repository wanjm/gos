package biz

import (
	"github.com/wan_jm/servlet_example/entity/mysql/note"
	"github.com/wanjm/common"
)

// noteFilters builds one query that MySQL and Mongo DAL methods both accept.
// It uses Eq, Ne, In, NotIn, Gt, Gte, Lt, Lte, and Exist on generated column constants.
func noteFilters(name string, ownerID int32) []common.Optioner {
	return []common.Optioner{
		common.Eq(note.C_Name, name),
		common.Ne(note.C_Status, NoteStatus_Published),
		common.In(note.C_OwnerID, []int32{ownerID}),
		common.NotIn(note.C_ID, []int32{0}),
		common.Gt(note.C_ID, int32(0)),
		common.Gte(note.C_OwnerID, ownerID),
		common.Lt(note.C_Status, int32(9)),
		common.Lte(note.C_Status, NoteStatus_Published),
		common.Exist(note.C_Name, true),
	}
}
