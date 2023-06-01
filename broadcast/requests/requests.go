package requests

type Broadcast struct {
	Type    string `json:"type"`
	Message int    `json:"message"`
}

type Topology struct {
	Type     string              `json:"type"`
	Topology map[string][]string `json:"topology"`
}

type Gossip struct {
	Type    string `json:"type"`
	Sender  string `json:"sender"`
	Counter int    `json:"counter"`
	Message int    `json:"message"`
}
