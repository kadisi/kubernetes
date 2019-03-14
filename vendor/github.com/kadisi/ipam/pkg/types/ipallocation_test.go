/*
#  #############################################
#  Copyright (c) 2019-2039 All rights reserved.
#  #############################################
#
#  Name:  cache_test.go
#  Date:  2019-02-14 19:00
#  Author:   zhangjie
#  Email:   iamzhangjie0619@163.com
#  Desc:
#
*/

package types

import (
	"fmt"
	"net"
	"reflect"
	"sync"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWithIPAllocationSubnet(t *testing.T) {
	t.Run("1", func(t *testing.T) {

		ipallo := new(IPAllocation)

		sub := IPNet{
			net.IPNet{

				IP:   net.IPv4(192, 168, 10, 0),
				Mask: net.IPv4Mask(255, 255, 255, 0),
			},
		}

		WithIPAllocationSubnet(sub)(ipallo)

		if !reflect.DeepEqual(sub, ipallo.Subnet) {
			t.Errorf("subnet need quail")
		}
	})

}

func TestNewIPAllocation(t *testing.T) {

	tables := []struct {
		Name string
		Pod  *v1.Pod
		Get  bool
	}{
		{
			Name: "success",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "default",
					Annotations: map[string]string{
						AnnotationPodFloatingIP: "192.168.10.11",
						AnnotationPodConfigMap:  "floatingip",
						AnnotationPodSubnet:     "192.168.10.0/24",
						AnnotationPodGateway:    "192.168.10.1",
					},
				},
			},
			Get: true,
		},
		{
			Name: "no floatingip",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "default",
					Annotations: map[string]string{
						AnnotationPodConfigMap: "floatingip",
						AnnotationPodSubnet:    "192.168.10.0/24",
						AnnotationPodGateway:   "192.168.10.1",
					},
				},
			},
			Get: false,
		},

		{
			Name: "no configmap",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "default",
					Annotations: map[string]string{
						AnnotationPodFloatingIP: "192.168.10.11",
						AnnotationPodSubnet:     "192.168.10.0/24",
						AnnotationPodGateway:    "192.168.10.1",
					},
				},
			},
			Get: false,
		},
		{
			Name: "no subnet",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "default",
					Annotations: map[string]string{
						AnnotationPodFloatingIP: "192.168.10.11",
						AnnotationPodConfigMap:  "floatingip",
						AnnotationPodGateway:    "192.168.10.1",
					},
				},
			},
			Get: false,
		},

		{
			Name: "no gateway",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "default",
					Annotations: map[string]string{
						AnnotationPodFloatingIP: "192.168.10.11",
						AnnotationPodConfigMap:  "floatingip",
						AnnotationPodSubnet:     "192.168.10.0/24",
					},
				},
			},
			Get: false,
		},
	}

	for _, c := range tables {
		t.Run(c.Name, func(t *testing.T) {
			_, ok := NewIPAllocation(c.Pod,
				WithIPAllocationIP(net.IPv4(192, 168, 10, 11)))
			if ok != c.Get {
				t.Errorf("must be %v while get %v", c.Get, ok)
			}
		})
	}
}
func TestNewIPAllocationSet(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		ipas := NewIPAllocationSet(WithIPAllocationSetNamespace("default"))
		if ipas == nil {
			t.Errorf("ipas must not nil")
		}
	})
}
func TestWithIPAllocationCMName(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		ipallo := new(IPAllocation)
		cm := "floatingip"

		WithIPAllocationCMName(cm)(ipallo)

		if ipallo.CMName != cm {
			t.Errorf("cmname need equail")
		}
	})

}

func TestWithIPAllocationGateway(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		ipallo := new(IPAllocation)
		gw := net.IPv4(192, 168, 10, 1)

		WithIPAllocationGateway(gw)(ipallo)

		if !reflect.DeepEqual(ipallo.Gateway, gw) {
			t.Errorf("gateway need equail")
		}
	})

}

func TestWithIPAllocationIP(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		ipallo := new(IPAllocation)
		ip := net.IPv4(192, 168, 10, 1)

		WithIPAllocationIP(ip)(ipallo)

		if !reflect.DeepEqual(ipallo.IP, ip) {
			t.Errorf("ip need equail")
		}
	})

}

func TestIPAllocation_GetKey(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		ipallo := new(IPAllocation)
		ns := "default"
		cm := "floatingip"

		ip := net.IPv4(192, 168, 10, 1)
		ops := []IPAllocationField{
			WithIPAllocationNamespace(ns),
			WithIPAllocationCMName(cm),
			WithIPAllocationIP(ip)}

		for _, o := range ops {
			o(ipallo)
		}

		if ipallo.GetKey() != ip.String() {
			t.Errorf("get key need  ip string")
		}
	})

}
func TestWithIPAllocationNamespace(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		ipallo := new(IPAllocation)
		ns := "default"

		WithIPAllocationNamespace(ns)(ipallo)

		if ipallo.Namespace != ns {
			t.Errorf("namespace need equail")
		}
	})

}

