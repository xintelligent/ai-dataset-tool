package cmd

import (
	"bufio"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"math"
	"os"
	"strconv"
)

func getCategoryName(cn string) string {
	cnint, err := strconv.ParseInt(cn, 10, 64)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, v := range classes {
		if v.Index == int(cnint) {
			return v.Name
		}
	}
	return ""
}

func WriteFile(c string, f *os.File) error {
	writeObj := bufio.NewWriterSize(f, 1024)
	//使用Write方法,需要使用Writer对象的Flush方法将buffer中的数据刷到磁盘
	buf := []byte(c)
	if _, err := writeObj.Write(buf); err == nil {
		if err := writeObj.Flush(); err != nil {
			panic(err)
		}
		return nil
	} else {
		return err
	}
}

// ----------------------tool private-----
func pxToString(a interface{}) string {
	if v, p := a.(int); p {
		return strconv.Itoa(v)
	}
	if v, p := a.(int16); p {
		return strconv.Itoa(int(v))
	}
	if v, p := a.(int32); p {
		return strconv.Itoa(int(v))
	}
	if v, p := a.(uint); p {
		return strconv.Itoa(int(v))
	}
	if v, p := a.(float32); p {
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	}
	if v, p := a.(float64); p {
		return strconv.FormatFloat(v, 'f', -1, 32)
	}
	return ""
}
func toFloat64(unk interface{}) float64 {
	switch i := unk.(type) {
	case float64:
		return i
	case float32:
		return float64(i)
	case int64:
		return float64(i)
	case int:
		return float64(i)
	default:
		return math.NaN()
	}
}
