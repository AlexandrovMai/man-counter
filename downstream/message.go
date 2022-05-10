package downstream

const (
	VisitorsMesName    = "visitors"
	ActiveUsersMesName = "active"
)

type Message struct {
	Name  string `json:"n"`
	Value int    `json:"v"`
}
