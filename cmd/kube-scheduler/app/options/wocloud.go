/* 
#  #############################################
#  Copyright (c) 2019-2039 All rights reserved.
#  #############################################
# 
#  Name:  wocloud.go
#  Date:  2019-02-21 10:24
#  Author:   zhangjie
#  Email:   iamzhangjie0619@163.com
#  Desc:  
# 
*/ 

package options

import (
	"github.com/spf13/pflag"
)
// WocloudOptions is wocloud options such as wocloud-ipam address
type WocloudOptions struct {
	IpamAddress string
}

func (o *WocloudOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.StringVar(&o.IpamAddress, "ipam-address", o.IpamAddress,
		"wocloud ipam address, such as 127.0.0.1:9000")
}



