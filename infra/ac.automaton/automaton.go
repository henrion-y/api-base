package ac_automaton

type acNode struct {
	fail      *acNode
	isPattern bool
	next      map[rune]*acNode
}

func newAcNode() *acNode {
	return &acNode{
		fail:      nil,
		isPattern: false,
		next:      map[rune]*acNode{},
	}
}

type acAutoMachine struct {
	root *acNode
}

func NewAcAutoMachine() *acAutoMachine {
	return &acAutoMachine{
		root: newAcNode(),
	}
}

func (ac *acAutoMachine) AddPattern(pattern string) {
	chars := []rune(pattern)
	iter := ac.root
	for _, c := range chars {
		if _, ok := iter.next[c]; !ok {
			iter.next[c] = newAcNode()
		}
		iter = iter.next[c]
	}
	iter.isPattern = true
}

func (ac *acAutoMachine) Build() {
	var queue []*acNode
	queue = append(queue, ac.root)
	for len(queue) != 0 {
		parent := queue[0]
		queue = queue[1:]

		for char, child := range parent.next {
			if parent == ac.root {
				child.fail = ac.root
			} else {
				failacNode := parent.fail
				for failacNode != nil {
					if _, ok := failacNode.next[char]; ok {
						child.fail = failacNode.next[char]
						break
					}
					failacNode = failacNode.fail
				}
				if failacNode == nil {
					child.fail = ac.root
				}
			}
			queue = append(queue, child)
		}
	}
}

func (ac *acAutoMachine) Query(content string) (results []string) {
	chars := []rune(content)
	iter := ac.root
	var start, end int
	for i, c := range chars {
		_, ok := iter.next[c]
		for !ok && iter != ac.root {
			iter = iter.fail
		}
		if _, ok = iter.next[c]; ok {
			if iter == ac.root { // this is the first match, record the start position
				start = i
			}
			iter = iter.next[c]
			if iter.isPattern {
				end = i // this is the end match, record one result
				results = append(results, string([]rune(content)[start:end+1]))
			}
		}
	}
	return
}
