package datastructure


type LinkedList struct {
	Head *Node
	End *Node
}

type Node struct {
	Next *Node
	Val interface{}
}


func (ll *LinkedList) InsertAtLast(node Node) {
	if ll.Head == nil {
		ll.Head = &node
	}
	temp := ll.Head
	for {
		if temp.Next == nil {
			temp.Next = &node
		}
		temp = temp.Next
	}
}

func (ll *LinkedList) InsertAtBeginning(node Node) {

}