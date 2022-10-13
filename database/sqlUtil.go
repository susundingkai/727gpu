package database

import (
	"database/sql"
)

type MachineObj struct {
	Ip    string `json:"Ip"`
	Name  string `json:"Name"`
	Cnt   int    `json:"Cnt"`
	Model string `json:"Model"`
}

type DataObj struct {
	Ip           string  `json:"Ip"`
	GpuId        int     `json:"GpuId"`
	MemTotal     float32 `json:"MemTotal"`
	MemUsed      float32 `json:"MemUsed"`
	MemFree      float32 `json:"MemFree"`
	GpuTemp      int     `json:"GpuTemp"`
	GpuFanSpeed  int     `json:"GpuFanSpeed"`
	GpuPowerStat int     `json:"GpuPowerStat"`
	GpuUtilRate  int     `json:"GpuUtilRate"`
	GpuMemRate   int     `json:"GpuMemRate"`
	Time         int     `json:"Time"`
}
type DataSlice []DataObj

func (l DataSlice) Len() int           { return len(l) }
func (l DataSlice) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l DataSlice) Less(i, j int) bool { return l[i].Time < l[j].Time }
func InsertMachine(db *sql.DB, d MachineObj) error {
	sql := `INSERT OR REPLACE into machine (Ip,Name,Cnt,Model) values(?,?,?,?)`
	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(d.Ip, d.Name, d.Cnt, d.Model)
	return err
}

func InsertData(db *sql.DB, d DataObj) error {
	sql := `insert into gpu (Ip, GpuId, MemTotal, MemUsed, MemFree, GpuTemp, GpuFanSpeed, GpuPowerStat, GpuUtilRate, GpuMemRate, Time) values(?,?,?,?,?,?,?,?,?,?,?)`
	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(d.Ip, d.GpuId, d.MemTotal, d.MemUsed, d.MemFree, d.GpuTemp, d.GpuFanSpeed, d.GpuPowerStat, d.GpuUtilRate, d.GpuMemRate, d.Time)
	return err
}

func QueryData(db *sql.DB, ip string) (l DataSlice, e error) {
	sql := `select * from gpu where Ip=?`
	rows, err := db.Query(sql, ip)
	if err != nil {
		return nil, err
	}
	var result = make([]DataObj, 0)
	for rows.Next() {
		var Ip string
		var Id, GpuId, GpuTemp, GpuFanSpeed, GpuPowerStat, GpuUtilRate, GpuMemRate, Time int
		var MemTotal, MemUsed, MemFree float32
		rows.Scan(&Id, &Ip, &GpuId, &MemTotal, &MemUsed, &MemFree, &GpuTemp, &GpuFanSpeed, &GpuPowerStat, &GpuUtilRate, &GpuMemRate, &Time)
		result = append(result, DataObj{Ip, GpuId, MemTotal, MemUsed, MemFree, GpuTemp, GpuFanSpeed, GpuPowerStat, GpuUtilRate, GpuMemRate, Time})
	}
	return result, nil
}
func QueryNewData(db *sql.DB, lastTime int, ip string) (l DataSlice, e error) {
	var _sql string
	var rows *sql.Rows
	var err error
	_sql = "select * from (select * from gpu order by `Time` desc ) group by `GpuId`"
	rows, err = db.Query(_sql, ip)
	if err != nil {
		return nil, err
	}
	var result = make([]DataObj, 0)
	for rows.Next() {
		var Ip string
		var Id, GpuId, GpuTemp, GpuFanSpeed, GpuPowerStat, GpuUtilRate, GpuMemRate, Time int
		var MemTotal, MemUsed, MemFree float32
		rows.Scan(&Id, &Ip, &GpuId, &MemTotal, &MemUsed, &MemFree, &GpuTemp, &GpuFanSpeed, &GpuPowerStat, &GpuUtilRate, &GpuMemRate, &Time)
		result = append(result, DataObj{Ip, GpuId, MemTotal, MemUsed, MemFree, GpuTemp, GpuFanSpeed, GpuPowerStat, GpuUtilRate, GpuMemRate, Time})
	}
	return result, nil
}
func QueryAllMachine(db *sql.DB) (l []MachineObj, e error) {
	sql := `select * from machine`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	var result = make([]MachineObj, 0)
	for rows.Next() {
		var Ip, Name, Model string
		var Cnt int
		rows.Scan(&Ip, &Name, &Cnt, &Model)
		result = append(result, MachineObj{Ip, Name, Cnt, Model})
	}
	return result, nil
}

func QueryMachine(db *sql.DB, ip string) (name string, e error) {
	sql := `select * from machine where Ip=?`
	err := db.QueryRow(sql, ip).Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}
