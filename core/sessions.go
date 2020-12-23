package core

type session struct {
	id      int
	implant implant
	rhost   string
	checkin string
	status  string
}
