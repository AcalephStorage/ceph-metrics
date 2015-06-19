package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	influxdb "github.com/influxdb/influxdb/client"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type (
	CephStatus struct {
		Quorum []int `json:"quorum"`
		OSDMap struct {
			OSDMap struct {
				Epoch int `json:"epoch"`
			} `json:"osdmap"`
		} `json:"osdmap"`
		Health struct {
			OverallStatus string `json:"overall_status"`
		} `json:"health"`
	}

	MonStatus struct {
		Name  string `json:"name"`
		State string `json:"state"`
	}

	CephDF struct {
		Stats struct {
			TotalBytes          int64 `json:"total_bytes"`
			TotalUsedBytes      int64 `json:"total_used_bytes"`
			TotalAvailableBytes int64 `json:"total_avail_bytes"`
		} `json:"stats"`
		Pools []struct {
			Name  string `json:"name"`
			Id    int    `json:"id"`
			Stats struct {
				UsedKb    int64 `json:"kb_used"`
				UsedBytes int64 `json:"bytes_used"`
				Available int64 `json:"max_avail"`
				Objects   int64 `json::"objects"`
			} `json:"stats"`
		} `json:"pools"`
	}

	PoolStats struct {
		PoolName     string `json:"pool_name"`
		PoolId       int    `json:"pool_id"`
		ClientIoRate struct {
			WriteBytesPerSecond int `json:"write_bytes_sec"`
			OpsPerSec           int `json:"op_per_sec"`
		} `json:"client_io_rate"`
	}

	PoolQuota struct {
		PoolName   string `json:"pool_name"`
		PoolId     int    `json:"pool_id"`
		MaxObjects int64  `json:"quota_max_objects"`
		MaxBytes   int64  `json:"quota_max_bytes"`
	}

	OsdDump struct {
		Osds []struct {
			OsdNum int    `json:"osd"`
			Uuid   string `json:"uuid"`
			Up     int    `json:"up"`
			In     int    `json:"in"`
		} `json:"osds"`
	}

	OsdPerf struct {
		PerfInfo []struct {
			Id    int `json:"id"`
			Stats struct {
				CommitLatency int `json:"commit_latency_ms"`
				ApplyLatency  int `json:"apply_latency_ms"`
			} `json:"perf_stats"`
		} `json:"osd_perf_infos"`
	}

	PgDump struct {
		PgStatSum struct {
			StatSum map[string]int64 `json:"stat_sum"`
		} `json:"pg_stats_sum"`
		PoolStats []struct {
			PoolId  int                    `json:"poolid"`
			StatSum map[string]interface{} `json:"stat_sum"`
		} `json:"pool_stats"`
		PgStats []struct {
			PgId          string `json:"pgid"`
			Up            []int  `json:"up"`
			Acting        []int  `json:"acting"`
			UpPrimary     int    `json:"up_primary"`
			ActingPrimary int    `json:"acting_primary"`
		} `json:"pg_stats"`
		OsdStats []struct {
			Osd         int   `json:"osd"`
			TotalKb     int64 `json:"kb"`
			UsedKb      int64 `json:"kb_used"`
			AvailableKb int64 `json:"kb_avail"`
		} `json:"osd_stats"`
	}

	PoolOsdPgMap map[int]map[int]int
)

