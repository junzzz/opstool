package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"encoding/csv"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var (
	masterPath = flag.String("master", "", "動画マスターファイル")

	dbUser string = "root"
	dbPass string = ""
	dbHost string = "localhost"
	dbPort string = "3306"
	dbName string = "classi_development"

	DB *sql.DB

	subjectCategories []subjectCategory
	subjects          []subject
	bigUnits          []bigUnit
	midUnits          []midUnit
	smallUnits        []smallUnit
	difficults        []difficult
)

type subjectCategory struct {
	Id           int
	Name         string
	SchoolAgesId int
}

type subject struct {
	Id                int
	SubjectCategoryId int
	Name              string
}

type bigUnit struct {
	Id        int
	SubjectId int
	Name      string
}

type midUnit struct {
	Id                int
	SubjectId         int
	Name              string
	BigTeachingUnitId int
}

type smallUnit struct {
	Id                   int
	SubjectId            int
	Name                 string
	BigTeachingUnitId    int
	MiddleTeachingUnitId int
}

type difficult struct {
	Id           int
	SchoolAgesId int
	ShortName    string
}

func init() {
	DB = initDB()
}

func main() {
	flag.Parse()
	var masterList [][]string
	if len(*masterPath) != 0 {
		masterList = parseMaster()
	} else {
		fmt.Println("動画マスターを指定してね")
		os.Exit(1)
	}
	if len(masterList) == 0 {
		fmt.Println("データがないよ")
		os.Exit(1)
	}
	loadDBData()

	newCSVData := make([][]string, 0, len(masterList) + 1)
	newCSVData = append(newCSVData, []string{"id","xxid","学齢","","教科","","科目","","大単元","","中単元","","小単元","","難易度","","作成者","","URL","動画タイトル","説明文1","説明文2","",""})
	for _, val := range masterList {
		newData := convData(val)
		newCSVData = append(newCSVData, newData)
	}

	defer DB.Close()
	writeCSV(newCSVData)
}

func writeCSV(csvData [][]string) {
	outFile, err := os.OpenFile("./convert.csv", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println("error", err)
	}
	defer outFile.Close()

	err = outFile.Truncate(0)
	if err != nil {
		fmt.Println("error", err)
	}
	writer := csv.NewWriter(transform.NewWriter(outFile, japanese.ShiftJIS.NewEncoder()))
	for _, line := range csvData {
		writer.Write(line)
	}
	writer.Flush()
}

func convData(data []string) []string {
	newData := make([]string, 24)
	// id
	newData[0] = data[0]

	// ビデオグid
	newData[1] = data[1]

	// 学齢
	newData[2] = data[2]
	if data[2] == "高校" {
		newData[3] = "2"
	} else {
		newData[3] = "1"
	}

	// 教科
	newData[4] = data[3]
	if len(data[3]) > 0 {
		for _, subcat := range subjectCategories {
			if subcat.Name == newData[4] && newData[3] == strconv.Itoa(subcat.SchoolAgesId) {
				newData[5] = strconv.Itoa(subcat.Id)
				break
			}
		}
		if len(data[3]) > 0 && len(newData[5]) == 0 {
			newData[5] = "エラー！！！！！！教科"
		}
	}

	// 科目
	newData[6] = data[4]
	if len(data[4]) > 0 {
		subject_category_id, _ := strconv.Atoi(newData[5])
		for _, sub := range subjects {
			if sub.Name == newData[6] && sub.SubjectCategoryId == subject_category_id {
				newData[7] = strconv.Itoa(sub.Id)
				break
			}
		}
		if len(data[4]) > 0 && len(newData[7]) == 0 {
			newData[7] = "エラー！！！！！！科目"
		}
	}

	// 大単元
	newData[8] = data[5]
	if len(data[5]) > 0 {
		subject_id, _ := strconv.Atoi(newData[7])
		for _, bunit := range bigUnits {
			if bunit.Name == newData[8] && bunit.SubjectId == subject_id {
				newData[9] = strconv.Itoa(bunit.Id)
				break
			}
		}
		if len(data[5]) > 0 && len(newData[9]) == 0 {
			newData[9] = "エラー！！！！！！大単元"
		}
	}

	// 中単元
	newData[10] = data[6]
	if len(data[6]) > 0 {
		subject_id, _ := strconv.Atoi(newData[7])
		big_id, _ := strconv.Atoi(newData[9])
		for _, munit := range midUnits {
			if munit.Name == newData[10] && munit.SubjectId == subject_id && munit.BigTeachingUnitId == big_id {
				newData[11] = strconv.Itoa(munit.Id)
				break
			}
		}
		if len(data[6]) > 0 && len(newData[11]) == 0 {
			newData[11] = "エラー！！！！！！中単元"
		}
	}

	// 小単元
	newData[12] = data[7]
	if len(data[7]) > 0 {
		subject_id, _ := strconv.Atoi(newData[7])
		big_id, _ := strconv.Atoi(newData[9])
		mid_id, _ := strconv.Atoi(newData[11])
		for _, sunit := range smallUnits {
			if sunit.Name == newData[12] && sunit.SubjectId == subject_id && sunit.BigTeachingUnitId == big_id && sunit.MiddleTeachingUnitId == mid_id {
				newData[13] = strconv.Itoa(sunit.Id)
				break
			}
		}
		if len(data[7]) > 0 && len(newData[13]) == 0 {
			newData[13] = "エラー！！！！！！小単元"
		}
	}

	// 難易度
	newData[14] = data[8]
	if len(data[8]) > 0 {
		school_age_id, _ := strconv.Atoi(newData[3])
		for _, diff := range difficults {
			if diff.ShortName == newData[14] && diff.SchoolAgesId == school_age_id {
				newData[15] = strconv.Itoa(diff.Id)
				break
			}
		}
		if len(data[8]) > 0 && len(newData[15]) == 0 {
			newData[15] = "エラー！！！！！！難易度"
		}
	}
	// 事業者
	newData[16] = data[9]
	switch newData[16] {
	// アレなので削る
	case "xxx":
		newData[17] = "jst"
	}
	if len(data[9]) > 0 && len(newData[17]) == 0{
		newData[17] = "エラー！！！！！！事業者"
	}
	// タイトル
	newData[18] = data[10]
	// URL
	newData[19] = data[11]
	// 説明文1
	newData[20] = data[12]
	// 説明文2
	newData[21] = data[13]

	return newData
}

