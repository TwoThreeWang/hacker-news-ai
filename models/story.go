package models

import "time"

// TbUser 用户表
type TbUser struct {
	ID        int `json:"id" gorm:"column:id"`
	PostCount int `gorm:"column:postCount;type:int"`
}

func (*TbUser) TableName() string {
	return "tb_user"
}

// TbPost 文章表
type TbPost struct {
	ID           int       `gorm:"column:id;primaryKey;autoIncrement"`
	Title        string    `gorm:"column:title;type:varchar(100);"`
	Link         string    `gorm:"column:link;type:varchar(1024)"`
	Status       string    `gorm:"column:status;type:varchar(20)"`
	Content      string    `gorm:"column:content;type:text"`
	UpVote       int       `gorm:"column:upVote;type:int"`
	CollectVote  int       `gorm:"column:collectVote;type:int"`
	Type         string    `gorm:"column:type;type:varchar(20)"`
	UserID       uint      `gorm:"column:user_id;type:int"`
	Pid          string    `gorm:"column:pid;type:varchar(20);unique"`
	CommentCount int       `gorm:"column:commentCount;type:int"`
	Point        float64   `gorm:"column:point;type:decimal(20,10)"`
	UpVoted      int       `gorm:"<-"`
	CollectVoted int       `gorm:"<-"`
	Top          int       `gorm:"column:top;type:int;default:0"`
	ClickVote    int       `gorm:"column:clickVote;type:int;default:0"`
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime"`
}

func (*TbPost) TableName() string {
	return "tb_post"
}

// TbPostTag 文章标签关联表
type TbPostTag struct {
	TbPostID int `gorm:"column:tb_post_id"`
	TbTagID  int `gorm:"column:tb_tag_id"`
}

func (*TbPostTag) TableName() string {
	return "tb_post_tag"
}

type Story struct {
	ID          int       `json:"id" gorm:"column:id;primaryKey"`
	Title       string    `json:"title" gorm:"column:title;type:varchar(100)"`
	URL         string    `json:"url" gorm:"column:url;type:varchar(1024)"`
	Score       int       `json:"score" gorm:"column:score"`
	Time        time.Time `json:"time" gorm:"column:time"`
	By          string    `json:"by" gorm:"column:by;type:varchar(50)"`
	Descendants int       `json:"descendants" gorm:"column:descendants"`
	Content     string    `json:"content" gorm:"column:content;type:text"`
	Summary     string    `json:"summary" gorm:"column:summary;type:text"`
}