func processMetrics() {
	logrus.Infoln("Metrics gathering started.")
	for {

		var monStatus MonStatus
		var cephStatus CephStatus
		var cephDf CephDF
		var poolStatsList []PoolStats
		var osdDump OsdDump
		var osdPerf OsdPerf
		var pgDump PgDump

		time.Sleep(time.Duration(*interval) * time.Second)
		logrus.Infoln("Sending Metrics...")

		if err := cephCommand(&monStatus, "mon_status"); err != nil {
			logrus.Errorln("error: ", err)
			// error possibly means no mon or not a client?
			continue
		}

		isLeader := monStatus.State == "leader"

		if isLeader {
			if err := cephCommand(&cephStatus, "status"); err != nil {
				logrus.Errorln("error: ", err)
				continue
			}

			if err := cephCommand(&cephDf, "df"); err != nil {
				logrus.Errorln("error: ", err)
				continue
			}

			if err := cephCommand(&poolStatsList, "osd", "pool", "stats"); err != nil {
				logrus.Errorln("error: ", err)
				continue
			}

			if err := cephCommand(&osdDump, "osd", "dump"); err != nil {
				logrus.Errorln("error: ", err)
				continue
			}

			if err := cephCommand(&osdPerf, "osd", "perf"); err != nil {
				logrus.Errorln("error: ", err)
				continue
			}

			if err := cephCommand(&pgDump, "pg", "dump"); err != nil {
				logrus.Errorln("error: ", err)
				continue
			}

			updateCephHealth(cephStatus.Health.OverallStatus)
			sendClusterMetrics(&cephStatus, &cephDf, &pgDump, poolStatsList)
			sendOSDMetrics(&osdDump, &osdPerf, &pgDump)
		}
	}
}

func updateCephHealth(overallStatus string) {
	status := HealthWarn
	if overallStatus == "CEPH_OK" {
		status = HealthOk
	}
	healthUpdateChannel <- HealthUpdate{
		uri:     CephHealthUri,
		status:  status,
		message: overallStatus,
	}
}

func sendClusterMetrics(cephStatus *CephStatus, cephDf *CephDF, pgDump *PgDump, poolStatsList []PoolStats) {
	sendCephQuorum(cephStatus)
	sendUtilization(cephDf)
	sendClientIO(poolStatsList)
	sendOsdEpoch(cephStatus)
	sendPgMetrics(pgDump, poolStatsList)
	sendPgDistribution(pgDump)
	sendObjectMetrics(pgDump)
	sendPoolMetrics(pgDump, poolStatsList)
	sendPoolUtilization(cephDf)
	sendPoolQuotas(cephDf)
	sendPoolIO(poolStatsList)
}

func sendCephQuorum(cephStatus *CephStatus) {
	monInQuorum := len(cephStatus.Quorum)

	publishMetrics(
		fmt.Sprintf("%s_quorum", *cluster),
		fmt.Sprintf("%d", monInQuorum),
	)

	status := HealthOk
	if monInQuorum == 1 {
		status = HealthWarn
	}
	healthUpdateChannel <- HealthUpdate{
		uri:     CephMonQuorumUri,
		status:  status,
		message: fmt.Sprintf("mon in quorum: %d", monInQuorum),
	}
}

func sendUtilization(cephDf *CephDF) {
	total := float64(cephDf.Stats.TotalBytes)
	used := float64(cephDf.Stats.TotalUsedBytes)
	utilized := 0.0
	if used != 0 {
		utilized = used / total * 100
	}

	publishMetrics(
		fmt.Sprintf("%s_utilization", *cluster),
		fmt.Sprintf("%0.0f", utilized),
	)
}

func sendClientIO(poolStatsList []PoolStats) {
	sumOps := 0
	sumWrs := 0
	for _, stat := range poolStatsList {
		sumOps += stat.ClientIoRate.OpsPerSec
		sumWrs += stat.ClientIoRate.WriteBytesPerSecond / 1024
	}

	publishMetrics(
		fmt.Sprintf("%s_client_io_kbs", *cluster),
		fmt.Sprintf("%d", sumWrs),
	)

	publishMetrics(
		fmt.Sprintf("%s_client_io_ops", *cluster),
		fmt.Sprintf("%d", sumOps),
	)
}

func sendOsdEpoch(cephStatus *CephStatus) {
	epoch := cephStatus.OSDMap.OSDMap.Epoch

	publishMetrics(
		fmt.Sprintf("%s_osd_epoch", *cluster),
		fmt.Sprintf("%d", epoch),
	)
}

func sendPgMetrics(pgDump *PgDump, poolStatsList []PoolStats) {

	pgCount := len(pgDump.PgStats)

	publishMetrics(
		fmt.Sprintf("%s_pg_count", *cluster),
		fmt.Sprintf("%d", pgCount),
	)

	poolPgs := make(map[string]int, len(pgDump.PoolStats))
	for _, stat := range pgDump.PgStats {
		poolId := strings.Split(stat.PgId, ".")[0]
		poolPgs[poolId] = poolPgs[poolId] + 1
	}

	for pool, pgs := range poolPgs {
		publishMetrics(
			fmt.Sprintf("%s_pg_pool_%s_count", *cluster, pool),
			fmt.Sprintf("%d", pgs),
		)
	}

	// do we need more metrics?

}

