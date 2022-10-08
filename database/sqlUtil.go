package database

import "database/sql"

type MachineObj struct {
	Ip   string `json:"Ip"`
	Name string `json:"Name"`
}

type dataObj struct {
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

func InsertMachine(db *sql.DB, d MachineObj) error {
	sql := `INSERT OR REPLACE into machine (Ip,Name) values(?,?)`
	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(d.Ip, d.Name)
	return err
}

func InsertData(db *sql.DB, d dataObj) error {
	sql := `insert into gpu (Ip, GpuId, MemTotal, MemUsed, MemFree, GpuTemp, GpuFanSpeed, GpuPowerStat, GpuUtilRate, GpuMemRate, Time) values(?,?,?,?,?,?,?,?,?,?,?)`
	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(d.Ip, d.GpuId, d.MemTotal, d.MemUsed, d.MemFree, d.GpuTemp, d.GpuFanSpeed, d.GpuPowerStat, d.GpuUtilRate, d.GpuMemRate, d.Time)
	return err
}

func QueryData(db *sql.DB, ip string) (l []dataObj, e error) {
	sql := `select * from users where Ip=?`
	rows, err := db.Query(sql, ip)
	if err != nil {
		return nil, err
	}
	var result = make([]dataObj, 0)
	for rows.Next() {
		var Ip string
		var GpuId, GpuTemp, GpuFanSpeed, GpuPowerStat, GpuUtilRate, GpuMemRate, Time int
		var MemTotal, MemUsed, MemFree float32
		rows.Scan(&Ip, &GpuId, &MemTotal, &MemUsed, &MemFree, &GpuTemp, &GpuFanSpeed, &GpuPowerStat, &GpuUtilRate, &GpuMemRate, &Time)
		result = append(result, dataObj{Ip, GpuId, MemTotal, MemUsed, MemFree, GpuTemp, GpuFanSpeed, GpuPowerStat, GpuUtilRate, GpuMemRate, Time})
	}
	return result, nil
}

func QueryMachine(db *sql.DB, ip string) (name string, e error) {
	sql := `select * from users where Ip=?`
	err := db.QueryRow(sql, ip).Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}
