package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"github.com/bitfield/script"
)

func newclient(apikey string) {
	if _, err := os.Stat("data.txt"); os.IsNotExist(err) {
		resp, _ := http.Get("https://api.masternetwork.dev/newclient/" + apikey)
		instance_encrypted_identification, _ := ioutil.ReadAll(resp.Body)
		identification := strings.ReplaceAll(string(instance_encrypted_identification), `"`, "")
		if string(identification) == "API_KEY_DOES_NOT_MATCH" {
			fmt.Println("Your Api key does not match. You won't make profits.")

		} else {

			f, _ := os.Create("data.txt")

			defer f.Close()

			_, err2 := f.WriteString(identification)
			if err2 != nil {
				log.Fatal(err2)
			}

			fmt.Println("Welcome !")
		}
	}
}

func main() {
	script.Exec("sudo apt-get update")
	script.Exec("sudo apt install qemu qemu-kvm libvirt-clients libvirt-daemon-system virtinst bridge-utils -y")
	script.Exec("sudo systemctl enable libvirtd")
	script.Exec("sudo systemctl start libvirtd")
	script.Exec("sudo apt-get install cpu-checker")
	script.Exec("sudo apt-get install python3")
	script.Exec("ulimit -n 250000") //avoid crash
	script.Exec("sudo apt-get install docker-ce docker-ce-cli containerd.io docker-compose-plugin3")
	script.Exec("curl -LOJ https://github.com/firecracker-microvm/firecracker/releases/download/v0.13.0/firecracker-v0.13.0")
	script.Exec("sudo mv firecracker-v0.13.0 firecracker")
	script.Exec("sudo chmod +x firecracker")
	script.Exec("sudo cp firecracker /usr/bin/")
	script.Exec("sudo setfacl -m u:${USER}:rw /dev/kvm")


	cmdkvm_ok, _ := exec.Command("kvm-ok").Output()

	kvm_ok := string(cmdkvm_ok)

	if strings.Contains(kvm_ok, "can be used") {

		api_key := flag.String("apikey", "Parmicciano", "api key from masternetwork.dev")
		ram_allowed := flag.Int("ram", 3500, "ram allowed")
		vcpus_allowed := flag.Int("vcpus", 2, "vcpus allowed")
		storage_allowed := flag.Int("storage", 20, "storage (G)")
		mode := flag.String("mode", "all", "different modes")
		nodename := flag.String("name", "default", "node name")

		flag.Parse()
		fmt.Println("apikey:", *api_key)
		fmt.Println("ram_allowed:", *ram_allowed)

		fmt.Println("vcpus_allowed:", *vcpus_allowed)
		fmt.Println("storage_allowed:", *storage_allowed)
		fmt.Println("mode:", *mode)
		fmt.Println("nodename:", *nodename)
		newclient(*api_key)
		if *api_key == "Parmicciano" {
			fmt.Println("fill with your apikey")
			os.Exit(1)
		}
		for true {
			tracking(*ram_allowed, *vcpus_allowed, *storage_allowed, *mode, *nodename)
		}

	} else {
		fmt.Println("your server does not support virtualization")
	}
	//kill_instance(instanceid)
	//check_impostor(instanceid)

	//   cmd := exec.Command("sh", "../shell.sh", "id-00233711122225")

}
