//
// FilePath    : go-utils\sonyflake.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 雪花 ID 生成器
//

package utils

import (
	"fmt"
	"net"
)

// getUniqueMachineIDFromIP 通过IP获取机器ID
//
// 导入 net 包以获取网络接口和IP地址信息。
// 使用 net.Interfaces() 获取计算机上所有网络接口的列表。
// 通过检查每个接口的 Addrs() 方法来找到一个有效的IPv4地址。注意：此示例仅针对 IPv4 地址。
// 从 IP 地址生成一个唯一的 uint16 值。您可以通过将IP地址的最后两个字节组合成一个 uint16 值来实现这一点。
func getUniqueMachineIDFromIP() (uint16, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return 0, err
	}

	for _, intf := range interfaces {
		addrs, err := intf.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}

			if ipv4 := ip.To4(); ipv4 != nil {
				machineID := (uint16(ipv4[2]) << 8) + uint16(ipv4[3])
				return machineID, nil
			}
		}
	}

	return 0, fmt.Errorf("no suitable IPv4 address found")
}

// getUniqueMachineIDFromMac 通过MAC获取机器ID
// 导入 net 包以获取网络接口和 MAC 地址信息。
// 使用 net.Interfaces() 获取计算机上所有网络接口的列表。
// 通过检查每个接口的 HardwareAddr 字段来找到一个有效的 MAC 地址。注意：此示例仅针对 MAC 地址。
func getUniqueMachineIDFromMac() (uint16, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return 0, err
	}

	for _, intf := range interfaces {
		if len(intf.HardwareAddr) >= 4 {
			machineID := (uint16(intf.HardwareAddr[2]) << 8) + uint16(intf.HardwareAddr[3])
			return machineID, nil
		}
	}

	return 0, fmt.Errorf("no suitable network interface found")
}

// GetUniqueMachineID 通过 IP 或 MAC 地址获取机器ID
func GetUniqueMachineID(ipOrMac string) (uint16, error) {
	switch ipOrMac {
	case "ip":
		return getUniqueMachineIDFromIP()
	case "mac":
		return getUniqueMachineIDFromMac()
	default:
		return 0, fmt.Errorf("no suitable network interface found")
	}
}
