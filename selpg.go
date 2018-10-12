package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	flag "github.com/spf13/pflag"

)

//selpg的参数结构体
type selpg_args struct {
	start_page  int
	end_page    int
	page_len    int
	dest        string
	page_type   int
	in_filename string
}

//绑定参数*
func getArgs(args *selpg_args ) {
	pflag.IntVarP(&args.start_page,"start", "s", 0, "Start page of file")
    pflag.IntVarP(&args.end_page,"end","e", 0, "End page of file") 
    pflag.IntVarP(&args.page_len,"linenum", "l", 10, "lines in one page")
    pflag.StringVarP(&args.page_type,"printdes","f", "l", "flag splits page")
    pflag.StringVarP(&args.dest, "destination","d", "", "name of printer")
    pflag.Parse()
	otherArgs := pflag.Args()
	if len(otherArgs) > 0 {
		args.in_filename = otherArgs[0]
	} else {
		args.in_filename = ""
	}
}

//处理输入参数的情况
func process_args(args []string) {
	//参数数量不够
	if len(args) < 3 {
		fmt.Fprintf(os.Stderr, "%s: not enough arguments\n", progname)
		Error()
		os.Exit(1)
	}
	//处理第一个参数
	if args[1][0] != '-' || args[1][1] != 's' {
		fmt.Fprintf(os.Stderr, "%s: 1st arg should be -sstart_page\n", progname)
		Error()
		os.Exit(1)
	}
	//提取开始页数
	sp, _ := strconv.Atoi(args[1][2:])
	if sp < 1 {
		fmt.Fprintf(os.Stderr, "%s: invalid start page %d\n", progname, sp)
		Error()
		os.Exit(1)
	}
	sa.start_page = sp
	//处理第二个参数
	if args[2][0] != '-' || args[2][1] != 'e' {
		fmt.Fprintf(os.Stderr, "%s: 2nd arg should be -eend_page\n", progname)
		Error()
		os.Exit(1)
	}
	//提取结束页数
	ep, _ := strconv.Atoi(args[2][2:])
	if ep < 1 || ep < sp {
		fmt.Fprintf(os.Stderr, "%s: invalid end page %d\n", progname, ep)
		Error()
		os.Exit(1)
	}
	sa.end_page = ep

	//其他参数处理
	argindex := 3
	for {
		if argindex > argcount-1 || args[argindex][0] != '-' {
			break
		}
		switch args[argindex][1] {
		case 'l':
			//获取一页的长度
			pl, _ := strconv.Atoi(args[argindex][2:])
			if pl < 1 {
				fmt.Fprintf(os.Stderr, "%s: invalid page length %d\n", progname, pl)
				Error()
				os.Exit(1)
			}
			sa.page_len = pl
			argindex++
		case 'f':
			//以分页符号分页
			if len(args[argindex]) > 2 {
				fmt.Fprintf(os.Stderr, "%s: option should be \"-f\"\n", progname)
				Error()
				os.Exit(1)
			}
			sa.page_type = 'f'
			argindex++
		case 'd':
			//打印
			if len(args[argindex]) <= 2 {
				fmt.Fprintf(os.Stderr, "%s: -d option requires a printer destination\n", progname)
				Error()
				os.Exit(1)
			}
			sa.dest = args[argindex][2:]
			argindex++
		default:
			fmt.Fprintf(os.Stderr, "%s: error option", progname)
			Error()
			os.Exit(1)
		}
	}
    //获取文件名称
	if argindex <= argcount-1 {
		sa.in_filename = args[argindex]
	}
}

//文件读取
func process_input() {
	var cmd *exec.Cmd
	var cmd_in io.WriteCloser
	var cmd_out io.ReadCloser
	//
	if sa.dest != "" {
		cmd = exec.Command("bash", "-c", sa.dest)
		cmd_in, _ = cmd.StdinPipe()
		cmd_out, _ = cmd.StdoutPipe()
		//执行设定的命令
		cmd.Start()
	}
	if sa.in_filename != "" {
		inf, err := os.Open(sa.in_filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		line_count := 1
		page_count := 1
		fin := bufio.NewReader(inf)
		for {
			//读取输入文件中的一行数据
			line, _, err := fin.ReadLine()
			if err != io.EOF && err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if err == io.EOF {
				break
			}
			if page_count >= sa.start_page && page_count <= sa.end_page {
				if sa.dest == "" {
					//打印到屏幕
					fmt.Println(string(line))
				} else {
					//写入文件中
					fmt.Fprintln(cmd_in, string(line))
				}
			}
			line_count++
			if sa.page_type == 'l' {
				if line_count > sa.page_len {
					line_count = 1
					page_count++
				}
			} else {
				if string(line) == "\f" {
					page_count++
				}
			}
		}
		if sa.dest != "" {
			cmd_in.Close()
			cmdBytes, err := ioutil.ReadAll(cmd_out)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Print(string(cmdBytes))
			//等待command退出
			cmd.Wait()
		}
	} else {
		//从标准输入读取内容
		ns := bufio.NewScanner(os.Stdin)
		line_count := 1
		page_count := 1
		out := ""

		for ns.Scan() {
			line := ns.Text()
			line += "\n"
			if page_count >= sa.start_page && page_count <= sa.end_page {
				out += line
			}
			line_count++
			if sa.page_type == 'l' {
				if line_count > sa.page_len {
					line_count = 1
					page_count++
				}
			} else {
				if string(line) == "\f" {
					page_count++
				}
			}
		}
		if sa.dest == "" {
			fmt.Print(out)
		} else {
			fmt.Fprint(cmd_in, out)
			cmd_in.Close()
			cmdBytes, err := ioutil.ReadAll(cmd_out)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Print(string(cmdBytes))
			//等待command退出
			cmd.Wait()
		}
	}
}

//出错后报错信息
func Error() {
	fmt.Println("\nUsage of selpg.")
	fmt.Println("\tselpg -s=Number -e=Number [options] [filename]")
	fmt.Println("\t-l:Determine the number of lines per page and default is 10.")
	fmt.Println("\t-f:Determine the type and the way to be seprated.")
	fmt.Println("\t-d:Determine the destination of output.")
	fmt.Println("\t[filename]: Read input from this file.")
	fmt.Println("\tIf filename is not given, read input from stdin. and Ctrl+D to cut out.")
}

//全局变量
var sa selpg_args   
var argcount int    //参数的数量
var progname string  //

func main() {
	args := os.Args
	sa.start_page = 1
	sa.end_page = 1
	sa.in_filename = ""
	sa.dest = ""
	sa.page_len = 10
	sa.page_type = 'l'
	argcount = len(args)
	process_args(args)
	process_input()
}
