package note_log

import "go.mongodb.org/mongo-driver/bson/primitive"

// NoteLog is the Mongo note_log collection.
// arrays and maps name the bson column note_id, the same way a MySQL entity names a gorm column.
// @gos tblName=note_log; dbVariable=mongoDB; arrays=note_id; maps=note_id
type NoteLog struct {
	ID     primitive.ObjectID `bson:"_id" json:"id"`
	NoteID int32              `bson:"note_id" json:"noteId"` // 笔记
	Action string             `bson:"action" json:"action"`  // 动作
}
