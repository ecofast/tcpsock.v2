# tcpsock.v2
Package tcpsock provides easy to use interfaces for TCP I/O, it's designed especially for developing online game servers.</br></br>

# How to use</br>
## there's a [chatroom](https://github.com/ecofast/tcpsock.v2/tree/master/samples/chatroom) demo which uses custom binary protocol</br>
run chatroom [server](https://github.com/ecofast/tcpsock.v2/tree/master/samples/chatroom/server) on a Aliyun VPS(CentOS 7.4) which has 2 CPUs of Intel(R) Xeon(R) E5-2682 v4 @ 2.50GHz, then make 1K connections by using [robot](https://github.com/ecofast/tcpsock.v2/tree/master/samples/chatroom/robot)(robot.exe -s=ip:port -n=100 -r=0[1,..9]):</br>
```shell
PID USER      PR  NI    VIRT    RES    SHR S  %CPU %MEM     TIME+ COMMAND
5156 root      20   0   76848  62636   1196 S  46.0  0.8   0:55.71 ./server
```
*****
![image](https://github.com/ecofast/tcpsock.v2/blob/master/samples/chatroom/server/server.png)</br>
*****
![image](https://github.com/ecofast/tcpsock.v2/blob/master/samples/chatroom/server/client.png)</br>
