package talkog

import "time"

type Talkog struct {
	UserId    string // UUID
	Timestamp int64  // Unix Time
	Talk      string
}

func New(userId string, talk string) Talkog {
	return Talkog{
		UserId:    userId,
		Timestamp: time.Now().Unix(),
		Talk:      talk,
	}
}
