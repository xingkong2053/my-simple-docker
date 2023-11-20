# my-simple-docker

### 运行程序之前需要先解压busybox文件系统
```shell
sudo docker pull busybox
sudo docker run -d busybox top -b
sudo docker export -o busybox.tar <container-id>
sudo tar -xvf busybox.tar -C /root/busybox
```
