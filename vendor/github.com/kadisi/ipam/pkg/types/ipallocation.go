package types

import (
	"fmt"
	"net"
	"sync"

	"github.com/kadisi/ipam/log"

	"k8s.io/api/core/v1"
)

const (
	// AnnotationPodFloatingIP is in pod annotation
	AnnotationPodFloatingIP = "wocloud.cn/floating-ip"
	// AnnotationPodSubnet is in pod annotation
	AnnotationPodSubnet = "wocloud.cn/floating-subnet"
	// AnnotationPodGateway is in pod annotation
	AnnotationPodGateway = "wocloud.cn/floating-gateway"
	// AnnotationPodConfigMap is in pod annotation
	AnnotationPodConfigMap = "wocloud.cn/floating-configmap"
)

// IPAllocation stand allocation ips
type IPAllocation struct {
	IP        net.IP `json:"Ip,omitempty"`
	Subnet    IPNet  `json:"subnet"`
	Gateway   net.IP `json:"gateway,omitempty"`
	Namespace string `json:"namespace"`
	CMName    string `json:"cmname"`
	PodName   string `json:"podname"`
}

// GetKey is get ipallocation key return string
func (ipa *IPAllocation) GetKey() string {
	k := ipa.IP.String()
	if k == "<nil>" {
		log.Fatalf("key(IP) of IPAllocation %++v is nil ", *ipa)
	}
	return k
}

// IPAllocationField is alias func(*IPAllocation)
type IPAllocationField func(*IPAllocation)

// WithIPAllocationFields is set ipallocation by IPAllocationFields
func WithIPAllocationFields(ipa *IPAllocation, ops ...IPAllocationField) {
	for _, o := range ops {
		o(ipa)
	}
}

// WithIPAllocationIP set ip field
func WithIPAllocationIP(ip net.IP) IPAllocationField {
	return func(allocation *IPAllocation) {
		allocation.IP = ip
	}
}

// WithIPAllocationSubnet set subnet field
func WithIPAllocationSubnet(subnet IPNet) IPAllocationField {
	return func(allocation *IPAllocation) {
		allocation.Subnet = subnet
	}
}

// WithIPAllocationGateway set gateway field
func WithIPAllocationGateway(gateway net.IP) IPAllocationField {
	return func(allocation *IPAllocation) {
		allocation.Gateway = gateway
	}
}

// WithIPAllocationNamespace set ns field
func WithIPAllocationNamespace(ns string) IPAllocationField {
	return func(a *IPAllocation) {
		a.Namespace = ns
	}
}

// WithIPAllocationCMName set cm field
func WithIPAllocationCMName(cm string) IPAllocationField {
	return func(allocation *IPAllocation) {
		allocation.CMName = cm
	}
}

// WithIPAllocationPodName set podname field
func WithIPAllocationPodName(pname string) IPAllocationField {
	return func(allocation *IPAllocation) {
		allocation.PodName = pname
	}
}

// NewIPAllocation create new IPAllocation
func NewIPAllocation(p *v1.Pod,
	options ...IPAllocationField) (*IPAllocation, bool) {

	ipa := new(IPAllocation)
	if p != nil {
		floatingip, ok := p.GetAnnotations()[AnnotationPodFloatingIP]
		if !ok {
			return nil, false
		}

		gateway, ok := p.GetAnnotations()[AnnotationPodGateway]
		if !ok {
			return nil, false
		}

		subnet, ok := p.GetAnnotations()[AnnotationPodSubnet]
		if !ok {
			return nil, false
		}

		cmname, ok := p.GetAnnotations()[AnnotationPodConfigMap]
		if !ok {
			return nil, false
		}

		sub, err := ParseCIDR(subnet)
		if err != nil {
			return nil, false
		}

		options = append(options, WithIPAllocationPodName(p.GetName()),
			WithIPAllocationNamespace(p.GetNamespace()),
			WithIPAllocationCMName(cmname),
			WithIPAllocationGateway(net.ParseIP(gateway).To4()),
			WithIPAllocationIP(net.ParseIP(floatingip).To4()),
			WithIPAllocationSubnet(IPNet{*sub}),
		)
	}

	WithIPAllocationFields(ipa, options...)
	return ipa, true
}

