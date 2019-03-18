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
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParseCIDR(t *testing.T) {

}

func TestWithIpamCMName(t *testing.T) {
	name := "floatingip"
	ipam := new(Ipam)

	WithIpamCMName(name)(ipam)

	if ipam.CMName != name {
		t.Errorf("cmname need equal")
	}
}

func TestWithIpamNamespace(t *testing.T) {
	ns := "default"
	ipam := new(Ipam)

	WithIpamNamespace(ns)(ipam)

	if ipam.Namespace != ns {
		t.Errorf("namespace need equal")
	}
}

func TestWithIpamRange(t *testing.T) {
	rang := new(Range)

	ipam := new(Ipam)
	WithIpamRange(rang)(ipam)

	if ipam.Range != rang {
		t.Errorf("range need equal")
	}
}

func TestIpam_GetKey(t *testing.T) {
	ns := "default"
	cmname := "floatingip"
	ipam := new(Ipam)

	ops := []IpamField{
		WithIpamNamespace(ns),
		WithIpamCMName(cmname)}

	for _, o := range ops {
		o(ipam)
	}

	if ipam.GetKey() != fmt.Sprintf("%v_%v", ns, cmname) {
		t.Errorf("getkey not equil")
	}
}

func TestNewIpam(t *testing.T) {
	type testCase1 struct {
		Name string
		CM   *v1.ConfigMap
		Get  bool
	}
	cases := []testCase1{
		{
			Name: "success",
			CM: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test1",
				},
				Data: map[string]string{
					ConfigMapFloatingIPKey: `{"range":{"rangeStart":"192.168.10.10",` +
						`"rangeEnd":"192.168.10.20","subnet":` +
						`"192.168.10.0/24","gateway":"192.168.10.1"}}`,
				},
			},
			Get: true,
		},
		{
			Name: "no floatingip key",
			CM: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test1",
				},
				Data: map[string]string{
					"sssssssss": `{"range":{"rangeStart":"192.168.10.10",` +
						`"rangeEnd":"192.168.10.20","subnet":` +
						`"192.168.10.0/24","gateway":"192.168.10.1"}}`,
				},
			},
			Get: false,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, ok := NewIpam(c.CM)
			if ok != c.Get {
				t.Errorf("expect %v while %v", c.Get, ok)
			}
		})
	}

	// vlan
	func() {
		c := testCase1{
			Name: "no-vlan-success",
			CM: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test1",
				},
				Data: map[string]string{
					ConfigMapFloatingIPKey: `{"range":{"rangeStart":"192.168.10.10",` +
						`"rangeEnd":"192.168.10.20","subnet":` +
						`"192.168.10.0/24","gateway":"192.168.10.1"}}`,
				},
			},
			Get: true,
		}

		t.Run(c.Name, func(t *testing.T) {
			a, ok := NewIpam(c.CM)
			if ok != c.Get {
				t.Errorf("expect %v while %v", c.Get, ok)
			}

			if a.Range.Vlan != NOVlanStr {
				t.Errorf("expect get vlan %v now %v", NOVlanStr, a.Range.Vlan)
			}
		})
	}()

	// routes
	func() {
		c := testCase1{
			Name: "route-success",
			CM: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test1",
				},
				Data: map[string]string{
					ConfigMapFloatingIPKey: `{"range":{"rangeStart":"192.168.10.10",` +
						`"rangeEnd":"192.168.10.20","subnet":` +
						`"192.168.10.0/24","gateway":"192.168.10.1"},` +
						`"routes":[ {"dst": "192.168.0.0/16"}]}`,
				},
			},
			Get: true,
		}

		t.Run(c.Name, func(t *testing.T) {
			a, ok := NewIpam(c.CM)
			if ok != c.Get {
				t.Errorf("expect %v while %v", c.Get, ok)
			}

			if a.Routes == nil {
				t.Errorf("can not get route info ")
			}

			if len(a.Routes) != 1 {
				t.Errorf("len of Routes != 1")
			}

			if a.Routes[0].Dst.String() != "192.168.0.0/16" {
				t.Errorf("get routes dst error")
			}

		})

	}()

}

const (
	CMPExpertLower   = -1
	CMPExpertEqual   = 0
	CMPExpertGreater = 1
)

func TestCmp(t *testing.T) {
	table := []struct {
		Name   string
		IP1    net.IP
		IP2    net.IP
		Expert int
	}{
		{
			Name:   "1",
			IP1:    net.IPv4(192, 168, 10, 1),
			IP2:    net.IPv4(192, 168, 10, 1),
			Expert: CMPExpertEqual,
		},
		{
			Name:   "2",
			IP1:    net.IPv4(192, 168, 10, 1),
			IP2:    net.IPv4(192, 168, 10, 2),
			Expert: CMPExpertLower,
		},
		{
			Name:   "3",
			IP1:    net.IPv4(192, 168, 10, 2),
			IP2:    net.IPv4(192, 168, 10, 1),
			Expert: CMPExpertGreater,
		},
	}

	for _, ta := range table {
		t.Run(ta.Name, func(t *testing.T) {
			if Cmp(ta.IP1, ta.IP2) != ta.Expert {
				t.Errorf("CMP expert %v", ta.Expert)
			}
		})
	}
}
