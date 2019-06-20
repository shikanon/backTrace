package backTrace

import (
	"fmt"
)

// 调度接口
type Scheduler interface {
	//入参是code数组，生成并启动Task
	schedulerTask([]string)
}

// 单机调度器
type LocalScheduler struct {
	backend         LocalBackend
	coresForPerTask int8
	cacheMap        StockMap
	runningTasks    map[string]bool
}

//根据股票ID进行任务调度
func (sc *LocalScheduler) schedulerTask([]string) {
	//m个买策略与n个卖策略 笛卡尔积得到 m * n个 anaylise

	//m * n 个ana 与 h个 code 笛卡尔

}

//资源管理接口
type backend interface {
	register(hostName string, ip string, port int32, core int8) (bool, error)
	leave(nodeId string) (bool, error)
	getAliveNode() []*Node
}

//单机版的计算资源管理
type LocalBackend struct {
	node Node
}

//注册节点
func (lb *LocalBackend) register(hostName string, ip string, port int32, core int8) (bool, error) {
	if lb.node == (Node{}) {
		lb.node = NewNode(hostName, ip, port, core)
		return true, nil
	} else {
		return false, fmt.Errorf(" LocalNode %s is already registered!", lb.node)
	}
}

func (lb *LocalBackend) leave(nodeId string) (bool, error) {
	if lb.node.nodeId != nodeId {
		return false, fmt.Errorf(" There is no node named %s .", nodeId)
	} else {
		lb.node = Node{}
		return true, nil
	}
}

func (lb *LocalBackend) getAliveNode() []*Node {
	return []*Node{&lb.node}
}

//节点信息，表示一个节点
type Node struct {
	nodeId   string
	hostName string
	core     int8
	ip       string
	port     int32
}

func (n *Node) string() string {
	return n.nodeId
}

//创建节点的方法
func NewNode(hostName string, ip string, port int32, core int8) Node {
	return Node{
		nodeId:   fmt.Sprintf("%s_%s_%d", hostName, ip, port),
		hostName: hostName,
		ip:       ip,
		port:     port,
		core:     core,
	}
}

// ---------------------------Task
func runTask(ana Analyzer, stockCode string) (bool, error) {
	return true, nil
}
