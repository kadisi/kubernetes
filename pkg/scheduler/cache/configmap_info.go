/* 
#  #############################################
#  Copyright (c) 2019-2039 All rights reserved.
#  #############################################
# 
#  Name:  configmap_info.go
#  Date:  2019-02-21 16:40
#  Author:   zhangjie
#  Email:   iamzhangjie0619@163.com
#  Desc:  
# 
*/ 

package cache

import (
	"errors"
	"k8s.io/api/core/v1"
)

type configMapInfo struct {
	configmap *v1.ConfigMap
}

// getConfigmapKey returns the string key of a configmap.
func getConfigmapKey(c *v1.ConfigMap) (string, error) {
	uid := string(c.UID)
	if len(uid) == 0 {
		return "", errors.New("Cannot get cache key for configmap with empty UID")
	}
	return uid, nil
}