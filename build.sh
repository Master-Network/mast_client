sudo rm /home/parmicciano/masternetwork/masternetwork_client/amd64_linux
sudo rm /home/parmicciano/masternetwork/masternetwork_client/arm_linux
sudo rm /home/parmicciano/masternetwork/masternetwork_client/arm64_linux
env GOOS=linux GOARCH=amd64 go build -o /home/parmicciano/masternetwork/masternetwork_client/amd64_linux
env GOOS=linux GOARCH=arm GOARM=5 go build -o /home/parmicciano/masternetwork/masternetwork_client/arm_linux
env GOOS=linux GOARCH=arm64 GOARM=5 go build -o /home/parmicciano/masternetwork/masternetwork_client/arm64_linux
