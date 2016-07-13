package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

type DBSetting struct {
	User      string
	Password  string
	Stage     string
	Host      string
	Port      string
	LoadedSQL []string
	Numbers   []string
}

var (
	defaultSQLPath string = "./load.sql"
	concurrent     bool   = false
	parallelNum    int
	certFilePath   string = "config/ca-cert.pem"
	certFile       []byte
)

var ds *DBSetting

func main() {
	ds = &DBSetting{}

	loadSQLFile()
	selectStage()
	loadCa()
	selectHost()
	selectDBNumber()
	selectUser()
	selectDBPassword()
	selectConcurrent()
	selectLogger()
	if finalCheck() == true {
		execSQL()
	} else {
		fmt.Println("やめたよ…")
	}

}

func finalCheck() bool {
	fmt.Println("以下のSQLをながすよ？")
	fmt.Printf("環境:%s\n", ds.Stage)
	fmt.Printf("DBナンバー:%v\n", ds.Numbers)
	fmt.Println("SQL:")
	for _, sql := range ds.LoadedSQL {
		fmt.Println(sql)
	}
	fmt.Printf("host:%s\n", ds.Host)
	fmt.Printf("user:%s\n", ds.User)

	if concurrent == true {
		fmt.Printf("並列実行数：%d\n", parallelNum)
	}

	var check string
	fmt.Println("yes or no")
	fmt.Scanf("%s", &check)
	if check == "yes" {
		return true
	} else {
		return false
	}

}

func dbPort() string {
	if ds.Stage == "development" {
		return "3306"
	} else {
		return "3306"
	}
}

func dbBName(i string) (dbName string) {
	if ds.Stage != "production" {
		dbName = fmt.Sprintf("school_%s_%s", i, ds.Stage)
	} else {
		dbName = fmt.Sprintf("school_%s", i)
	}
	return
}

func selectUser() {
	user := os.Getenv("DB_USERNAME")
	if len(user) > 0 {
		ds.User = user
	} else {
		ds.User = "root"
		fmt.Printf("dbのuserを入れてね\nデフォルト %s:", ds.User)
		fmt.Scanf("%s", &ds.User)

	}
}

func selectStage() {
	stage := os.Getenv("STAGE")
	if len(stage) > 0 {
		ds.Stage = stage
	} else {
		ds.Stage = "development"
		fmt.Printf("環境を入れてね\nデフォルト %s:", ds.Stage)
		fmt.Scanf("%s", &ds.Stage)

	}
}

func loadCa() {
	if ds.Stage == "production" || ds.Stage == "staging" {
		var err error
		certFile, err = Asset(certFilePath)
		if err != nil {
			os.Exit(1)
		}
	}
}
func selectHost() {
	host := os.Getenv("DB_SERVERNAME")
	if len(host) > 0 {
		ds.Host = host
	} else {
		ds.Host = "localhost"
		fmt.Printf("環境を入れてね\nデフォルト %s:", ds.Host)
		fmt.Scanf("%s", &ds.Host)

	}
}

func selectDBPassword() {
	pass := os.Getenv("DB_PASSWORD")
	if len(pass) == 0 {
		fmt.Println("DBのパスワードを入れてね")
		fmt.Scanf("%s", &ds.Password)
	} else {
		ds.Password = pass
	}

}

func selectConcurrent() {
	fmt.Printf("並列実行する？(selectするときはおすすめできない) yes or no\n[デフォルト no]:")
	var conc string
	fmt.Scanf("%s", &conc)
	if conc == "yes" {
		concurrent = true
		parallelNum = runtime.NumCPU()
		if parallelNum <= 4 {
			parallelNum = 4
		}
	}
	fmt.Println(".\n.\n.\n.\n.\n.\n.\n.\n.\n.\n")
}

func selectDBNumber() {
	var tmpDBNumbers string
	fmt.Printf("実行するDBの数字をいれてね ex) 1..100 20,21,22\n")
	fmt.Scanf("%s", &tmpDBNumbers)
	switch {
	case strings.Contains(tmpDBNumbers, ".."):
		ary := strings.Split(tmpDBNumbers, "..")
		start, _ := strconv.Atoi(ary[0])
		end, _ := strconv.Atoi(ary[1])

		for now := start; now <= end; now++ {
			ds.Numbers = append(ds.Numbers, strconv.Itoa(now))
		}

	case strings.Contains(tmpDBNumbers, ","):
		ary := strings.Split(tmpDBNumbers, ",")
		ds.Numbers = make([]string, len(ary))
		for i, num := range ary {
			ds.Numbers[i] = num
		}
	}

	fmt.Println(ds.Numbers, "\n")
}

