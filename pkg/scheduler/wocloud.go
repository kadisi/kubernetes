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
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/types"

	"github.com/golang/glog"
	"github.com/kadisi/ipam/api/services/ipams"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

const (
	// AnnotationCMFloatingIP is in configmap annotation
	AnnotationCMFloatingIP = "wocloud.cn/floatingip"

	// ConfigMapFloatingIPKey is in configmap data key
	ConfigMapFloatingIPKey = "ipam"

	// TrueStr is true string
	TrueStr = "true"

	// stand this pod need Floatingip
	AnnotationPodNeedFloatingIP = "wocloud.cn/floatingip"

	// AnnotationPodFloatingIP is in pod annotation
	AnnotationPodFloatingIP = "wocloud.cn/floating-ip"
	// AnnotationPodSubnet is in pod annotation
	AnnotationPodSubnet = "wocloud.cn/floating-subnet"
	// AnnotationPodGateway is in pod annotation
	AnnotationPodGateway = "wocloud.cn/floating-gateway"
	// AnnotationPodConfigMap is in pod annotation
	AnnotationPodConfigMap = "wocloud.cn/floating-configmap"
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

	addAnnotationPatch := func(ip, subnet, gw, cm string) []byte {
		return []byte(fmt.Sprintf(`{"metadata":{"annotations":{"%s":"%s","%s":"%s","%s":"%s","%s":"%s"}}}`,
			AnnotationPodFloatingIP, ip,
			AnnotationPodSubnet, subnet,
			AnnotationPodGateway, gw,
			AnnotationPodConfigMap, cm))
	}

	_, err = c.Client.CoreV1().Pods(pod.GetNamespace()).Patch(pod.Name, types.MergePatchType, addAnnotationPatch(
		respon.Ipaminfo.Ip, respon.Ipaminfo.Subnet, respon.Ipaminfo.Gateway, respon.Ipaminfo.ConfigMap))
	if err != nil {
		glog.V(3).Infof("patch pod annotation for floatingip error %v", err)
		return err
	}

	return nil
}