func sendPgDistribution(pgDump *PgDump) {

	numOfPool := len(pgDump.PoolStats)
	numOfOsds := len(pgDump.OsdStats)

	poolOsdPgMap := make(PoolOsdPgMap, numOfPool)
	totalOsdPgs := make(map[int]int, numOfOsds)

	for _, pgStat := range pgDump.PgStats {
		poolId, _ := strconv.Atoi(strings.Split(pgStat.PgId, ".")[0])

		osdPgMap := poolOsdPgMap[poolId]
		if osdPgMap == nil {
			osdPgMap = make(map[int]int, numOfOsds)
			poolOsdPgMap[poolId] = osdPgMap
		}

		for _, osd := range pgStat.Up {
			osdPgMap[osd] = osdPgMap[osd] + 1
			totalOsdPgs[osd] = totalOsdPgs[osd] + 1
		}
	}

	for poolId, osdPgMap := range poolOsdPgMap {
		poolPg := 0
		for osdId, pgs := range osdPgMap {
			poolPg += pgs
			publishMetrics(
				fmt.Sprintf("%s_pg_distribution_pool_%d_osd_%d", *cluster, poolId, osdId),
				fmt.Sprintf("%d", pgs),
			)
		}
		publishMetrics(
			fmt.Sprintf("%s_pg_distribution_pool_%d", *cluster, poolId),
			fmt.Sprintf("%d", poolPg),
		)
	}

	for osd, pg := range totalOsdPgs {
		publishMetrics(
			fmt.Sprintf("%s_pg_distribution_osd_%d", *cluster, osd),
			fmt.Sprintf("%d", pg),
		)
	}

}

func sendObjectMetrics(pgDump *PgDump) {
	for k, v := range pgDump.PgStatSum.StatSum {
		publishMetrics(
			fmt.Sprintf("%s_pg_stats_%s", *cluster, k),
			fmt.Sprintf("%d", v),
		)
	}
}

func sendPoolMetrics(pgDump *PgDump, poolStatsList []PoolStats) {
	for _, pgPoolStat := range pgDump.PoolStats {
		for k, v := range pgPoolStat.StatSum {
			publishMetrics(
				fmt.Sprintf("%s_pool_%d_%s", *cluster, pgPoolStat.PoolId, k),
				fmt.Sprint(v),
			)
		}
	}
}

func sendPoolUtilization(cephDf *CephDF) {
	for _, pool := range cephDf.Pools {
		used := float64(pool.Stats.UsedBytes)
		total := float64(pool.Stats.Available)
		utilized := 0.0
		if used != 0 {
			utilized = (used / total) * 100.0
		}
		publishMetrics(
			fmt.Sprintf("%s_utilization", *cluster),
			fmt.Sprintf("%0.0f", utilized),
		)
	}
}

func sendPoolQuotas(cephDf *CephDF) {
	for _, pool := range cephDf.Pools {
		poolId := pool.Id
		numObjects := pool.Stats.Objects
		numBytes := pool.Stats.UsedBytes

		var quota PoolQuota

		err := cephCommand(&quota, "osd", "pool", "get-quota", pool.Name)
		if err != nil {
			logrus.Errorln("error: ", err)
			continue
		}
		maxObjects := quota.MaxObjects
		maxBytes := quota.MaxBytes

		publishMetrics(
			fmt.Sprintf("%s_pool_%d_objects", *cluster, poolId),
			fmt.Sprintf("%d", numObjects),
		)

		publishMetrics(
			fmt.Sprintf("%s_pool_%d_bytes", *cluster, poolId),
			fmt.Sprintf("%d", numBytes),
		)

		publishMetrics(
			fmt.Sprintf("%s_pool_%d_maxobjects", *cluster, poolId),
			fmt.Sprintf("%d", maxObjects),
		)

		publishMetrics(
			fmt.Sprintf("%s_pool_%d_maxbytes", *cluster, poolId),
			fmt.Sprintf("%d", maxBytes),
		)
	}
}

