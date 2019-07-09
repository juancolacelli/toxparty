package tox

type Message struct {
	SenderNumber uint32
	SenderName   string
	Text         string
	IsAction     bool
}
