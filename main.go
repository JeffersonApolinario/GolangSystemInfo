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
    "time"
)


// Consts.
const VALUEDIVISOR uint64 = 1024

type (
    DiskInfo struct {

        Filesystem string `json:"filesystem"`
        Total uint64 `json:"size"`
        Used uint64 `json:"used"`
        Free uint64 `json:"available"`
        UsedPercent float64 `json:"capacity"`
        Amount string `json:"amout"`

    }

    HostInfo struct {
    
        Uptime               uint64 `json:"uptime"`
        OS                   string `json:"os"`              
        Platform             string `json:"platform"`        
        PlatformFamily       string `json:"platformFamily"`  
        PlatformVersion      string `json:"platformVersion"` 
    
    }

    VirtualMemoryStat struct {

        // Total amount of RAM on this system
        Total uint64 `json:"total"`
        // Free ram on this system
        Free uint64 `json:"free"`
    }

    OsInfo struct {

        Arch string `json:"arch"`
        FreeMem uint64  `json:"freemem"`
        HomeDir string `json:"homedir"`
        Hostname string `json:"hostname"`
        LoadAvg [3] float64 `json:"loadavg"`
        Platform string `json:"platform"`
        Release string `json:"release"`
        TmpDir string `json:"tmpdir"`
        TotalMem uint64  `json:"totalmem"`
        Uptime uint64 `json:"uptime"`
    }   

    RequestBody struct {

        Command string `json:"command"`
        Options map[string]interface{} `json:"options"`
    }

    ResponseError struct {

        Error bool `json:"error"`
        Killed bool `json:"killed"`
        Code int `json:"code"`
        Signal int `json:"signal"`
        Cmd string `json:"cmd"`
        StdOut string `json:"stdout"`
        StdErr string `json:"stderr"` 
    }

    ResponseSuccess struct {

        StdOut string `json:"stdout"`
        StdErr string `json:"stderr"`
    }

    SystemInfo struct {

        OS OsInfo `json:"os"`
        Disk []DiskInfo `json:"disk"`
    }
)

// Init to app and define routes.

func main() {    
    e := echo.New()

    e.POST("/execute", execute)
    e.GET("/ping", ping)
    e.GET("/info", info)

    e.Logger.Fatal(e.Start(":1323"))
}

// when request /execute this method is call, this method get command json and execute.
func execute(c echo.Context) error {
    
    var(
        statusExit int
        timer *time.Timer
        timeoutProcessExit float64 
    ) 
    

    requestBody := new(RequestBody)

    if err := c.Bind(requestBody); err != nil {
        fmt.Println(err)    
        return err

    }

    cmd := exec.Command("bash", "-c", string(requestBody.Command))
    
    if len(requestBody.Options) > 0 {
        for key, value := range requestBody.Options {
            if ( key == "cwd") {
                if ExistsDirectoty(fmt.Sprint(value)) {
                    cmd.Dir = fmt.Sprint(value);    
                }           
            }
            if (key == "timeout") {
                timeoutProcessExit = value.(float64) 
                if timeoutProcessExit > 0 {
                    timer = time.AfterFunc(time.Duration(timeoutProcessExit) * time.Millisecond, func() {       
                        timer.Stop()
                        cmd.Process.Kill()
                    })  
                }               
            }
        }
    }
        
    stdout, _ := cmd.StdoutPipe()
    stderr, _ := cmd.StderrPipe()
    cmd.Start()
    stdoutStr, _ := ioutil.ReadAll(stdout)
    stderrStr, _ := ioutil.ReadAll(stderr)
    
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
            Cmd : "bash -c " + string(requestBody.Command),
            StdOut : "",
            StdErr : string(stderrStr)}

        return c.JSON(http.StatusOK,result)
    }

    result := ResponseSuccess{StdOut : string(stdoutStr),StdErr : ""}

    return c.JSON(http.StatusOK, result)
}