func sendPoolIO(poolStatsList []PoolStats) {
	for _, stat := range poolStatsList {
		poolId := stat.PoolId
		kbs := stat.ClientIoRate.WriteBytesPerSecond / 1024
		ops := stat.ClientIoRate.OpsPerSec

		publishMetrics(
			fmt.Sprintf("%s_pool_%d_io_kbs", *cluster, poolId),
			fmt.Sprintf("%d", kbs),
		)

		publishMetrics(
			fmt.Sprintf("%s_pool_io_%d_io_ops", *cluster, poolId),
			fmt.Sprintf("%d", ops),
		)

	}
}

func sendOSDMetrics(osdDump *OsdDump, osdPerf *OsdPerf, pgDump *PgDump) {
	sendOsdStatus(osdDump)
	sendOsdUtilization(pgDump)
	// sendOsdBalance() // how to get metrics for this?
	// sendOsdScrubbing() // and this one too?
	sendOsdLatency(osdPerf)
}

func sendOsdStatus(osdDump *OsdDump) {
	totalUp := 0
	totalIn := 0
	for _, osd := range osdDump.Osds {
		osdNum := osd.OsdNum
		up := osd.Up
		in := osd.In

		totalUp += up
		totalIn += in

		publishMetrics(
			fmt.Sprintf("%s_osd_%d_up", *cluster, osdNum),
			fmt.Sprintf("%d", up),
		)

		publishMetrics(
			fmt.Sprintf("%s_osd_%d_in", *cluster, osdNum),
			fmt.Sprintf("%d", in),
		)
	}

	publishMetrics(
		fmt.Sprintf("%s_osd_up_total", *cluster),
		fmt.Sprintf("%d", totalUp),
	)

	publishMetrics(
		fmt.Sprintf("%s_osd_in_total", *cluster),
		fmt.Sprintf("%d", totalIn),
	)
}

func sendOsdUtilization(pgDump *PgDump) {
	for _, osdStat := range pgDump.OsdStats {
		osdNum := osdStat.Osd
		used := float64(osdStat.UsedKb)
		total := float64(osdStat.TotalKb)
		utilized := (used / total) * 100.0
		publishMetrics(
			fmt.Sprintf("%s_osd_%d_utilization", *cluster, osdNum),
			fmt.Sprintf("%0.0f", utilized),
		)
	}
}

func sendOsdLatency(osdPerf *OsdPerf) {
	for _, perf := range osdPerf.PerfInfo {
		osdNum := perf.Id
		commit := perf.Stats.CommitLatency
		apply := perf.Stats.ApplyLatency

		publishMetrics(
			fmt.Sprintf("%s_osd_%d_latency_commit", *cluster, osdNum),
			fmt.Sprintf("%d", commit),
		)

		publishMetrics(
			fmt.Sprintf("%s_osd_%d_latency_apply", *cluster, osdNum),
			fmt.Sprintf("%d", apply),
		)
	}
}

func publishMetrics(name, value string) {

	var (
		pts = make([]influxdb.Point, 1)
	)

	pts[0] = influxdb.Point{
		Measurement: name,
		Fields: map[string]interface{}{
			"value": value,
		},
		Time:      time.Now(),
		Precision: "s",
	}

	bps := influxdb.BatchPoints{
		Points:          pts,
		Database:        "metricdb",
		RetentionPolicy: "default",
	}

	url, err := url.Parse(fmt.Sprintf("http://%s:%d", "localhost", 8086))
	conf := influxdb.Config{
		URL:      *url,
		Username: "root",
		Password: "root",
	}

	con, err := influxdb.NewClient(conf)

	if err != nil {
		logrus.Errorln("Unable to send establish connection to Influxdb", ":8086", err)
	}

	con.Write(bps)

	if err != nil {
		logrus.Errorln("Can't insert data to Influxdb")

	}

}
