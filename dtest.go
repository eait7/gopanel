package main
import "os/exec"
import "fmt"
func main(){ out, err := exec.Command("docker", "cp", "gopanel-dashboard:/app/data", "/tmp/data2").CombinedOutput(); fmt.Println(string(out), err) }
