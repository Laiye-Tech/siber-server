package core

type FlowAction interface {
	Create()
	LinkedByFlow(flow Flow, option *FlowActionOption)
	Execute() error
	GetOptions() []*FlowActionOption
}

type FlowActionOption struct {
	Id    int
	Name  string
	Extra string
}
