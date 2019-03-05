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
	"github.com/golang/glog"
	"github.com/kadisi/ipam/api/services/ipams"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"time"

	corelisters "k8s.io/client-go/listers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
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
	AssiginFloattingIP(pod *v1.Pod)  error
}

type WoclouderClient struct {
	IpamServiceClient ipams.IpamServiceClient
	ConfigmapLister corelisters.ConfigMapLister
	Client  	clientset.Interface
}

func (c * WoclouderClient) AssiginFloattingIP(pod *v1.Pod)  error {

	need, ok := pod.GetAnnotations()[AnnotationPodNeedFloatingIP]
	if !ok || need != TrueStr {
		// Do nothing
		return nil
	}

	cachedcms, err := c.ConfigmapLister.List(labels.Everything())
	if err != nil {
		glog.Warningf("can not find any config in cached")
	}

	keyFunc := func(ns, name string) string {
		return fmt.Sprintf("%v_%v", ns, name)
	}

	cachemap := make(map[string]struct{})
	for _, cm := range cachedcms {
		cachemap[keyFunc(cm.GetNamespace(), cm.GetName())] = struct{}{}
		glog.Infof("get cached floatingip configmap ns %v name %v",
			cm.GetNamespace(), cm.GetName())
	}

	requestcms := make([]string, 0, 2)
	for _, v := range pod.Spec.Volumes {
		if v.ConfigMap != nil {
			if _, ok := cachemap[keyFunc(pod.GetNamespace(), v.ConfigMap.Name)]; ok {
				glog.Info("find pod %v in ns %v Volumes has floatingip configmap %v ",
					pod.GetName(), pod.GetNamespace(), v.ConfigMap.Name)
				requestcms = append(requestcms, v.ConfigMap.Name)
			}
		}
	}

	ctx, cancle := context.WithTimeout(context.Background(), time.Second * 5)
	defer cancle()

	respon, err := c.IpamServiceClient.AcquireIP(ctx, &ipams.AcquireIPRequest{
		Podname: pod.GetName(),
		Namespace: pod.GetNamespace(),
		ConfigMaps: requestcms,
	})
	if err != nil {
		glog.Warningf("rpc client acquireip error %v", err)
	}

	copypod := pod.DeepCopy()
	if copypod.Annotations == nil {
		copypod.Annotations = make(map[string]string)
	}
	copypod.Annotations[AnnotationPodFloatingIP] = respon.Ipaminfo.Ip
	copypod.Annotations[AnnotationPodSubnet] = respon.Ipaminfo.Subnet
	copypod.Annotations[AnnotationPodGateway] = respon.Ipaminfo.Gateway
	copypod.Annotations[AnnotationPodConfigMap] = respon.Ipaminfo.ConfigMap

	_, err = c.Client.CoreV1().Pods(pod.GetNamespace()).Update(copypod)
	if err != nil {
		glog.Warningf("update pod annotation for floatingip error %v", err)
		return err
	}

	return nil
}

