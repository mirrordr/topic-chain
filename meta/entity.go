package meta

type Session struct {
	Total            int
	Topics           []string
	Chains           []*TalkChain
	OrderedIndexList *SingLinkList
}

type TalkChain struct {
	Id       int
	Total    int
	Topic    string
	Question []string
	Answer   []string
}

type SingLinkList struct {
	Val  any
	Next *SingLinkList
}

func (l *SingLinkList) HeadInsert(node *SingLinkList) {
	if l.Next == nil {
		l.Next = node
	} else {
		nextNode := l.Next
		l.Next = node
		node.Next = nextNode
	}
}

func (l *SingLinkList) TailInsert(node *SingLinkList) {
	prevNode := l
	curNode := l.Next
	for curNode != nil {
		prevNode = curNode
		curNode = curNode.Next
	}
	prevNode.Next = node
}

func (l *SingLinkList) SearchNodeByVal(target any) (prevNode, targetNode *SingLinkList) {
	prevNode = l
	curNode := l.Next
	for curNode != nil {
		if curNode.Val == target {
			return prevNode, curNode
		}
		prevNode = curNode
		curNode = curNode.Next
	}
	return nil, nil
}

func (l *SingLinkList) ChangeNodeToHead(prevNode, targetNode *SingLinkList) {
	if targetNode != nil && prevNode != nil && prevNode.Next == targetNode {
		prevNode.Next = targetNode.Next
		l.HeadInsert(targetNode)
	}
}

func NewSession() *Session {
	return &Session{
		Total:  0,
		Topics: make([]string, 0, 10),
		Chains: make([]*TalkChain, 0, 10),
		OrderedIndexList: &SingLinkList{
			Val:  -1,
			Next: nil,
		},
	}
}

func NewTalkChain() *TalkChain {
	return &TalkChain{
		Total:    0,
		Topic:    "",
		Question: make([]string, 0, 10),
		Answer:   make([]string, 0, 10),
	}
}
