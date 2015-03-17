package main

import (
	"encoding/json"
	"os/exec"
)

func cephCommand(v interface{}, args ...string) error {
	args = append(args, "-f", "json")
	out, err := exec.Command(*cephBinary, args...).Output()
	if err != nil {
		return err
	}
	return json.Unmarshal(out, v)
}

// func execOsdUtilCommand(cluster string) ([]OsdUtilization, error) {
// 	df := exec.Command("df")
// 	out, err := df.Output()
// 	if err != nil {
// 		return nil, err
// 	}
// 	lines := strings.Split(string(out), "\n")
// 	osdUtilList := make([]OsdUtilization, 0, len(lines))
// 	for _, line := range lines {
// 		if strings.Contains(line, cluster) {
// 			fields := strings.Fields(line)
// 			fs := fields[0]
// 			mnt := fields[len(fields)-1]
// 			tmpUtl := fields[len(fields)-2]
// 			utl, err := strconv.Atoi(tmpUtl[:len(tmpUtl)-1])
// 			if err != nil {
// 				return nil, err
// 			}
// 			id, err := strconv.Atoi(strings.Split(mnt, "-")[1])
// 			if err != nil {
// 				return nil, err
// 			}
// 			osdUtil := OsdUtilization{
// 				OsdNum:      id,
// 				FileSystem:  fs,
// 				Mount:       mnt,
// 				Utilization: utl,
// 			}
// 			osdUtilList = append(osdUtilList, osdUtil)
// 		}
// 	}
// 	return osdUtilList, nil
// }