func createipa(t *testing.T, i int) (*IPAllocation, bool) {

	var ipa *IPAllocation

	ns := "default"
	cm := "floatingip"
	podname := fmt.Sprintf("pod%v", i)

	ip := net.IPv4(192, 168, 10, byte(i))

	ipa, ok := NewIPAllocation(nil,
		WithIPAllocationNamespace(ns),
		WithIPAllocationCMName(cm),
		WithIPAllocationPodName(podname),
		WithIPAllocationIP(ip))
	if !ok {
		t.Errorf("must be ok")
		return nil, false
	}
	return ipa, true
}

func TestIPAllocationSet_AddUpdateIPAllocation(t *testing.T) {
	const count = 254

	t.Run("1", func(t *testing.T) {

		ipas := NewIPAllocationSet()

		group := sync.WaitGroup{}

		for i := 1; i < count; i++ {
			group.Add(1)
			go func(i int) {
				defer group.Done()

				ipallo, ok := createipa(t, i)
				if !ok {
					t.Errorf("wrong")
				}

				ipas.AddUpdateIPAllocation(ipallo)
			}(i)
		}
		group.Wait()
		group1 := sync.WaitGroup{}
		for i := 1; i < count; i++ {
			group1.Add(1)
			go func(i int) {
				defer group1.Done()

				ipallo, ok := createipa(t, i)
				if !ok {
					t.Errorf("wrong")
				}

				_, ok = ipas.GetIPAllocation(ipallo)
				if !ok {
					t.Errorf("key %v must ok must found", ipallo.GetKey())
					return
				}
			}(i)
		}

		group1.Wait()

	})

}

func TestIPAllocationSet_DeleteIPAllocation(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		ipas := NewIPAllocationSet()

		ipallo := new(IPAllocation)

		ns := "default"
		cm := "floatingip"
		podname := "pod1"
		ip := net.IPv4(192, 168, 10, 2)

		ops := []IPAllocationField{
			WithIPAllocationNamespace(ns),
			WithIPAllocationCMName(cm),
			WithIPAllocationPodName(podname),
			WithIPAllocationIP(ip),
		}

		for _, o := range ops {
			o(ipallo)
		}

		ipas.AddUpdateIPAllocation(ipallo)

		ipas.DeleteIPAllocation(ipallo)

		_, ok := ipas.GetIPAllocation(ipallo)
		if ok {
			t.Errorf("should not get this ipallocation")
		}
	})

}

func TestIPAllocationSet_GetIPAllocation(t *testing.T) {

	const count = 252

	t.Run("1", func(t *testing.T) {

		ipas := NewIPAllocationSet()

		group := sync.WaitGroup{}

		for i := 1; i < count; i++ {
			group.Add(1)
			go func(i int) {
				defer group.Done()

				ipallo, ok := createipa(t, i)
				if !ok {
					t.Errorf("wrong")
				}

				ipas.AddUpdateIPAllocation(ipallo)
			}(i)
		}
		group.Wait()

		group1 := sync.WaitGroup{}
		for i := 1; i < count; i++ {
			group1.Add(1)
			go func(i int) {
				defer group1.Done()

				ipallo, ok := createipa(t, i)
				if !ok {
					t.Errorf("wrong")
				}

				_, ok = ipas.GetIPAllocation(ipallo)
				if !ok {
					t.Errorf("must get , while it is %v", ok)
				}
			}(i)
		}

		group1.Wait()

	})
}

func TestIPAllocationSet_GetIPAllocationOperator(t *testing.T) {

	const count = 252

	t.Run("1", func(t *testing.T) {

		ipas := NewIPAllocationSet()

		group := sync.WaitGroup{}

		for i := 1; i < count; i++ {
			group.Add(1)
			go func(i int) {
				defer group.Done()

				ipallo, ok := createipa(t, i)
				if !ok {
					t.Errorf("wrong")
				}

				ipas.AddUpdateIPAllocation(ipallo)
			}(i)
		}
		group.Wait()

		group1 := sync.WaitGroup{}
		for i := 1; i < count; i++ {
			group1.Add(1)
			go func(i int) {
				defer group1.Done()

				ipallo, ok := createipa(t, i)
				if !ok {
					t.Errorf("wrong")
				}

				ipas.GetIPAllocationOperator(ipallo, func(ips *IPAllocation) {
					ips.Namespace = "zhangjie"
				}, func() {

				})
			}(i)
		}

		group1.Wait()

	})
}

