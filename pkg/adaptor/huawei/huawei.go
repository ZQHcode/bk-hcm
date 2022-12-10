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

package huawei

import (
	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/errf"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/region"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
)

// NewHuawei new huawei.
func NewHuawei() types.Factory {
	return new(huawei)
}

// NewHuaweiProxy new huawei proxy.
func NewHuaweiProxy() types.HuaWeiProxy {
	return new(huawei)
}

var (
	_ types.Factory     = new(huawei)
	_ types.HuaWeiProxy = new(huawei)
)

type huawei struct{}

func (h *huawei) iamClient(secret *types.BaseSecret, region *region.Region) (*iam.IamClient, error) {
	auth := basic.NewCredentialsBuilder().
		WithAk(secret.ID).
		WithSk(secret.Key).
		Build()

	client := iam.NewIamClient(
		iam.IamClientBuilder().
			WithRegion(region).
			WithCredential(auth).
			WithHttpConfig(config.DefaultHttpConfig()).
			Build())

	return client, nil
}

func validateSecret(s *types.Secret) error {
	if s == nil {
		return errf.New(errf.InvalidParameter, "secret is required")
	}

	if s.HuaWei == nil {
		return errf.New(errf.InvalidParameter, "huawei secret is required")
	}

	if err := s.HuaWei.Validate(); err != nil {
		return err
	}

	return nil
}