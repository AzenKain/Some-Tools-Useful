package main

import (
	"errors"
	assetMeta "language-change/asset-meta"
)

func Get32BitHashConst(s string) int32 {
	var hash1 int32 = 5381
	var hash2 int32 = hash1

	bytes := []byte(s)
	length := len(bytes)

	for i := 0; i < length; i += 2 {
		hash1 = ((hash1 << 5) + hash1) ^ int32(bytes[i])
		if i+1 < length {
			hash2 = ((hash2 << 5) + hash2) ^ int32(bytes[i+1])
		}
	}

	return int32(uint32(hash1) + uint32(hash2)*1566083941)
}

func GetAssetData(assets *assetMeta.DesignIndex, name string) (*assetMeta.DataEntry, *assetMeta.FileEntry, error) {
	dataEntry, fileEntry, err := assets.FindDataAndFileByTarget(Get32BitHashConst("BakedConfig/ExcelOutput/" + name + ".bytes"))
	if err == nil {
		return &dataEntry, &fileEntry, nil
	}
	dataEntry, fileEntry, err = assets.FindDataAndFileByTarget(Get32BitHashConst("BakedConfig/ExcelOutputGameCore/" + name + ".bytes"))
	if err == nil {
		return &dataEntry, &fileEntry, nil
	}
	dataEntry, fileEntry, err = assets.FindDataAndFileByTarget(Get32BitHashConst("BakedConfig/ExcelOutputAdventureGame/" + name + ".bytes"))
	if err == nil {
		return &dataEntry, &fileEntry, nil
	}
	return nil, nil, errors.New("not found")
}
