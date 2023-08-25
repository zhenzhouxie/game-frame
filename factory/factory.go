package factory

import (
	"fmt"
	"testing"
	"time"
)

type FlowMsg struct {
	From  string
	MsgId int
	Data  []byte
}
type DefaultNode struct {
	Name string
	C    chan any
	Last []Node
	Next []Node
}

func NewDefaultNode(name string) DefaultNode {
	return DefaultNode{
		Name: name,
		C:    make(chan any, 1000),
		Last: make([]Node, 0),
		Next: make([]Node, 0),
	}
}

func (i *DefaultNode) work() {
	go func() {
		for {
			for _, v := range i.Last {
				ret := v.read().(int)
				i.send(ret * ret)
			}
			time.Sleep(time.Second * 1)
		}
	}()
}

func (i *DefaultNode) read() any {
	select {
	case r := <-i.C:
		return r
	default:
		return nil
	}
}

func (i *DefaultNode) send(msg any) {
	i.C <- msg
}

func (i *DefaultNode) append(next Node) Node {
	if i.Next == nil {
		i.Next = make([]Node, 0)
	}
	i.Next = append(i.Next, next)
	next.link(i)
	return next
}

func (i *DefaultNode) getName() string {
	return i.Name
}

func (i *DefaultNode) link(last Node) {
	if i.Last == nil {
		i.Last = make([]Node, 0)
	}
	i.Last = append(i.Last, last)
}

func (i *DefaultNode) getCap() int {
	return cap(i.C)
}
func (i *DefaultNode) getLen() int {
	return len(i.C)
}

type Node interface {
	getName() string
	getCap() int
	getLen() int
	read() any
	append(next Node) Node
	link(last Node)
	send(msg any)
	work()
}

type Factory struct {
	Name     string
	NodeMap  map[string]Node   //工厂具体要的所有工序
	FlowLink map[string][]Node //所有流水线
}

func (f *Factory) StartAllNode() {
	for _, v := range f.NodeMap {
		v.work()
	}
}

func (f *Factory) StartAllFlowLink() {
	for _, v := range f.NodeMap {
		v.work()
	}
}

func (f *Factory) addFlowLink(nodeList ...Node) {
	if len(nodeList) == 0 {
		return
	}
	f.addNode(nodeList[0])
	for i := 0; i < len(nodeList)-1; i++ {
		f.addNode(nodeList[i+1])
		thisNode := nodeList[i]
		nextNode := nodeList[i+1]
		thisNode.append(nextNode)
	}
}

func (f *Factory) addNode(node Node) {
	if _, ok := f.NodeMap[node.getName()]; !ok {
		f.NodeMap[node.getName()] = node
	}
}

type InputInt struct {
	DefaultNode
}

func (i *InputInt) work() {
	go func() {
		start := time.Now()
		for n := 0; n < 10000; n++ {
			i.send(n)
			time.Sleep(time.Nanosecond * 1)
		}
		fmt.Println("Use TimeTime", time.Now().UnixMilli()-start.UnixMilli())
	}()
}

type OuputNode struct {
	DefaultNode
}

func (i *OuputNode) NeedSleep() bool {
	for _, v := range i.Last {
		if v.getLen() > 0 {
			return false
		}
	}
	return true
}

func (i *OuputNode) work() {
	go func() {
		for {
			if i.NeedSleep() {
				time.Sleep(time.Second * 1)
			}
			for _, v := range i.Last {
				ret := v.read()
				if ret != nil {
					fmt.Printf("from %s read %d \n", v.getName(), ret.(int))
				}
			}
		}
	}()
}
func NewDefaultOuput() OuputNode {
	return OuputNode{
		DefaultNode: NewDefaultNode("output"),
	}
}

// mock data stream Socket -> worker  and kafka-consumer -> worker
func TestFlow(t *testing.T) {
	factory := Factory{
		NodeMap: make(map[string]Node, 0),
	}
	ouput := NewDefaultOuput()
	worker := NewDefaultNode("worker")

	input1 := InputInt{
		DefaultNode: NewDefaultNode("Socket"),
	}
	input2 := InputInt{
		DefaultNode: NewDefaultNode("KafKa Consumer"),
	}

	factory.addFlowLink(&input1, &worker, &ouput)
	factory.addFlowLink(&input2, &worker)

	factory.StartAllNode()
	time.Sleep(500 * time.Second)
}

// type WorkFun func(n Node)

// 自动扩缩容节点
type DilatationNode struct {
	DefaultNode
	Brother []*DilatationNode
	IsNode0 bool
	IsStop  bool
}

// 扩容函数
func (d *DilatationNode) dilatation() {
	for _, v := range d.Brother {
		if v.IsStop {
			v.work()
			return
		}
	}

	name := fmt.Sprint("dilatation-", len(d.Brother)+1)
	newNode := DilatationNode{
		DefaultNode: NewDefaultNode(name),
		IsNode0:     false,
	}
	for _, last := range d.Last {
		last.append(&newNode)
	}
	for _, next := range d.Next {
		newNode.append(next)
	}

	newNode.work()
	d.Brother = append(d.Brother, &newNode)
	fmt.Println("dilatation+++, Node:", newNode)
}

func (d *DilatationNode) GetLastStatus() float64 {
	totalCap := 0
	totalLen := 0
	for _, last := range d.Last {
		totalCap += last.getCap()
		totalLen += last.getLen()
	}

	return float64(totalLen) / float64(totalCap)
}

func (d *DilatationNode) work() {
	go func() {
		d.IsStop = false
		for {
			if d.IsStop {
				return
			}
			for _, v := range d.Last {
				ret := v.read()
				if ret != nil {
					d.send(ret.(int))
				}
			}
		}
	}()

	if !d.IsNode0 {
		return
	}

	go func() {
		ticker := time.NewTicker(time.Second * 2)
		for range ticker.C {
			status := d.GetLastStatus()
			runB := d.GetRunB()
			if status > float64(runB)/float64(runB+1) {
				d.dilatation()
			}
			if status < float64(runB-1)/float64(runB+1) {
				d.StopOne()
			}
			fmt.Println("status:  ", status, "runB : ", runB)
		}
	}()
}

func (d *DilatationNode) GetRunB() int {
	ret := 0
	for _, v := range d.Brother {
		if !v.IsStop {
			ret += 1
		}
	}
	return ret + 1
}

func (d *DilatationNode) StopOne() {
	for i := len(d.Brother) - 1; i >= 0; i-- {
		d := d.Brother[i]
		if !d.IsStop {
			d.IsStop = true
			return
		}
	}
}

func TestDilatation(t *testing.T) {
	ouput := NewDefaultOuput()
	worker := DilatationNode{
		DefaultNode: NewDefaultNode("dilatation-0"),
		IsNode0:     true,
	}

	input := InputInt{
		DefaultNode: NewDefaultNode("Socket"),
	}

	factory := Factory{
		NodeMap: make(map[string]Node, 0),
	}
	factory.addFlowLink(&input, &worker, &ouput)
	factory.StartAllNode()

	for {
		time.Sleep(20 * time.Second)
	}
}

func TestAAA(t *testing.T) {
	start := time.Now()
	for n := 0; n < 1000000; n++ {
		fmt.Println(n)
		time.Sleep(time.Nanosecond * 1)
	}
	fmt.Println(time.Now().UnixMilli() - start.UnixMilli())
}
