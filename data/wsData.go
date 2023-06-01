package wsData

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`

	Board [][]int `json:"board,omitempty"`
	Pos   string  `json:"pos,omitempty"`
}
type BoardState struct {
	Board  [13][13]int `json:"board"`
	Turn   string      `json:"turn"`
	Winner string      `json:"winner"`
}

func HelloWord() string {
	return "Hello world"

}

type Response struct {
	Status  int         `json:"status,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Type    string      `json:"type,omitempty"`
}
