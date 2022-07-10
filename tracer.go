package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/bitfield/script"
	"github.com/digitalocean/go-libvirt"
	"github.com/gorilla/websocket"
)

type VMJSONSTRUCT struct {
	VM_NAME    string       `json:"VM_NAME"`
	VM_UUID    libvirt.UUID `json:"VM_UUID"`
	ACTIVE_VM  int32        `json:"ACTIVE_VM"`
	AUTOSTART  int32        `json:"AUTOSTART"`
	IMPOSTOR   bool         `json:"IMPOSTOR"`
	NODEKEY    string       `json:"NODEKEY"`
	NODENAME   string       `json:"NODENAME"`
	NODEMODE   string       `json:"NODEMODE"`
	CPUMODEL   [32]int8     `json:"CPUMODEL"`
	RAM        uint64       `json:"Memorysize"`
	Storage    int32        `json:"Storage"`
	CPUCORES   int32        `json:"CPUCORES"`
	CPUMHZ     int32        `json:"CPUMHZ"`
	RNODES     int32        `json:"RNODES"`
	RSOCKETS   int32        `json:"RSOCKETS"`
	RCORES     int32        `json:"RCORES"`
	RTHREADS   int32        `json:"RTHREADS"`
	VCPUSTOTAL int32          `json:"VCPUSTOTAL"`
	VM_VCPUS   int32        `json:"VM_VCPUS"`
	VM_MEMORY  int          `json:"VM_MEMORY"`
}

func Newinstance(instancename string, instance_ram int32, instance_vcpus int8, instance_storage int16, instance_os string) {
	fmt.Println("Yes ! New vm in creation !")
	script.Exec("sudo apt install cloud-image-utils -y").Stdout()
	script.Exec("sudo apt install virtinst -y").Stdout()
	script.Exec("sudo apt install virt-manager -y").Stdout()
	script.Exec("sudo apt install guestfish -y").Stdout()
	script.Exec("wget https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64-disk-kvm.img").Stdout()
	script.Exec("sudo rm wget-log").Stdout()
	script.Exec("sudo cloud-localds /var/lib/libvirt/images/" + instancename + ".img " + instancename + ".txt").Stdout()
	script.Exec("sudo rm " + instancename + ".txt").Stdout()
	script.Exec("sudo qemu-img convert -f qcow2 jammy-server-cloudimg-amd64-disk-kvm.img /var/lib/libvirt/images/1604" + instancename + ".img").Stdout()
	script.Exec("sudo qemu-img resize  /var/lib/libvirt/images/1604" + instancename + ".img +"+ strconv.Itoa(int(instance_storage))+"G").Stdout()
	script.Exec("sudo rm jammy-server-cloudimg-amd64-disk-kvm.img").Stdout()
	script.Exec("sudo qemu-img create -f qcow2 /var/lib/libvirt/images/disk" + instancename + ".qcow2 " + strconv.Itoa(int(instance_storage)) + " -o preallocation=full").Stdout()
	script.Exec("sudo virt-install --name " + instancename + " --vcpus " + strconv.Itoa(int(instance_vcpus)) + " --memory " + strconv.Itoa(int(instance_ram)) + " --disk /var/lib/libvirt/images/1604" + instancename + ".img,device=disk,bus=virtio --disk /var/lib/libvirt/images/" + instancename + ".img,device=cdrom --os-variant " + instance_os + " --virt-type kvm --graphics none --noautoconsole --network network=default,model=virtio --import ").Stdout()
	script.Exec("sudo virsh attach-disk --domain " + instancename + " /var/lib/libvirt/images/disk" + instancename + ".qcow2  --target vdb --persistent --config --live").Stdout()
}
func kill_instance(instanceid string) {
	script.Exec("sudo virsh destroy  " + instanceid)
	script.Exec("sudo virsh undefine " + instanceid)
}


func check_impostor(instanceid string) bool {

	exec.Command("/bin/bash", `PROMT_COMMAND="history -a; history -r"`).Output()

	cmd_usr_history, _ := exec.Command("tail", "/home/"+os.Getenv("SUDO_USER")+"/.bash_history", "").Output()

	user_history := string(cmd_usr_history)

	var is_impostor bool
	is_impostor = strings.Contains(user_history, instanceid)

	return is_impostor

}

func instance_is_alive(instance_name string) (is_alive int8) {

	connection, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(connection)

	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	domains, err := l.Domains()
	if err != nil {
		log.Fatalf("failed to retrieve domains: %v", err)
	}

	is_alive = 0

	for _, d := range domains {
		if instance_name == d.Name {
			is_alive = 1
		}

	}

	if err := l.Disconnect(); err != nil {
		log.Fatalf("failed to disconnect: %v", err)
	}
	return is_alive
}

func stopvm(instance_name string) {
	connection, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(connection)

	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	domains, err := l.Domains()
	if err != nil {
		log.Fatalf("failed to retrieve domains: %v", err)
	}


	for _, d := range domains {
		if instance_name == d.Name {
			l.DomainShutdown(d)
		}

	}

	if err := l.Disconnect(); err != nil {
		log.Fatalf("failed to disconnect: %v", err)
	}
	

}