func initDB() *sql.DB {
	if len(dbPass) > 0 {
		dbPass = fmt.Sprintf(":%s", dbPass)
	}
	//dbstring := fmt.Sprintf("root@tcp(localhost:3306)/%s?charset=utf8&parseTime=true", dbname)
	dbstr := fmt.Sprintf("%s%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", dbstr)
	if err != nil {
		panic(fmt.Sprintf("Got error when connect database, the error is '%v'", err))
	}
	return db
}

func loadDBData() {
	loadSubjectCategories()
	loadSubjects()
	loadBigUnits()
	loadMidUnits()
	loadSmallUnits()
	loadDifficults()
}

func loadSubjectCategories() {
	subjectCategories = make([]subjectCategory, 0, 13)
	res, err := DB.Query("select id, name, school_ages_id from attribute_subject_categories;")
	if err != nil {
		fmt.Println("DB error", err)
		os.Exit(1)
	}
	defer res.Close()
	for res.Next() {
		subcat := subjectCategory{}
		err = res.Scan(&subcat.Id, &subcat.Name, &subcat.SchoolAgesId)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		subjectCategories = append(subjectCategories, subcat)

	}
}

func loadSubjects() {
	subjects = make([]subject, 0, 33)
	res, err := DB.Query("select id, subject_category_id, name from attribute_subjects;")
	if err != nil {
		fmt.Println("DB error", err)
		os.Exit(1)
	}
	defer res.Close()
	for res.Next() {
		sub := subject{}
		err = res.Scan(&sub.Id, &sub.SubjectCategoryId, &sub.Name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		subjects = append(subjects, sub)

	}
}

func loadBigUnits() {
	bigUnits = make([]bigUnit, 0, 239)
	res, err := DB.Query("select id, subject_id, name from attribute_big_teaching_units;")
	if err != nil {
		fmt.Println("DB error", err)
		os.Exit(1)
	}
	defer res.Close()
	for res.Next() {
		bigunit := bigUnit{}
		err = res.Scan(&bigunit.Id, &bigunit.SubjectId, &bigunit.Name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		bigUnits = append(bigUnits, bigunit)
	}
}

func loadMidUnits() {
	midUnits = make([]midUnit, 0, 684)
	res, err := DB.Query("select id, subject_id, name, big_teaching_units_id from attribute_middle_teaching_units;")
	if err != nil {
		fmt.Println("DB error", err)
		os.Exit(1)
	}
	defer res.Close()
	for res.Next() {
		midunit := midUnit{}
		err = res.Scan(&midunit.Id, &midunit.SubjectId, &midunit.Name, &midunit.BigTeachingUnitId)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		midUnits = append(midUnits, midunit)
	}
}

func loadSmallUnits() {
	smallUnits = make([]smallUnit, 0, 1226)
	res, err := DB.Query("select id, subject_id, name, big_teaching_units_id, middle_teaching_units_id from attribute_small_teaching_units;")
	if err != nil {
		fmt.Println("DB error", err)
		os.Exit(1)
	}
	defer res.Close()
	for res.Next() {
		smallunit := smallUnit{}
		err = res.Scan(&smallunit.Id, &smallunit.SubjectId, &smallunit.Name, &smallunit.BigTeachingUnitId, &smallunit.MiddleTeachingUnitId)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		smallUnits = append(smallUnits, smallunit)
	}
}

func loadDifficults() {
	difficults = make([]difficult, 0, 9)
	res, err := DB.Query("select id, school_ages_id, short_name  from attribute_difficulties;")
	if err != nil {
		fmt.Println("DB error", err)
		os.Exit(1)
	}
	defer res.Close()
	for res.Next() {
		diff := difficult{}
		err = res.Scan(&diff.Id, &diff.SchoolAgesId, &diff.ShortName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		difficults = append(difficults, diff)
	}
}

func parseMaster() [][]string {
	file, err := os.Open(*masterPath)
	if err != nil {
		fmt.Println("マスターファイルを取得できない：", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	header := 0
	masterList := make([][]string, 0, 1000)
	for scanner.Scan() {
		if header == 0 {
			header = header + 1
			continue
		}
		line := strings.Split(scanner.Text(), ",")
		masterList = append(masterList, line)
	}
	return masterList
}
