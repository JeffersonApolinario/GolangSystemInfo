package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
    "github.com/shirou/gopsutil/load"
	"net/http"
	"os"
	"os/user"
	"os/exec"
	"io/ioutil"
	"log"
	"syscall"
	"runtime"
    "math"
)

type JsonRequest struct {
	Command string `json:"command"`
}

type ResponseOK struct {
	StdOut string `json:"stdout"`
	StdErr string `json:"stderr"`
}


type ResponseError struct {
	Error bool `json:"error"`
	Killed bool `json:"killed"`
	Code int `json:"code"`
	Signal int `json:"signal"`
	Cmd string `json:"cmd"`
	StdOut string `json:"stdout"`
	StdErr string `json:"stderr"` 
}



type OsInfo struct {
    Arch string `json:"arch"`
    FreeMem uint64  `json:"freemem"`
	HomeDir string `json:"homedir"`
	Hostname string `json:"hostname"`
    LoadAvg [3] float64 `json:"loadavg"`
	Platform string `json:"platform"`
	Release string `json:"release"`
	// Type string `json:"type"`
	TmpDir string `json:"tmpdir"`
    TotalMem uint64  `json:"totalmem"`
	Uptime uint64 `json:"uptime"`
}

type DiskInfo struct {
	Filesystem string `json:"filesystem"`
	Total uint64 `json:"size"`
	Used uint64 `json:"used"`
	Free uint64 `json:"available"`
    UsedPercent float64 `json:"capacity"`
	Amount string `json:"amout"`

}

type Info struct {
	OS OsInfo `json:"os"`
	Disk []DiskInfo `json:"disk"`
}

type VirtualMemoryStat struct {
    Total uint64 `json:"total"`
    Free uint64 `json:"free"`
}


type HostInfo struct {
 	
    Uptime               uint64 `json:"uptime"`
    OS                   string `json:"os"`              // ex: freebsd, linux
    Platform             string `json:"platform"`        // ex: ubuntu, linuxmint
    PlatformFamily       string `json:"platformFamily"`  // ex: debian, rhel
    PlatformVersion      string `json:"platformVersion"` // version of the complete OS
    
}

func main() {
	e := echo.New()

	e.POST("/execute", execute)
	e.GET("/ping",ping)
	e.GET("/info",info)

	e.Logger.Fatal(e.Start(":1323"))
}

func execute(c echo.Context) error {
	
	r := new(JsonRequest)

	if err := c.Bind(r); err != nil {
		return err
	}

	cmd := exec.Command("bash", "-c", string(r.Command))

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()
	stdoutStr, _ := ioutil.ReadAll(stdout)
	stderrStr, _ := ioutil.ReadAll(stderr)
	
	var statusExit int
	if err := cmd.Wait(); err != nil {
        if exiterr, ok := err.(*exec.ExitError); ok {
            if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
                statusExit = status.ExitStatus()
            }
        } 
    }


    if len(string(stderrStr)) > 0 {
   		result := ResponseError{
   			Error : true,
   			Killed : false,
   			Code : statusExit,
   			Signal : 0,
   			Cmd : "bash -c " + string(r.Command),
   			StdOut : "",
   			StdErr : string(stderrStr)}
        return c.JSON(http.StatusOK,result)
    }
	result := ResponseOK{
		StdOut : string(stdoutStr),
		StdErr : ""} 
	return c.JSON(http.StatusOK, result)
}


func ping(c echo.Context) error {
	return c.JSON(http.StatusOK, "pong")
}

func info(c echo.Context)  error {
	usr, err := user.Current()
    if err != nil {
        log.Fatal( err )
    }
    var memInfo VirtualMemoryStat
    var hostInfo HostInfo
    memInfo = getMemInfo()
    hostInfo = getHostInfo()

    osInfo := OsInfo{
    	Hostname : usr.Username,
    	HomeDir :  usr.HomeDir,
    	TmpDir : os.Getenv("TMPDIR"),
    	Arch : runtime.GOARCH,
    	Platform : hostInfo.Platform,
    	FreeMem : memInfo.Free,
    	TotalMem : memInfo.Total,
    	Release : hostInfo.PlatformVersion,
    	Uptime : hostInfo.Uptime,
        LoadAvg : getLoadAverages()}

    
    info := Info {
    	OS : osInfo,
    	Disk : getDrives()}	
 	

	return c.JSON(http.StatusOK, info)
}

func getMemInfo() VirtualMemoryStat {

	var MemoryStat VirtualMemoryStat
	memory, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal(err)
        fmt.Println("erro informacos da memoria")
	}

	MemoryStat = VirtualMemoryStat{Total : memory.Total, Free : memory.Free}

	return MemoryStat	
}


func getHostInfo() HostInfo {
	var hostInfo HostInfo

	hostNow, err := host.Info()
	if err != nil {
        log.Fatal(err)
        fmt.Println("erro informacos do host")
    }

    hostInfo = HostInfo{Uptime : hostNow.Uptime,
                        Platform : hostNow.Platform,
    			        PlatformFamily : hostNow.PlatformFamily,
                        PlatformVersion : hostNow.PlatformVersion}

    return hostInfo
}

func getDrives() []DiskInfo  {

	var disks []DiskInfo
	var diskAppend DiskInfo
    var percentage float64
	partitions, err := disk.Partitions(true)
    if err != nil {
        log.Fatal(err)
        fmt.Println("erro procurando particoes")
    }

    for _, p := range partitions {

        d, err := disk.Usage(p.Mountpoint)
        if err != nil {
            if err.Error() == "no such file or directory" {
                continue
            }
            log.Fatal(err)
            fmt.Println("erro procurando informacoes do disco")
        }

        if(math.IsNaN(d.UsedPercent)){
            percentage = 1;
        } else {
            percentage = d.UsedPercent / 100
        }

        percentage = RoundUp(percentage,2)

        diskAppend = DiskInfo{Filesystem : p.Device, Total:d.Total / 1024,
                              Used : d.Used / 1024,
                              Amount : p.Mountpoint,
                              Free : d.Free / 1024,
                              UsedPercent : percentage}

        disks = append(disks,diskAppend)
    }
    return disks
}

func getLoadAverages() [3] float64 {
   
    var loadAverages [3] float64 
    loadAvg, err := load.Avg()
    if err != nil {
        log.Fatal(err)
        fmt.Println("erro ao pegar os loadaverage")
    }

    loadAverages[0] = loadAvg.Load1
    loadAverages[1] = loadAvg.Load5
    loadAverages[2] = loadAvg.Load15

    return loadAverages
}

func RoundUp(input float64, places int) (newVal float64) {
    var round float64
    pow := math.Pow(10, float64(places))
    digit := pow * input
    round = math.Ceil(digit)
    newVal = round / pow
    return
 }