func startvm(instance_name string) {
	script.Exec("sudo virsh start  " + instance_name)
}
func instance_info(DEFINED_ram int, DEFINED_vcpus int, DEFINED_storage int, DEFINED_mode string, DEFINED_nodename string) []VMJSONSTRUCT {

	connection, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(connection)

	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	domains, err := l.Domains()
	if err != nil {
		log.Fatalf("failed to retrieve domains: %v", err)
	}

	var VMJSON []VMJSONSTRUCT
	// Unmarshall it
	err = json.Unmarshal([]byte(`[]`), &VMJSON)
	if err != nil {
		log.Fatalf("failed to unmarshall json : %v", err)
	}
	NODE_KEY, err := ioutil.ReadFile("./data.txt")
	if err != nil {
		fmt.Println(err)
	}

	

	Model, MemorySize, CPUs, CPUmhz, rNodes, CPUsockets, CPUcorespersocket, CPUthreadpercore, _ := l.NodeGetInfo()

	if DEFINED_ram < int(MemorySize) {
		MemorySize = uint64(DEFINED_ram) *1000
	}
	if DEFINED_vcpus < int(CPUs) {
		CPUs = int32(DEFINED_vcpus)
	}
	NODE_VCPUS_TOTAL := CPUs

	VMJSON = append(VMJSON, VMJSONSTRUCT{VM_NAME: "Masternetwork", NODEKEY: string(NODE_KEY), CPUMODEL: Model, RAM: MemorySize, CPUCORES: CPUs, CPUMHZ: CPUmhz, RNODES: rNodes, RSOCKETS: CPUsockets, RCORES: CPUcorespersocket, RTHREADS: CPUthreadpercore, VCPUSTOTAL: NODE_VCPUS_TOTAL, Storage: int32(DEFINED_storage), NODEMODE: DEFINED_mode, NODENAME: DEFINED_nodename})

	for _, d := range domains {
		is_impostor := check_impostor(d.Name)
		domain_auto_start, _ := l.DomainGetAutostart(d)
		domain_active, _ := l.DomainIsActive(d)

		//DomainGetMaxVcpus, _ := l.DomainGetMaxVcpus(d)
	

		VIRTUAL_MACHINE_VCPUS, _ := l.DomainGetMaxVcpus(d)
		VIRTUAL_MACHINE_MEMORY, _ := l.DomainGetMaxMemory(d)

		VMJSON = append(VMJSON, VMJSONSTRUCT{VM_NAME: d.Name, VM_UUID: d.UUID, ACTIVE_VM: domain_active, AUTOSTART: domain_auto_start, IMPOSTOR: is_impostor, NODEKEY: string(NODE_KEY), CPUMODEL: Model, RAM: MemorySize, CPUCORES: CPUs, CPUMHZ: CPUmhz, RNODES: rNodes, RSOCKETS: CPUsockets, RCORES: CPUcorespersocket, RTHREADS: CPUthreadpercore, VCPUSTOTAL: NODE_VCPUS_TOTAL, VM_VCPUS: VIRTUAL_MACHINE_VCPUS, VM_MEMORY: int(VIRTUAL_MACHINE_MEMORY), Storage: int32(DEFINED_storage), NODEMODE: DEFINED_mode, NODENAME: DEFINED_nodename})

	}

	if err := l.Disconnect(); err != nil {
		log.Fatalf("failed to disconnect: %v", err)
	}
	return VMJSON
}

func tracking(ram int, vcpus int, storage int, mode string, nodename string) {
	websocket_communication(ram, vcpus, storage, mode, nodename)
}

var done chan interface{}
var interrupt chan os.Signal

type NewVm struct {

	// defining struct variables
	VM_NAME    string
	VM_RAM     int32
	VM_VCPUS   int8
	VM_STORAGE int16
	VM_OS      string
	VM_STATE_WANTED int 
}

func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)

		var newvm []NewVm
		json.Unmarshal([]byte(msg), &newvm)
	
		for i := range newvm {
			//newvm[i].VM_NAME

			if newvm[i].VM_STATE_WANTED == 0 {
				stopvm(newvm[i].VM_NAME)
			}
			if newvm[i].VM_STATE_WANTED == 1 {
				startvm(newvm[i].VM_NAME)
			}
			if newvm[i].VM_STATE_WANTED == -42 {
				kill_instance(newvm[i].VM_NAME)
			}
			if instance_is_alive(newvm[i].VM_NAME) == 0 && newvm[i].VM_STATE_WANTED == 1 {
				resp, _ := http.Get("https://api.masternetwork.dev/GetCLOUDinit/" + newvm[i].VM_NAME)
				body, _ := ioutil.ReadAll(resp.Body)

				f, _ := os.Create(newvm[i].VM_NAME + ".txt")
				defer f.Close()

				data := strings.ReplaceAll(string(body), "<br>", "")
				_, err2 := f.WriteString(data)

				if err2 != nil {
					log.Fatal(err2)
				}

				fmt.Println("done")
				Newinstance(newvm[i].VM_NAME, newvm[i].VM_RAM, newvm[i].VM_VCPUS, newvm[i].VM_STORAGE, newvm[i].VM_OS)
			}

		}

	}
}

func websocket_communication(ram int, vcpus int, storage int, mode string, nodename string) {

	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	socketUrl := "wss://api.masternetwork.dev:443" + "/client_connection"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)

	}
	defer conn.Close()
	go receiveHandler(conn)

	// Our main loop for the client
	// We send our relevant packets here
	for {
		select {
		case <-time.After(time.Duration(5) * time.Millisecond * 1000):
			// Send an echo packet every second
			VMJSON := instance_info(ram, vcpus, storage, mode, nodename)

			result, _ := json.Marshal(VMJSON)

			err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(string(result))))
			if err != nil {
				log.Println("Error during writing to websocket:", err)
				return
			}

		case <-interrupt:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return
		}
	}
}
