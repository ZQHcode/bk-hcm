/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package handlers

import (
	"fmt"

	hcprotoinstancetype "hcm/pkg/api/hc-service/instance-type"
)

// GetTCloudInstanceType 查询机型
func (a *BaseApplicationHandler) GetTCloudInstanceType(
	accountID, region, zone, instanceType string,
) (*hcprotoinstancetype.TCloudInstanceTypeResp, error) {
	resp, err := a.Client.HCService().TCloud.InstanceType.List(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&hcprotoinstancetype.TCloudInstanceTypeListReq{AccountID: accountID, Region: region, Zone: zone},
	)
	if err != nil {
		return nil, err
	}

	// 遍历查找
	for _, i := range resp {
		if i.InstanceType == instanceType {
			return i, nil
		}
	}

	return nil, fmt.Errorf(
		"not found tcloud instanceType by accountID(%s), region(%s), zone (%s)",
		accountID, region, zone,
	)
}
