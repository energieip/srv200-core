Server Core: Service Management for configuration and data
==========================================================

Server core is responsible for:
* managing configuration database for all switchs/servers/drivers configuration and authorization
* setup-ing switchs 
* sending command to switchs

Build Requirement: 
* golang-go > 1.9
* glide
* devscripts
* make

Run dependancies:
* rethindkb
* mosquitto

To compile it:
* GOPATH needs to be configured, for example:
```
    export GOPATH=$HOME/go
```

* Install go dependancies:
```
    make prepare
```

* To clean build tree:
```
    make clean
```

* Multi-target build:
```
    make all
```

* To build x86 target:
```
    make bin/sensorservice-amd64
```

* To build armhf target:
```
    make bin/sensorservice-armhf
```
* To create debian archive for x86:
```
    make deb-amd64
```

* To install debian archive on the target:
```
    scp build/*.deb <login>@<ip>:~/
    ssh <login>@<ip>
    sudo dpkg -i *.deb
```

For development:
* recommanded logger: *rlog*
* For network connection: use *common-network-go* library
* For database management: use *common-database-go* library
