package types

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net"

	"github.com/kadisi/ipam/log"
	"k8s.io/api/core/v1"
)

const (
	// AnnotationCMFloatingIP is in configmap annotation
	AnnotationCMFloatingIP = "wocloud.cn/floatingip"

	// ConfigMapFloatingIPKey is in configmap data key
	ConfigMapFloatingIPKey = "ipam"

	// TrueStr is true string
	TrueStr = "true"
)

// IPNet is net.IPNet
type IPNet struct {
	net.IPNet
}

// Cmp compares two IPs, returning the usual ordering:
// a < b : -1
// a == b : 0
// a > b : 1
func Cmp(a, b net.IP) int {
	aa := ipToInt(a)
	bb := ipToInt(b)
	return aa.Cmp(bb)
}

// SubIP return e - s
func SubIP(e, s net.IP) int {
	ee := ipToInt(e)
	ss := ipToInt(s)

	sub := big.NewInt(0)
	sub.Sub(ee, ss)

	return int(sub.Int64())
}

// NextIP returns IP incremented by 1
func NextIP(ip net.IP) net.IP {
	i := ipToInt(ip)
	return intToIP(i.Add(i, big.NewInt(1)))
}

// PrevIP returns IP decremented by 1
func PrevIP(ip net.IP) net.IP {
	i := ipToInt(ip)
	return intToIP(i.Sub(i, big.NewInt(1)))
}

func ipToInt(ip net.IP) *big.Int {
	if v := ip.To4(); v != nil {
		return big.NewInt(0).SetBytes(v)
	}
	return big.NewInt(0).SetBytes(ip.To16())
}

func intToIP(i *big.Int) net.IP {
	return net.IP(i.Bytes())
}

// MarshalJSON for json.Marshal
func (n IPNet) MarshalJSON() ([]byte, error) {
	return json.Marshal((&n).String())
}

// UnmarshalJSON for json.Unmarshal
func (n *IPNet) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	tmp, err := ParseCIDR(s)
	if err != nil {
		return err
	}

	*n = IPNet{*tmp}
	return nil
}

// ParseCIDR takes a string like "10.2.3.1/24" and
// return IPNet with "10.2.3.1" and /24 mask
func ParseCIDR(s string) (*net.IPNet, error) {
	ip, ipn, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}

	ipn.IP = ip
	return ipn, nil
}

// Range is Rang for ipam
type Range struct {
	RangeStart net.IP `json:"rangeStart,omitempty"` // The first ip, inclusive
	RangeEnd   net.IP `json:"rangeEnd,omitempty"`   // The last ip, inclusive
	Subnet     IPNet  `json:"subnet"`
	Gateway    net.IP `json:"gateway,omitempty"`
}

// TODO need add route struct
// TODO need add vlan , default no vlan

// Ipam  is
type Ipam struct {
	Range     *Range `json:"range,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	CMName    string `json:"cnname,omitempty"`
}

// NewIpam create new Ipam obj by V1.ConfigMap
func NewIpam(cm *v1.ConfigMap, opts ...IpamField) (*Ipam, bool) {

	ipam := new(Ipam)

	if cm != nil {
		data, ok := cm.Data[ConfigMapFloatingIPKey]
		if !ok {
			return nil, false
		}

		err := json.Unmarshal([]byte(data), ipam)
		if err != nil {
			log.Debugf("configmap date unmarshal type.Ipam error %v", err)
			return nil, false
		}

		opts = append(opts,
			WithIpamNamespace(cm.GetNamespace()),
			WithIpamCMName(cm.GetName()),
		)
	}
	WithIpamFields(ipam, opts...)

	return ipam, true

}

// IpamField alias func(*Ipam)
type IpamField func(*Ipam)

// WithIpamFields set ipam by IpamFields
func WithIpamFields(ipam *Ipam, ips ...IpamField) {
	for _, o := range ips {
		o(ipam)
	}
}

// WithIpamNamespace set ns field
func WithIpamNamespace(ns string) IpamField {
	return func(ipam *Ipam) {
		ipam.Namespace = ns
	}
}

// WithIpamCMName set cm field
func WithIpamCMName(cm string) IpamField {
	return func(ipam *Ipam) {
		ipam.CMName = cm
	}
}

// WithIpamRange set range field
func WithIpamRange(r *Range) IpamField {
	return func(ipam *Ipam) {
		ipam.Range = r
	}
}

// GetKey get ipam Key
func (ipa *Ipam) GetKey() string {
	if ipa.Namespace == "" || ipa.CMName == "" {
		log.Fatalf("IPAM %++v key param NameSpace or CMName is nil", *ipa)
	}
	return fmt.Sprintf("%v_%v", ipa.Namespace, ipa.CMName)
}

// MetaData is interface
type MetaData interface {
	GetKey() string
}
