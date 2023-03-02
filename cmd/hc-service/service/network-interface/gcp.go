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

// Package networkinterface defines network interface service.
package networkinterface

import (
	"fmt"

	adcore "hcm/pkg/adaptor/types/core"
	typesniproto "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/api/core"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	dataproto "hcm/pkg/api/data-service/cloud/network-interface"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// GcpNetworkInterfaceSync sync gcp cloud network interface.
func (n networkInterfaceAdaptor) GcpNetworkInterfaceSync(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.GcpNetworkInterfaceSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	err := req.Validate()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// sync network interface list from cloudapi.
	allCloudIDMap, err := n.SyncGcpNetworkInterfaceList(cts, req)
	if err != nil {
		logs.Errorf("%s-networkinterface request cloudapi response failed. accountID: %s, zone: %s, err: %v",
			enumor.Gcp, req.AccountID, req.Zone, err)
		return nil, err
	}

	// compare and delete network interface idmap from db.
	err = n.compareDeleteGcpNetworkInterfaceList(cts, req, allCloudIDMap)
	if err != nil {
		logs.Errorf("%s-networkinterface compare delete and dblist failed. accountID: %s, zone: %s, err: %v",
			enumor.Gcp, req.AccountID, req.Zone, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// SyncGcpNetworkInterfaceList sync network interface list from cloudapi.
func (n networkInterfaceAdaptor) SyncGcpNetworkInterfaceList(cts *rest.Contexts,
	req *hcservice.GcpNetworkInterfaceSyncReq) (map[string]bool, error) {

	allCloudIDMap := make(map[string]bool, 0)
	cli, err := n.adaptor.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return allCloudIDMap, err
	}

	opt := new(adcore.GcpListOption)
	opt.Zone = req.Zone
	tmpList, tmpErr := cli.ListNetworkInterface(cts.Kit, opt)
	if tmpErr != nil {
		logs.Errorf("%s-networkinterface batch get cloudapi failed. accountID: %s, zone: %s, err: %v",
			enumor.Gcp, req.AccountID, req.Zone, tmpErr)
		return allCloudIDMap, tmpErr
	}

	cloudIDs := make([]string, 0)
	for _, item := range tmpList.Details {
		tmpID := converter.PtrToVal(item.CloudID)
		cloudIDs = append(cloudIDs, tmpID)
		allCloudIDMap[tmpID] = true
	}

	// get network interface info from db.
	resourceDBMap, err := n.BatchGetNetworkInterfaceMapFromDB(cts, enumor.Gcp, cloudIDs)
	if err != nil {
		logs.Errorf("%s-networkinterface get routetabledblist failed. accountID: %s, zone: %s, err: %v",
			enumor.Gcp, req.AccountID, req.Zone, err)
		return allCloudIDMap, err
	}

	// compare and update network interface list.
	err = n.compareUpdateGcpNetworkInterfaceList(cts, req, tmpList, resourceDBMap)
	if err != nil {
		logs.Errorf("%s-networkinterface compare and update routetabledblist failed. accountID: %s, "+
			"zone: %s, err: %v", enumor.Gcp, req.AccountID, req.Zone, err)
		return allCloudIDMap, err
	}

	return allCloudIDMap, nil
}

// compareUpdateGcpNetworkInterfaceList compare and update network interface list.
func (n networkInterfaceAdaptor) compareUpdateGcpNetworkInterfaceList(cts *rest.Contexts,
	req *hcservice.GcpNetworkInterfaceSyncReq, list *typesniproto.GcpInterfaceListResult,
	resourceDBMap map[string]coreni.BaseNetworkInterface) error {
	createResources, updateResources, err := n.filterGcpNetworkInterfaceList(cts, req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &dataproto.NetworkInterfaceBatchUpdateReq[dataproto.GcpNICreateExt]{
			NetworkInterfaces: updateResources,
		}
		if err = n.dataCli.Gcp.NetworkInterface.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("%s-networkinterface batch compare db update failed. accountID: %s, zone: %s, err: %v",
				enumor.Gcp, req.AccountID, req.Zone, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &dataproto.NetworkInterfaceBatchCreateReq[dataproto.GcpNICreateExt]{
			NetworkInterfaces: createResources,
		}
		if _, err = n.dataCli.Gcp.NetworkInterface.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq); err != nil {
			logs.Errorf("%s-networkinterface batch compare db create failed. accountID: %s, zone: %s, err: %v",
				enumor.Gcp, req.AccountID, req.Zone, err)
			return err
		}
	}

	return nil
}

// filterGcpNetworkInterfaceList filter gcp network interface list
func (n networkInterfaceAdaptor) filterGcpNetworkInterfaceList(_ *rest.Contexts,
	req *hcservice.GcpNetworkInterfaceSyncReq, list *typesniproto.GcpInterfaceListResult,
	resourceDBMap map[string]coreni.BaseNetworkInterface) (
	createResources []dataproto.NetworkInterfaceReq[dataproto.GcpNICreateExt],
	updateResources []dataproto.NetworkInterfaceUpdateReq[dataproto.GcpNICreateExt], err error) {

	if list == nil || len(list.Details) == 0 {
		return createResources, updateResources,
			fmt.Errorf("cloudapi networkinterfacelist is empty, accountID: %s, zone: %s",
				req.AccountID, req.Zone)
	}

	for _, item := range list.Details {
		// need compare and update resource data
		tmpCloudID := converter.PtrToVal(item.CloudID)
		if resourceInfo, ok := resourceDBMap[tmpCloudID]; ok {
			if resourceInfo.CloudID == tmpCloudID && resourceInfo.CloudVpcID == converter.PtrToVal(item.CloudVpcID) &&
				resourceInfo.Name == converter.PtrToVal(item.Name) {
				continue
			}

			tmpRes := dataproto.NetworkInterfaceUpdateReq[dataproto.GcpNICreateExt]{
				ID:            resourceInfo.ID,
				AccountID:     req.AccountID,
				Name:          converter.PtrToVal(item.Name),
				Region:        converter.PtrToVal(item.Region),
				Zone:          converter.PtrToVal(item.Zone),
				CloudID:       converter.PtrToVal(item.CloudID),
				CloudVpcID:    converter.PtrToVal(item.CloudVpcID),
				CloudSubnetID: converter.PtrToVal(item.CloudSubnetID),
				PrivateIP:     converter.PtrToVal(item.PrivateIP),
				PublicIP:      converter.PtrToVal(item.PublicIP),
				InstanceID:    converter.PtrToVal(item.InstanceID),
			}
			if item.Extension != nil {
				tmpRes.Extension = &dataproto.GcpNICreateExt{
					CanIpForward: item.Extension.CanIpForward,
					Status:       item.Extension.Status,
					StackType:    item.Extension.StackType,
				}
				// 网卡私网IP信息列表
				var tmpAccConfigs []*dataproto.AccessConfig
				for _, accConfigItem := range item.Extension.AccessConfigs {
					tmpAccConfigs = append(tmpAccConfigs, &dataproto.AccessConfig{
						Name:        accConfigItem.Name,
						NatIP:       accConfigItem.NatIP,
						NetworkTier: accConfigItem.NetworkTier,
						Type:        accConfigItem.Type,
					})
				}
				tmpRes.Extension.AccessConfigs = tmpAccConfigs
			}

			updateResources = append(updateResources, tmpRes)
		} else {
			// need add resource data
			tmpRes := dataproto.NetworkInterfaceReq[dataproto.GcpNICreateExt]{
				AccountID:     req.AccountID,
				Vendor:        string(enumor.Gcp),
				Name:          converter.PtrToVal(item.Name),
				Region:        converter.PtrToVal(item.Region),
				Zone:          converter.PtrToVal(item.Zone),
				CloudID:       converter.PtrToVal(item.CloudID),
				CloudVpcID:    converter.PtrToVal(item.CloudVpcID),
				CloudSubnetID: converter.PtrToVal(item.CloudSubnetID),
				PrivateIP:     converter.PtrToVal(item.PrivateIP),
				PublicIP:      converter.PtrToVal(item.PublicIP),
				InstanceID:    converter.PtrToVal(item.InstanceID),
			}
			if item.Extension != nil {
				if item.Extension != nil {
					tmpRes.Extension = &dataproto.GcpNICreateExt{
						CanIpForward: item.Extension.CanIpForward,
						Status:       item.Extension.Status,
						StackType:    item.Extension.StackType,
					}
					// 网卡私网IP信息列表
					var tmpAccConfigs []*dataproto.AccessConfig
					for _, accConfigItem := range item.Extension.AccessConfigs {
						tmpAccConfigs = append(tmpAccConfigs, &dataproto.AccessConfig{
							Name:        accConfigItem.Name,
							NatIP:       accConfigItem.NatIP,
							NetworkTier: accConfigItem.NetworkTier,
							Type:        accConfigItem.Type,
						})
					}
					tmpRes.Extension.AccessConfigs = tmpAccConfigs
				}
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, nil
}

// compareDeleteGcpNetworkInterfaceList compare and delete network interface list from db.
func (n networkInterfaceAdaptor) compareDeleteGcpNetworkInterfaceList(cts *rest.Contexts,
	req *hcservice.GcpNetworkInterfaceSyncReq, allCloudIDMap map[string]bool) error {

	page := uint32(0)
	for {
		count := core.DefaultMaxPageLimit
		offset := page * uint32(count)
		expr := &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.Gcp,
				},
			},
		}
		dbQueryReq := &core.ListReq{
			Filter: expr,
			Page:   &core.BasePage{Count: false, Start: offset, Limit: count},
		}
		dbList, err := n.dataCli.Global.NetworkInterface.List(cts.Kit.Ctx, cts.Kit.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("%s-networkinterface batch get networkinterfacelist db error. offset: %d, limit: %d, "+
				"err: %v", enumor.Gcp, offset, count, err)
			return err
		}

		if len(dbList.Details) == 0 {
			return nil
		}

		deleteCloudIDMap := make(map[string]string, 0)
		for _, item := range dbList.Details {
			if _, ok := allCloudIDMap[item.CloudID]; !ok {
				deleteCloudIDMap[item.CloudID] = item.ID
			}
		}

		// batch query need delete network interface list
		deleteIDs := n.GetNeedDeleteGcpNetworkInterfaceList(cts, req, deleteCloudIDMap)
		if len(deleteIDs) > 0 {
			err = n.BatchDeleteNetworkInterfaceByIDs(cts, deleteIDs)
			if err != nil {
				logs.Errorf("%s-networkinterface batch compare db delete failed. deleteIDs: %v, err: %v",
					enumor.Gcp, deleteIDs, err)
				return err
			}
		}
		deleteIDs = nil

		if len(dbList.Details) < int(count) {
			break
		}
		page++
	}
	allCloudIDMap = nil

	return nil
}

// GetNeedDeleteGcpNetworkInterfaceList get need delete gcp network interface list
func (n networkInterfaceAdaptor) GetNeedDeleteGcpNetworkInterfaceList(_ *rest.Contexts,
	_ *hcservice.GcpNetworkInterfaceSyncReq, deleteCloudIDMap map[string]string) []string {

	deleteIDs := make([]string, 0, len(deleteCloudIDMap))
	if len(deleteCloudIDMap) == 0 {
		return deleteIDs
	}

	for _, tmpID := range deleteCloudIDMap {
		deleteIDs = append(deleteIDs, tmpID)
	}

	return deleteIDs
}