func otherSQL(db *sql.DB, sql string) {
	_, err := db.Exec(sql)
	if err != nil {
		log.Printf("error!!!:%s", err)
	}
}

func selectSQL(db *sql.DB, sql string) {
	res, err := db.Query(sql)
	if err != nil {
		log.Println(err)
	}
	log.Printf("%s\n start-----------\n", sql)
	flag := 0
	for res.Next() {
		columns, _ := res.Columns()
		if flag == 0 {
			header := ""
			for _, ss := range columns {
				header = header + ss + "\t"
			}
			log.Println(header)
			flag = 1
		}

		size := len(columns)
		valuePtrs := make([]interface{}, size)
		values := make([]interface{}, size)
		for i, _ := range columns {
			valuePtrs[i] = &values[i]
		}

		res.Scan(valuePtrs...)
		for i, _ := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				if val == nil {
					v = "null"
				} else {
					t, ok := val.(time.Time)
					if ok {
						v = t.Format("2006-01-02T15:04:05Z07:00")
					} else {
						v = "null"
					}
				}
			}
			values[i] = v
		}
		record := ""
		for _, val := range values {
			//fmt.Printf("%v\t", val)
			record = record + val.(string) + "\t"
		}
		//fmt.Printf("\n")
		log.Println(record)
	}
	log.Println("-----------end")
}

func execSQL() {
	if concurrent == false {
		for _, i := range ds.Numbers {
			execSQLOne(i)
		}
	} else {
		var wg sync.WaitGroup
		semaphore := make(chan int, parallelNum)
		for _, i := range ds.Numbers {
			wg.Add(1)
			go func(i2 string) {
				defer wg.Done()
				semaphore <- 1
				execSQLOne(i2)
				<-semaphore
			}(i)
		}
		wg.Wait()
	}
}

func execSQLOne(i string) {
	dbname := dbBName(i)
	dbuser := ds.User
	dbpassword := ds.Password
	dbhost := ds.Host
	dbport := dbPort()

	if len(dbpassword) > 0 {
		dbpassword = fmt.Sprintf(":%s", dbpassword)
	}
	//dbstring := fmt.Sprintf("root@tcp(localhost:3306)/%s?charset=utf8&parseTime=true", dbname)
	dbstring := fmt.Sprintf("%s%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true", dbuser, dbpassword, dbhost, dbport, dbname)

	if ds.Stage == "production" || ds.Stage == "staging" {
		certpool, _ := getCertPool(certFile)
		mysql.RegisterTLSConfig("custom", &tls.Config{
			RootCAs: certpool,
		})
	}

	db, err := sql.Open("mysql", dbstring)
	if err != nil {
		panic(fmt.Sprintf("Got error when connect database, the error is '%v'", err))
	}
	log.Printf("school %s start\n", i)
	for _, sql := range ds.LoadedSQL {
		if len(sql) <= 1 {
			continue
		}
		if strings.Contains(sql, "select") {
			selectSQL(db, sql)
		} else {
			otherSQL(db, sql)
		}

	}
	db.Close()
	fmt.Printf("school%s done\n\n", i)
	log.Printf("school %s end\n", i)
}

func loadSQLFile() {
	var tmpSqlPath string
	var sqlPath string
	fmt.Printf("実行するsqlのファイルを指定してね\n[デフォルト %s]:", defaultSQLPath)
	fmt.Scanf("%s", &tmpSqlPath)

	if len(tmpSqlPath) > 0 {
		sqlPath = tmpSqlPath
	} else {
		sqlPath = defaultSQLPath
	}
	fmt.Printf("実行ファイル: %s\n\n", sqlPath)
	readFile, err := os.Open(sqlPath)
	if err != nil {
		log.Println("sqlファイルが読み取れないよ", err)
		os.Exit(1)
	}
	defer readFile.Close()

	scanner := bufio.NewScanner(readFile)
	joinedSql := ""
	for scanner.Scan() {
		joinedSql = joinedSql + scanner.Text()
	}
	str := strings.Split(joinedSql, ";")

	for i, s := range str {
		if len(s) > 0 {
			str[i] = s + ";"
		}
	}
	ds.LoadedSQL = str
}

func selectLogger() {
	logfile, err := os.OpenFile("./d4c.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		os.Exit(1)
	}
	log.SetOutput(logfile)

}

func getCertPool(cert_file []byte) (*x509.CertPool, error) {
	certs := x509.NewCertPool()
	certs.AppendCertsFromPEM(cert_file)
	return certs, nil
}
