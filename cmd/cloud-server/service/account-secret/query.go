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

package accountsecret

import (
	"hcm/pkg/api/core"
	coreas "hcm/pkg/api/core/cloud/account-secret"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// getAccountSecretByID gets account secret by id.
func (s *service) getAccountSecretByID(kt *kit.Kit, id string) (*coreas.BaseAccountSecret, error) {
	result, err := s.client.DataService().Global.AccountSecret.ListAccountSecret(kt, &protocloud.AccountSecretListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("list account secret failed, id: %s, err: %v, rid: %s", id, err, kt.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "account secret %s not found", id)
	}

	return &result.Details[0], nil
}

func (s *service) getTCloudAccountSecretByID(kt *kit.Kit, id string) (
	*coreas.AccountSecret[coreas.TCloudAccountSecretExtension], error) {

	req := &protocloud.AccountSecretExtListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := s.client.DataService().TCloud.AccountSecret.ListAccountSecretWithExtension(kt, req)
	if err != nil {
		logs.Errorf("list account secret failed, id: %s, err: %v, rid: %s", id, err, kt.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "account secret %s not found", id)
	}

	return &result.Details[0], nil
}
