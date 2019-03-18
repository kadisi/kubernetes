/*
#  #############################################
#  Copyright (c) 2019-2039 All rights reserved.
#  #############################################
#
#  Name:  wocloud.go
#  Date:  2019-02-21 15:44
#  Author:   zhangjie
#  Email:   iamzhangjie0619@163.com
#  Desc:
#
*/

package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kadisi/ipam/api/services/ipams"
	ipamtypes "github.com/kadisi/ipam/pkg/types"

	"k8s.io/apimachinery/pkg/types"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

const (
	// AnnotationCMFloatingIP is in configmap annotation
	AnnotationCMFloatingIP = ipamtypes.AnnotationCMFloatingIP

	// ConfigMapFloatingIPKey is in configmap data key
	ConfigMapFloatingIPKey = ipamtypes.ConfigMapFloatingIPKey

	// TrueStr is true string
	TrueStr = ipamtypes.TrueStr

	// stand this pod need Floatingip
	AnnotationPodNeedFloatingIP = "wocloud.cn/floatingip"

	// AnnotationPodFloatingIP is in pod annotation
	AnnotationPodFloatingIP = ipamtypes.AnnotationPodFloatingIP
	// AnnotationPodSubnet is in pod annotation
	AnnotationPodSubnet = ipamtypes.AnnotationPodSubnet
	// AnnotationPodGateway is in pod annotation
	AnnotationPodGateway = ipamtypes.AnnotationPodGateway
	// AnnotationPodConfigMap is in pod annotation
	AnnotationPodConfigMap = ipamtypes.AnnotationPodConfigMap

	AnnotationPodRoutes = ipamtypes.AnnotationPodRoutes

	AnnotationPodVlan = ipamtypes.AnnotationPodVlan
)

type Ipamer interface {
	AssiginFloattingIP(pod *v1.Pod) error
}

type WoclouderClient struct {
	IpamServiceClient ipams.IpamServiceClient
	ConfigmapLister   corelisters.ConfigMapLister
	Client            clientset.Interface
}

func (c *WoclouderClient) AssiginFloattingIP(pod *v1.Pod) error {

	need, ok := pod.GetAnnotations()[AnnotationPodNeedFloatingIP]
	if !ok || need != TrueStr {
		// Do nothing
		return nil
	}

	cachedcms, err := c.ConfigmapLister.List(labels.Everything())
	if err != nil {
		glog.V(3).Info("can not find any configmap in cached")
		return fmt.Errorf("can not find any configmap in cached")
	}

	keyFunc := func(ns, name string) string {
		return fmt.Sprintf("%v_%v", ns, name)
	}

	cachemap := make(map[string]struct{})
	for _, cm := range cachedcms {
		cachemap[keyFunc(cm.GetNamespace(), cm.GetName())] = struct{}{}
	}

	requestcms := make([]string, 0, 2)
	for _, v := range pod.Spec.Volumes {
		if v.ConfigMap != nil {
			if _, ok := cachemap[keyFunc(pod.GetNamespace(), v.ConfigMap.Name)]; ok {
				glog.V(3).Infof("find pod %v in ns %v Volumes has floatingip configmap %v ",
					pod.GetName(), pod.GetNamespace(), v.ConfigMap.Name)
				requestcms = append(requestcms, v.ConfigMap.Name)
			}
		}
	}
	if len(requestcms) == 0 {
		// Do nothing
		return fmt.Errorf("can not find any configmap in pod[%s][%s].spec.volumes ", pod.GetNamespace(), pod.GetName())
	}

	ctx, cancle := context.WithTimeout(context.Background(), time.Second*5)
	defer cancle()

	respon, err := c.IpamServiceClient.AcquireIP(ctx, &ipams.AcquireIPRequest{
		Podname:    pod.GetName(),
		Namespace:  pod.GetNamespace(),
		ConfigMaps: requestcms,
	})
	if err != nil {
		glog.V(1).Infof("rpc client acquireip error %v", err)
		return fmt.Errorf("rpc client acquireip error %v", err)
	}

	addAnnotationPatch := func(ip, subnet, gw, cm, vlan, routes string) ([]byte, error) {

		type Metadata struct {
			annotations map[string]string `json:"annotations"`
		}
		type mergedata struct {
			metadata Metadata `json:"metadata"`
		}
		annotation := make(map[string]string)

		annotation[AnnotationPodFloatingIP] = ip
		annotation[AnnotationPodSubnet] = subnet
		annotation[AnnotationPodGateway] = gw
		annotation[AnnotationPodConfigMap] = cm
		annotation[AnnotationPodVlan] = vlan
		annotation[AnnotationPodRoutes] = routes

		mergepod := &mergedata{
			metadata: Metadata{annotations: annotation},
		}

		return json.Marshal(mergepod)
	}
	routes, err := json.Marshal(respon.Ipaminfo.Routes)
	if err != nil {
		glog.V(1).Infof("json marshal error %v", err)
		return err
	}
	mergeannotation, err := addAnnotationPatch(
		respon.Ipaminfo.Ip, respon.Ipaminfo.Subnet, respon.Ipaminfo.Gateway, respon.Ipaminfo.ConfigMap,
		respon.Ipaminfo.Vlan, string(routes))
	if err != nil {
		glog.V(3).Infof("marshal annotation error %v", err)
		return err
	}
	_, err = c.Client.CoreV1().Pods(pod.GetNamespace()).Patch(pod.Name, types.MergePatchType, mergeannotation)
	if err != nil {
		glog.V(3).Infof("patch pod annotation for floatingip error %v mergeannotation %v", err, string(mergeannotation))
		return err
	}

	return nil
}
