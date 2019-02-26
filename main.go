/*
 * @Description: In User Settings Edit
 * @Author: caoyouming
 * @LastEditors: Please set LastEditors
 * @Date: 2019-02-20 16:06:55
 * @LastEditTime: 2019-02-26 15:20:43
 */
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/widuu/goini"
)

//错误信息
var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

//ini文件路径
var iniPath string = "./config/moose.conf"

//配置文件host_ip.conf解析后的结构
type Host_ip struct {
	Body []struct {
		Ip_list   []string
		Isp       string
		Region    string
		Source_ip string
		Status    string
		Ep_name   string
	}
	Taskname string
}

//输出结构体
type AllRes struct {
	Rtt       string
	Mdev      string
	Loss      string
	Score     int
	Mos       int
	Cip       string
	Timestamp string
	Line      string
	Isp       string
	Region    string
	Src_ip    string
	Status    string
	Ep_name   string
}

/**
 * @description: 代码前置
 * @param {type}
 * @return:
 */
func init() {
	conf := goini.SetConfig(iniPath)
	var logPathAndName string = conf.GetValue("log", "file")
	logFile, err := os.OpenFile(logPathAndName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	//defer logFile.Close()
	Info = log.New(io.MultiWriter(os.Stderr, logFile), "Info:", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(io.MultiWriter(logFile), "Warning:", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(os.Stderr, logFile), "Error:", log.Ldate|log.Ltime|log.Lshortfile)
}

/**
 * @description: 读取ini配置文件
 * @param {type}
 * @return:string
 */
func readIniString(filePath string, key1 string, key2 string) string {
	conf := goini.SetConfig(filePath)
	return conf.GetValue(key1, key2)
}

/**
 * @description: 读取文件
 * @param filePath {string}
 * @return:f []byte
 */
func readFile(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

/**
 * @description: 读取所有配置文件
 * @param iniPath {string}
 * @return:
 */
func allConfig(iniPath string) Host_ip {
	ipsPath := readIniString(iniPath, "ips", "ipList")

	fileContent, err := readFile(ipsPath)
	if err != nil {
		Error.Println(err)
		// fmt.Println(err)
	}
	//  fmt.Println(fileContent)
	var host_ip Host_ip
	err = json.Unmarshal([]byte(fileContent), &host_ip)
	if err != nil {
		Error.Println(err)
		// fmt.Println(err)
	}
	// fmt.Println(host_ip)
	return host_ip
}

/**
* @description: tcping
* @param ipAddress {string}
* @param port {string}
* @param network {string}
* @param timeOfTcping {int} 执行次数
* @return:
 */
func tcping(ipAddress string, port string, network string, timeOfTcping int) {
	var confTimeout = readIniString(iniPath, "worker", "timeout")
	timeInt, err := strconv.Atoi(confTimeout) //string转换为int
	if err != nil {
		Error.Println(err)
		os.Exit(0)
	}
	timeout := time.Duration(timeInt/1e3) * time.Millisecond //Millisecond 毫秒
	var address string = ipAddress + ":" + port
	sumSucc := 0
	sumFail := 0
	var t time.Duration
	var num []float64
	var sum float64
	for i := 0; i < timeOfTcping; i++ {
		startTime := time.Now()
		conn, err := net.DialTimeout(network, address, timeout)
		endTime := time.Now()
		t = endTime.Sub(startTime)
		num1 := float64(t) / float64(time.Millisecond)
		if err != nil {
			sumFail++
			Warning.Println(err)
			fmt.Printf("Probing %s:%s/%s - No response - time=%s \n", ipAddress, port, network, t)
		} else {
			sumSucc++
			num = append(num, num1)
			sum += num1
			fmt.Printf("Probing %s/%s - Port is open - time=%s \n", conn.RemoteAddr().String(), network, t)
		}
	}
	rateOfFailure := float64(sumFail / timeOfTcping * 100)
	fmt.Printf("\n")
	fmt.Printf("Ping statistics for %s \n", ipAddress)
	fmt.Printf("\t%d probes sent.\n", sumSucc)
	fmt.Printf("\t%d successful, %d failed.\t(%.2f%%\tfail)\n", sumSucc, sumFail, rateOfFailure)
	if sumFail != timeOfTcping {
		sort.Float64s(num)
		length := len(num)
		minimum := time.Duration(num[0]*1e3) * time.Microsecond
		maximum := time.Duration(num[length-1]*1e3) * time.Microsecond
		average := time.Duration((sum)/float64(length)*1e3) * time.Microsecond
		mdev := mdev(num)
		fmt.Printf("Approximate trip times in milli-seconds:\n")
		fmt.Printf("\tMinimum = %s, Maximum = %s, Average = %s, mdev = %.3fms\n", minimum, maximum, average, mdev)
	} else {
		fmt.Println("Was unable to connect, cannot provide trip statistics.")
	}
}

/**
 * @description:计算mdev
 * @param num {array}
 * @return:mdev
 */

func mdev(num []float64) (mdev float64) {
	// var num = []float64{4.44, 4.30, 4.19, 4.27}
	var sumSqu float64 = 0
	var sum float64 = 0
	length := float64(len(num))
	for i := 0; i < len(num); i++ {
		sum = sum + num[i]
		sumSqu = sumSqu + num[i]*num[i]
	}
	average := sum / length
	averageSqu := average * average
	squsum := sumSqu / length
	mdev = math.Sqrt(squsum - averageSqu)
	mdev, err := strconv.ParseFloat(fmt.Sprintf("%.4f", mdev), 64)
	if err != nil {
		Error.Println(err)
	}
	return mdev
}

/**
 * @description: 将struct数据转为json保存
 * @param {type}
 * @return:
 */
func structToJson() {
	data := AllRes{
		Rtt:       "1",
		Mdev:      "2.22",
		Loss:      "",
		Score:     0,
		Mos:       0,
		Cip:       "",
		Timestamp: "",
		Line:      "",
		Isp:       "",
		Region:    "",
		Src_ip:    "",
		Status:    "",
		Ep_name:   "",
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		Error.Println(err)
	}
	fmt.Println(string(jsonBytes))
}

/**
 * @description:
 * @param {type}
 * @return:
 */
func main() {
	hostIp := allConfig(iniPath)
	bodyLen := len(hostIp.Body)
	timeOfTcping := 4 //单个IP ping的次数
	port := "80"
	for i := 0; i < bodyLen; i++ {
		ipListLen := len(hostIp.Body[i].Ip_list)
		for j := 0; j < ipListLen; j++ {
			// fmt.Println(hostIp.Body[i].Ip_list[j])
			tcping(hostIp.Body[i].Ip_list[j], port, "tcp", timeOfTcping)
			os.Exit(0)
		}
	}
	// structToJson()
	/* 测试数据 */
	// ipAddress := "www.baidu.com"
	// var port string = "80"
	// network := "tcp"
	// num := 4
	// tcping(ipAddress, port, network, num)
	/*测试数据 */
	// arg_num := len(os.Args)
	// if arg_num < 2 {
	// 	Error.Println("无参数")
	// 	os.Exit(0)
	// }
	/* ipAddress := os.Args[0]
	var port string = "80"
	fmt.Println(ipAddress)
	fmt.Println(port)
	fmt.Println(os.Args[0]) */
}