func TestIPAllocationSet_GetKey(t *testing.T) {

	ns := "default"
	cm := "floating"

	t.Run("1", func(t *testing.T) {
		ipas := NewIPAllocationSet(WithIPAllocationSetNamespace(ns),
			WithIPAllocationSetCMName(cm))

		if ipas.GetKey() != fmt.Sprintf("%v_%v", ns, cm) {
			t.Errorf("key need namespace and cmname")
		}
	})

}

func TestWithIPAllocationSetFields(t *testing.T) {
	t.Run("1", func(t *testing.T) {

		ns := "default"
		cm := "floatingip"

		ipas := NewIPAllocationSet()

		WithIPAllocationSetFields(ipas,
			WithIPAllocationSetCMName(cm),
			WithIPAllocationSetNamespace(ns))

		if ipas.CMName != cm {
			t.Errorf("cm name not equal")
		}

		if ipas.Namespace != ns {
			t.Errorf("namespace not equal")
		}

	})
}

func TestSubIP(t *testing.T) {

	tables := []struct {
		Name   string
		First  net.IP
		Second net.IP
		Value  int
	}{
		{
			Name:   "20-10",
			First:  net.IPv4(192, 168, 10, 20),
			Second: net.IPv4(192, 168, 10, 10),
			Value:  10,
		},
		{
			Name:   "20-20",
			First:  net.IPv4(192, 168, 10, 20),
			Second: net.IPv4(192, 168, 10, 20),
			Value:  0,
		},
		{
			Name:   "10-20",
			First:  net.IPv4(192, 168, 10, 10),
			Second: net.IPv4(192, 168, 10, 20),
			Value:  -10,
		},
	}

	for _, c := range tables {
		t.Run(c.Name, func(t *testing.T) {
			get := SubIP(c.First, c.Second)
			if get != c.Value {
				t.Errorf("need %v now get %v", c.Value, get)
			}
		})
	}
}

func TestPrevIP(t *testing.T) {
	tables := []struct {
		Name string
		IP   net.IP
		PRE  net.IP
	}{
		{
			Name: "20",
			IP:   net.IPv4(192, 168, 10, 20),
			PRE:  net.IPv4(192, 168, 10, 19),
		},
		{
			Name: "0",
			IP:   net.IPv4(192, 168, 10, 0),
			PRE:  net.IPv4(192, 168, 9, 255),
		},
		{
			Name: "255",
			IP:   net.IPv4(192, 168, 10, 255),
			PRE:  net.IPv4(192, 168, 10, 254),
		},
	}

	for _, c := range tables {
		t.Run(c.Name, func(t *testing.T) {
			get := PrevIP(c.IP)
			if Cmp(get, c.PRE) != 0 {
				t.Errorf("need %v now get %v", c.PRE, get)
			}
		})
	}
}

func TestNextIP(t *testing.T) {
	tables := []struct {
		Name string
		IP   net.IP
		Next net.IP
	}{
		{
			Name: "20",
			IP:   net.IPv4(192, 168, 10, 20),
			Next: net.IPv4(192, 168, 10, 21),
		},
		{
			Name: "255",
			IP:   net.IPv4(192, 168, 10, 255),
			Next: net.IPv4(192, 168, 11, 0),
		},
		{
			Name: "0",
			IP:   net.IPv4(192, 168, 10, 0),
			Next: net.IPv4(192, 168, 10, 1),
		},
	}

	for _, c := range tables {
		t.Run(c.Name, func(t *testing.T) {
			get := NextIP(c.IP)
			if Cmp(get, c.Next) != 0 {
				t.Errorf("need %v now get %v", c.Next, get)
			}
		})
	}
}

func BenchmarkIPAllocationSet_AddUpdateIPAllocation(b *testing.B) {

	b.Run("1", func(b *testing.B) {

		ipas := NewIPAllocationSet(WithIPAllocationSetNamespace("default"),
			WithIPAllocationSetCMName("floatingip"))

		ipa, ok := NewIPAllocation(nil,
			WithIPAllocationPodName("pod"),
			WithIPAllocationCMName("floatingip"),
			WithIPAllocationNamespace("default"),
			WithIPAllocationIP(net.IPv4(192, 168, 10, 11)),
			WithIPAllocationGateway(net.IPv4(192, 168, 10, 1)),
			WithIPAllocationSubnet(
				IPNet{
					net.IPNet{
						IP:   net.IPv4(192, 168, 10, 0),
						Mask: net.IPv4Mask(255, 255, 255, 0),
					},
				}))
		if !ok {
			b.Errorf("need ok when create ipallocation")
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ipas.AddUpdateIPAllocation(ipa)
		}
	})
}
