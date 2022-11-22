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

package errf

// NOTE: 错误码规则
// 20号段 + 5位错误码共7位
// 注意：
// - 特殊错误码, 2030403（未授权）, 内部保留

// common error code.
const (
	OK               int32 = 0
	PermissionDenied int32 = 2030403
)

// Note:
// this scope's error code ranges at [4000000, 4089999], and works for all the scenario
// except sidecar related scenario.
const (
	// Unknown is unknown error, it is always used when an
	// error is wrapped, but the error code is not parsed.
	Unknown int32 = 2000000
	// InvalidParameter means the request parameter  is invalid
	InvalidParameter int32 = 2000001
	// DecodeRequestFailed means decode the request body failed.
	DecodeRequestFailed int32 = 2000002
)