func ping(c echo.Context) error {
    return c.JSON(http.StatusOK, "pong")
}
// when to request /info call this method, this method response info of system.
func info(c echo.Context) error {

    usr, err := user.Current()
    if err != nil {
        log.Fatal( err )
    }

    var (
        memInfo VirtualMemoryStat
        hostInfo HostInfo
        osInfo OsInfo
        systemInfo SystemInfo
    ) 
       
    
    memInfo = getMemInfo()
    hostInfo = getHostInfo()

    osInfo.Hostname = usr.Username
    osInfo.HomeDir = usr.HomeDir
    osInfo.TmpDir = os.Getenv("TMPDIR")
    osInfo.Arch = runtime.GOARCH
    osInfo.Platform = hostInfo.Platform
    osInfo.FreeMem = memInfo.Free
    osInfo.TotalMem = memInfo.Total
    osInfo.Release = hostInfo.PlatformVersion
    osInfo.Uptime = hostInfo.Uptime
    osInfo.LoadAvg = getLoadAverages()
        
    systemInfo.OS = osInfo
    systemInfo.Disk = getDrives()

    return c.JSON(http.StatusOK, systemInfo)
}

// this method get Memory Ram informations and return to struct. 
func getMemInfo() (memoryState VirtualMemoryStat) {

    memory, err := mem.VirtualMemory()
    if err != nil {
        log.Fatal(err)
        fmt.Println("erro informacos da memoria")
    }

    memoryState.Total = memory.Total
    memoryState.Free = memory.Free

    return 

}

// this method get Host and platform informations and return to struct. 
func getHostInfo() (hostInfo HostInfo) {

    hostSystem, err := host.Info()
    if err != nil {
        log.Fatal(err)
        fmt.Println("erro informacos do host")
    }

    hostInfo.Uptime = hostSystem.Uptime
    hostInfo.Platform = hostSystem.Platform
    hostInfo.PlatformFamily = hostSystem.PlatformFamily
    hostInfo.PlatformVersion = hostSystem.PlatformVersion

    return

}

// this method get Disks informations and return to struct. 
func getDrives() (disks []DiskInfo) {

    var (
        diskAmount DiskInfo
        percentage float64
    )

    partitions, err := disk.Partitions(true)
    if err != nil {
        log.Fatal(err)
        fmt.Println("erro procurando particoes")
    }

    for _, partition := range partitions {

        diskPartition, err := disk.Usage(partition.Mountpoint)
        if err != nil {
            if err.Error() == "no such file or directory" {
                continue
            }
            log.Fatal(err)
            fmt.Println("erro procurando informacoes do disco")
        }

        if(math.IsNaN(diskPartition.UsedPercent)){
            percentage = 1;
        } else {
            percentage = diskPartition.UsedPercent / 100
        }

        percentage = RoundUp(percentage,2)

        diskAmount.Filesystem = partition.Device
        diskAmount.Total = divisor(diskPartition.Total)
        diskAmount.Amount = partition.Mountpoint
        diskAmount.Free = divisor(diskPartition.Free)
        diskAmount.UsedPercent = percentage
        diskAmount.Used = divisor(diskPartition.Used)

        disks = append(disks,diskAmount)
    }
    return

}
// this method serve to divider values by VALUEDIVISOR
func divisor(value uint64) uint64{

    return value / VALUEDIVISOR
}

// this method get LoadAverages and return a array with loads
func getLoadAverages() (loadAverages [3]float64) { 

    loadAvg, err := load.Avg()
    if err != nil {
        log.Fatal(err)
        fmt.Println("erro ao pegar os loadaverage")
    }

    loadAverages[0] = loadAvg.Load1
    loadAverages[1] = loadAvg.Load5
    loadAverages[2] = loadAvg.Load15

    return

}

//this method RoundUp and return value rouded, primary param serve to indicate number to round and second serve to Fixed Point.
func RoundUp(input float64, places int) (newVal float64) {

    var round float64
    pow := math.Pow(10, float64(places))
    digit := pow * input
    round = math.Ceil(digit)
    newVal = round / pow
    return

 }

 func ExistsDirectoty(name string) bool {
    if _, err := os.Stat(name); err != nil {
        if os.IsNotExist(err) {
            return false
        }
    }
    return true
}