package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

const compose string = "docker-compose.yml"

var backstage bool

func init() {
	flag.BoolVar(&backstage, "d", false, "容器是否后台运行")
	flag.Parse()
}

func main() {
	filePath, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}
	_, fn := path.Split(filePath)
	dir, err := os.ReadDir(filePath)
	if err != nil {
		log.Println(err)
	}
	var mainFile string
	for _, v := range dir {
		if strings.HasSuffix(v.Name(), ".go") {
			f, err := os.Open(path.Join(filePath, v.Name()))
			if err != nil {
				log.Println(err)
			}
			br := bufio.NewReader(f)
			line, _, err := br.ReadLine()
			if err == io.EOF {
				continue
			}
			if err != nil {
				log.Println(err)
			}
			if strings.EqualFold(string(line), "package main") {
				mainFile = v.Name()
			}
		}
	}
	f, err := os.Open(compose)
	var msg string = "%s file is exist.\n"
	if err != nil {
		if os.IsNotExist(err) {
			f, _ = os.Create(compose)
			defer f.Close()
			f.WriteString("version: '3'\n")
			f.WriteString("services:\n")
			f.WriteString("  app:\n")
			f.WriteString("    image: golang:latest\n")
			f.WriteString("    volumes:\n")
			f.WriteString("    - $PWD:/go/src/" + fn + "\n")
			f.WriteString("    ports:\n")
			f.WriteString("      - \"8080:8080\"\n")
			f.WriteString("    command: /go/src/" + fn + "/" + strings.TrimSuffix(mainFile, ".go"))
			msg = "%s file create completed.\n"
		}
	}
	log.Printf(msg, f.Name())
	execCmd("CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build " + mainFile)
	var startCmd string
	if backstage {
		startCmd = "docker-compose up -d"
	} else {
		startCmd = "docker-compose up"
	}
	execCmd(startCmd)
}

func execCmd(cmdLine string) {
	cmd := exec.Command("/bin/bash", "-c", cmdLine)
	// 创建获取命令输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error:can not obtain stdout pipe for command:%s\n", err)
		return
	}
	// 执行命令
	if err := cmd.Start(); err != nil {
		log.Println("Error:The command is err,", err)
		return
	}
	// 使用带缓冲的读取器
	outputBuf := bufio.NewReader(stdout)
	for {
		// 一次获取一行,_ 获取当前行是否被读完
		output, _, err := outputBuf.ReadLine()
		if err != nil {
			// 判断是否到文件的结尾了否则出错
			if err.Error() != "EOF" {
				log.Printf("Error :%s\n", err)
			}
			return
		}
		fmt.Printf("%s\n", string(output))
	}
	// wait 方法会一直阻塞到其所属的命令完全运行结束为止
	if err := cmd.Wait(); err != nil {
		log.Println("wait:", err.Error())
		return
	}
}
