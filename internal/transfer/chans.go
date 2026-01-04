package transfer

type Channels struct {
	ComChannel     chan ImportResult
	ErrChannel     chan string
	BoundedChannel chan bool
}

type ImportResult struct {
	Message   string
	IsPrivate bool
}