// IPAllocationSet stand for a group of IPAllocation
type IPAllocationSet struct {

	// IPAllocationsMap Map or Index For IPAllocations,
	// key is ip string, value is *IPAllocation
	IPAllocationsMap map[string]*IPAllocation

	// Add Lock When change IPAllocations
	Lock *sync.Mutex

	// Namespace and CMName is key of IPAllocationSet
	Namespace string `json:"namespace,omitempty"`
	CMName    string `json:"cnname,omitempty"`
}

// NewIPAllocationSet get new IPAllocationSet obj
func NewIPAllocationSet(ops ...IPAllocationSetField) *IPAllocationSet {
	set := &IPAllocationSet{
		IPAllocationsMap: make(map[string]*IPAllocation),
		Lock:             new(sync.Mutex),
	}

	WithIPAllocationSetFields(set, ops...)
	return set
}

// IPAllocationSetField is func(*IPAllocationSet)
type IPAllocationSetField func(*IPAllocationSet)

// WithIPAllocationSetNamespace is set namespace
func WithIPAllocationSetNamespace(ns string) IPAllocationSetField {
	return func(set *IPAllocationSet) {
		set.Namespace = ns
	}
}

// WithIPAllocationSetCMName set configmap name
func WithIPAllocationSetCMName(cm string) IPAllocationSetField {
	return func(set *IPAllocationSet) {
		set.CMName = cm
	}
}

// WithIPAllocationSetFields set fields
func WithIPAllocationSetFields(set *IPAllocationSet,
	ops ...IPAllocationSetField) {
	for _, o := range ops {
		o(set)
	}
}

// GetIPAllocationOperator is get ipallocation by meta
// found is when find, do someting
// not is when not find , do something
// found and not are all locked
func (ipas *IPAllocationSet) GetIPAllocationOperator(meta MetaData,
	found func(ips *IPAllocation), not func()) {
	ipas.Lock.Lock()
	defer ipas.Lock.Unlock()

	_, ok := meta.(*IPAllocation)
	if !ok {
		log.Fatal("meta need *IPAllocation")
		return
	}
	ipa, ok := ipas.IPAllocationsMap[meta.GetKey()]
	if ok && found != nil {
		found(ipa)
	} else if not != nil {
		not()
	}

}

// GetIPAllocation is get ipallocation by meta
// it has been locked
func (ipas *IPAllocationSet) GetIPAllocation(meta MetaData) (
	*IPAllocation, bool) {

	ipas.Lock.Lock()
	defer ipas.Lock.Unlock()

	_, ok := meta.(*IPAllocation)
	if !ok {
		log.Warnf("meta need *IPAllocation type")
		return nil, ok
	}
	ipa, ok := ipas.IPAllocationsMap[meta.GetKey()]
	return ipa, ok
}

// AddUpdateIPAllocation is Add Or Update IPAllocation To IPAllocationSet
func (ipas *IPAllocationSet) AddUpdateIPAllocation(ia *IPAllocation) {
	ipas.Lock.Lock()
	defer ipas.Lock.Unlock()

	ipas.IPAllocationsMap[ia.GetKey()] = ia
}

// DeleteIPAllocation is delete IPAllocation by key
func (ipas *IPAllocationSet) DeleteIPAllocation(meta MetaData) {
	ipas.Lock.Lock()
	defer ipas.Lock.Unlock()

	_, ok := meta.(*IPAllocation)
	if !ok {
		log.Fatal("meta need *IPAllocation type")
		return
	}

	delete(ipas.IPAllocationsMap, meta.GetKey())
}

// GetKey is get ipallocationset key
func (ipas *IPAllocationSet) GetKey() string {
	if ipas.Namespace == "" || ipas.CMName == "" {
		log.Fatalf("IPAllocationSet %++V key param Namespace or CMName is nil", *ipas)
	}
	return fmt.Sprintf("%v_%v", ipas.Namespace, ipas.CMName)
}
