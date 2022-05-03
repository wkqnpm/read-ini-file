
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

type MysqlConfig struct {
	Password string `ini:"password"`
	Port     int    `ini:"port"`
	Username string `ini:"username"`
	Address  string `ini:"address"`
}
type MongoConfig struct {
	Url  string `ini:"url"`
	Port int    `ini:"port"`
}

type Config struct {
	MysqlConfig `ini:"mysql"`
	MongoConfig `ini:"mongo"`
}

//读文件
func ReadIniFile(filepath string) ([]byte, error) {
	res, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println("open file fiald")
		return nil, err
	}
	return res, err
}

func loadIni(filepath string, x interface{}) (err error) {
	//x 应为结构体指针类型
	t := reflect.TypeOf(x)
	if t.Kind() != reflect.Ptr {
		err = errors.New("invalid data,x type is ptr")
		return err
	}
	if t.Elem().Kind() != reflect.Struct {
		err = errors.New("invalid data x type is struct")
		return err
	}
	//读取配置文件
	res, err := ReadIniFile(filepath)
	if err != nil {
		return err
	}
	//分割文件
	strSplice := strings.Split(string(res), "\n")
	var structName string
	for index, item := range strSplice {
		item = strings.TrimSpace(item) //去除空格
		//判断注释，并跳过
		if strings.HasPrefix(item, ";") || strings.HasPrefix(item, "#") || len(item) == 0 {
			continue
		}
		if strings.HasPrefix(item, "[") {
			//判断类型 mysql mongo
			if item[0] != '[' || item[len(item)-1] != ']' {
				err = fmt.Errorf("line: %dsyntax error", index+1)
				return err
			}
			sectionName := strings.TrimSpace(item[1 : len(item)-1])
			if len(sectionName) == 0 { //当 [  ]时为false
				err = fmt.Errorf("line: %dsyntax error", index+1)
				return err
			}
			for i := 0; i < t.Elem().NumField(); i++ {
				field := t.Elem().Field(i)
				if sectionName == field.Tag.Get("ini") {
					structName = field.Name
					// fmt.Println(structName, sectionName)
					/*
						MysqlConfig mysql
						MongoConfig mongo
					*/
				}
			}
		} else {
			equalSignIndex := strings.Index(item, "=")
			if equalSignIndex == -1 || strings.HasPrefix(item, "=") {
				err = fmt.Errorf("line: %dsyntax error", index+1)
				return err
			}
			key := strings.TrimSpace(item[:equalSignIndex])
			value := strings.TrimSpace(item[equalSignIndex+1:])
			v := reflect.ValueOf(x)
			structobj := v.Elem().FieldByName(structName) //config 结构体的值
			structType := structobj.Type()                //config 结构体的类型
			// fmt.Println(structobj, structType)
			if structType.Kind() != reflect.Struct {
				err = fmt.Errorf("x中%s应该是结构体", structName)
			}
			var fieldName string
			//遍历嵌套结构体
			for i := 0; i < structobj.NumField(); i++ {
				field := structType.Field(i)
				if field.Tag.Get("ini") == key {
					//找到对应字段，如 xx=
					fieldName = field.Name
				}
			}
			if len(fieldName) == 0 { //在结构体中无此字段
				continue
			}
			//根据fileName取值
			fieldobj := structobj.FieldByName(fieldName)
			//赋值
			switch fieldobj.Type().Kind() {
			case reflect.String:
				{
					fieldobj.SetString(value)
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				{
					valInt, err := strconv.ParseInt(value, 10, 64)
					if err != nil {
						return err
					}
					fieldobj.SetInt(valInt)
				}
			}
		}

	}
	return nil
	// for i := 0; i < t.NumField(); i++ {
	// 	field := t.Field(i)
	// 	fmt.Printf("label:%s index:%d type:%v json tag:%v\n", field.Name, field.Index, field.Type, field.Tag.Get("ini"))
	// }
}

func main() {
	// i := MysqlConfig{}
	c := Config{}
	err := loadIni("./config.ini", &c)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(c.MysqlConfig, c.MongoConfig)
	// fmt.Printf("%#v", c)

}
