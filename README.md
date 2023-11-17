# go-internet-failover

go-internet-failover service (referred as `IFS` elsewhere in this document) is internet failover service. It allows automatic switching of internet connections in case of failure of any of the internet connections. The end user ranks the various internet connections based on their preferences, and IFS ensures that the highest ranked working internet connection gets configured, and ready for use. Note it is not a load balancing service as that is not its intent. Further, it works with **IPv4** only.

## Prerequisites

1. **Linux router**: This is the machine where IFS will be running. Note, IFS supports only GNU/Linux. As a prerequisite, the router needs to be configured to forward all packets from your local network (LAN) to internet (WAN). All the different WANs need to be reachable from the Linux router. Additionally following commands should be present: `sudo`, `ip`, `iptables`, `ping`.
2. **Disable Ipv6**: IPv6 needs to be disabled on your LAN. This is not a prerequisite per say only that your IPv6 traffic will not behave actually make use of failover capabilities offered by `IFS`.
3. **Linux knowhow**: You should be sufficiently comfortable with using Linux command line and should know networking basics. IFS is not at a stage where it can detect and autoconfigure itself as per your network configuration.

## Typical usage

Typical usage is when you have a primary broadband connection which typically caters to your internet requirements, and you have one secondary internet connection, to which you want to fallback in case your primary internet connection fails. The fallback to secondary connection should be automatic, and once primary internet connection becomes available, it should again switch back to the primary internet connection.

![Architecture Diagram](https://cdn-0.plantuml.com/plantuml/png/NP1FR_8m38Vl_XGMTtZLz8eGhxJB9FP7ko8EfJ4bMfeWDu6XgTzzSL9jAhSyJtwn7M-7q728lONM-gZn6n3tpouGEme7R0OqC084tr4u4vVxTpPzmtTyyPfwh7ACAxafHXMZ_anTJ0qZE1y8Zpu4twC_YKDyS_QEtf68r0OxSoMNV2-F_x6FWNZ0cx4MZvHy74ZZoJEQQL9isfQ60H2RT7RtKW63wMa5v3HALm0mk5x5oseZuygPJNSEWYgZzZSdL0DmXKl1_TbeJUwmp25M3xQ4BqtxsNI4Yzt_rqNPqSwu-8KVUVJyUSj9p_QOKazqDIrDAzJLJAeYrKpMHShIHXZpcEiyqN8Z7LWbp9-Qk4uSBmN0yaJIIk0tgbNXKadgWtvOG4bfDRSbl2MdOsz_0000)

## Building and installing

To build the code execute command `make build`. This will build executable: `bin/goifs`. Executing command `./bin/goifs` will print the command line usage instructions.

To build deb installer package (which can be installed on Debian, Ubuntu, Raspbian OS among others) execute the command `make debian`. The package gets built in the directory `build/debian/goifs.deb`.

To install the above command using **sudo** execute the command `make debian-install`.

Alternatively, you can build, and install in one step simply by executing command: `make debian-install`.

## Configuration

The default configuration file resides in `/etc/goifs/ifs.yaml`.

### Global configuration parameters

| Attribute | Type | Default Value | Description |
| --------- | ---- | ------------- | ----------- |
| maxPacketLoss | integer | 50 | The maximum number of failed pings. |
| minPacketLoss | integer | 20 | The minimum number of failed pings. |
| minSuccessivePacketsRecvd | integer | 20 | The minimum number of consecutive received pings. Used to decide whether connection is up. |
| maxSuccessivePacketsLost | integer | 10 | The maximum number of consecutive lost pings. Used to decide whether connection is down. |
| useSudo | boolean | false | Use sudo for executing commmands. Useful if you want to run as non-root user. |
| cleanIfRequired | boolean | true | Clean network configuration if required. Useful if `IFS` did not clean up network configuration during last run. |
| ping | string | 1.1.1.1 | IP address to be pinged for determining network connection status. |
| connections | array | | Internet connections configuration. Each item of the array is as defined in the next section. |

### Connection configuration parameters

| Attribute | Type | Description |
| --------- | ---- | ----------- |
| name | string | Name of connection. Should be unique and follow all rules of Linux network namespace naming. |
| rank | integer | Priority/rank of network connection for failover. Lower value means higher priority. |
| ip | string | IPv4 address of virtual interface created in global namespace. |
| peerIp | string | IPv4 address of virtual interface create in network namespace. |
| mask | integer | Network mask used for the network |
| gwIf | string | Interface where Gateway is available. |
| gw | string | Gateway IP address. |
| mark | integer | Unique mark to be used for packets. |

### Decision algorithm

For each connection, this service maintains a record of
1. Number of successful pings in last 100 pings (A)
2. Number of failed pings in the last 100 pings (B)
3. Number of consecutive successful pings (C) 
4. Number of consecutive failed pings (D)

When connection status is up: if `B > maxPacketLoss` or `D >= maxSuccessivePktsLost` the connection status is marked as down.

When connection status is down: if `B <= minPacketLoss` or `C >= minSuccessivePktsRcved` the connection status is marked as up.

Once connection status changes for any connection from up to down or vice-versa, the connection with better priority/rank (lower number has better priority) having status as up is activated.

### Example

Consider the case where there are two internet connections - broadband and 4G. The requirement is to have broadband as primary connection and 4G as failover internet connection. The broadband is on network `192.168.1.0/24` and 4G on network `192.168.3.0/24`. Here we assume that network forwarding with IP Masquerade is enabled on Linux router.

The following diagrams illustrates the original configuration where broadband connection is being used:

![Broadband Architecture Diagram](https://www.planttext.com/api/plantuml/png/fLFVIyCm47xFNt7u8XGxJJP4HSkGmHHKP8ZlsN9pOMaoqYjpYl-xNTfkcTsR3ot9TtVVZoiT5YGzbRbX1kS4vC3hZmM1qXvdf9rbf_3WlFvobgG1eoqMDv2csHdSfcbuPLNBEthIiwWXrBTFnaYJGNX1MZk_XBMc1KozaseghM9iPbofG5j2Rv60iLoku2H9xjkM93a1MK3k5EOSlXd0ssQ5S9V1hgN29j8kjnYlpe-luNGjUlpogiTxdFtiQ0VZc4ySy0t64P7e4FKRejf8bNldcZLa1rYh-AHj-GaJLrPue-t39MWZBIuwXiMv6BIriIHSoytVHA7AEmxta_pOP3zCZd0kIqEr9qYshkjGhMO7uX4aecsEi5YIMpMnL4ZKsJwlFHqD8lPs8ZVrVYATPpVL1jilTrH6_4TcY5PL_y0l)

The following diagram illustrates the case where failover internet connection is being used (the 4G connection):

![4G Architecture Diagram](https://www.planttext.com/api/plantuml/png/fLBTIyCm47_FNt7u8XGxJJT4HSkGmHHKP8ZlsN9pOMaoqYjpYl-xMVevkpSVMfBVtVq-hXtd91mLcMOQbmHamQcF5O5K3XUah66dy62T-hA6X0Qj3EOt4CVf6Tp6SNYblkKT7Qb5fo7Kzq_AI956U47QMhU6hQ8LZAQCBRGj92X3w0mIC9Q93Pn4qRspGiw5aL5q0YA7p4hCE7mpWBVT1k4kXQR5OpDXrdkC-_FZQpXT9mD-UJNW0yv-6jhHjCkN7F1vEorHqYCQLaIb4H-uSSqgiWEiLdnIflY4cIqyNccxS0dQIAkBdk7aF1dLjQaWt8hTtqIXodiEsqb-R78O9YUu58McqXD4UzTbW5gpXHgXRbUxpGdcQxARAfOCYNhRvwMdeq6Irax7JRcVZiwpdIkDzPSxAfFyHsQCHlgV_nS0)

The configuration file to support the above use case is as follows:

```yaml
maxPacketLoss: 50
minPacketLoss: 20
minSuccessivePacketsRecvd: 20
maxSuccessivePacketsLost: 10
useSudo: true
cleanIfRequired: true
ping: 1.1.1.1
connections:
  - name: fiber
    rank: 1
    ip: 192.168.2.1
    peerIp: 192.168.2.2
    mask: 24
    gwIf: eth0
    gw: 192.168.1.1
    mark: 101
  - name: fourg
    rank: 2
    ip: 192.168.4.1
    peerIp: 192.168.4.2
    mask: 24
    gwIf: eth1
    gw: 192.168.3.1
    mark: 102
```
With the above configuration you will end up with the following network configuration:

![Final Configuration](https://www.planttext.com/api/plantuml/png/dLJlQzim4Fskl-Bmbq5PtDeDeojRQ64iAwobjAFVPVkIYCgIaSyDsyZ_Flsm4xl4TPd0icpTlNll_AohchYXffHCueg0D1YntX0Kmc1EGYls0Nve8_veHLo250hhIvZD5X_XospcfuKDUK938ky5-7rtBHW9aWXtI5jjdc4hQ0Fp9MEvr4q1GX4QXSHOoIieTO5b0dyPmCqzV5r0yZcDaqyNSH8df-dSllwkpPQR0avitKFohfUU7aa_dqndAoqBD61qOrzwY5oNbLQe2ABhgf9Mdkj71Bo6l2oOZMXpegNKcTNUIRnxZ3m0W2E5j3bh7zruhIiDnC9OSi8j_nteXMRulLqVBHb5Evz2Itje7Va7grYeDVpYcSHeZtrhYqNVSCKRILwpeLxTQ-kD5uHs7nm6XZfERXkswqDvU4ZEaXRChdfu_m6Kxe7IIQ1a07JevafaQFXeFzFEDc4CpECLl8RJZLcFmyqdu0vVPzSpry5LWQwNxLTjeRTzwYDjt51pxEVi3UK39zsZhUPvXevD4dMD24JTwdr4NIGRzZBsypZD-IXD_oLwDEqltZgwVYPTdVxabRlTwamurriAnT5ZvoJW7x6LyUU69GxX2Iv1LEJCMREkxx1lMTjmrWc5FSjM8sDNg0fx4Fy3)

As and when required, the active default route will point to Gateway for eth0 or eth1.

## Starting as service or command

To start `IFS` as a service, execute the command `systemctl start goifs.service`.

To enable starting of `IFS` on boot execute the command `systemctl enable goifs.service`.

To see the current status of `IFS` execute the command `goifs status`.

For example, you will see output like so:

```sh
# goifs status
2023/11/09 12:41:23 goifs is running with process id: 1931
2023/11/09 12:41:23 Connection Details::
  Name|     Gateway| Is Up| Successes| Failures| Total| Consecutive Successes| Consecutive Failures|Active
 fiber| 192.168.1.1| false|         0|      100|   100|                     0|                  100|false
 fourg| 192.168.3.1|  true|       100|        0|   100|                   100|                    0|true
```

## Addendum

### Do not use IPv6 in your LAN

In my humble opinion there is no real reason to use IPv6 on your LAN, unless you have a device that only supports only IPv6 (in which case I would prefer to get an alternative device, rather than start using IPv6). Of course, people will suggest that IPv6 is the future, and I agree except that, the future is not near enough.

If you agree and want to get rid of IPv6 on your LAN, read on.

#### Getting rid of IPv6

Getting rid of IPv6 from your desktop and laptop devices is easy enough. There is tons of documentation available on the internet for your particular OS/platform.

The more complicated part is managing to do this for mobile/other devices. Since we have little to no control over how mobile devices configure their network. The only way forward is to block IPv6 access to mobile devices. The whole process involves configuring modems to disable Router Advertisement (RA) for IPv6 and/or to block the RA packets sent by modems from reaching mobile devices.

##### Importance of getting rid of IPv6

You might be wondering why bother getting rid of IPv6. The truth is even if you keep IPv6 around the failover will work for IPv4. However, the IPv6 traffic of your LAN devices will still go via the corresponding modem directly. As a result you might run into a situation where IPv4 is working due to failover, but IPv6 not working because the modem used as IPv6 Gateway has no internet connectivity. Additionally, if you want to restrict internet access using your Linux router, it will allow LAN devices using IPv6 to bypass that.

##### Disabling RA in your modem configuration

This should be easy enough. Just find the relevant configuration and disable it. But do make it a point to validate that modem is actually honouring its configuration using `tcpdump`.

##### Validating that RA is disabled

One can use `tcpdump` to verify that RA is disabled. From your Linux router issue the following command:
`sudo tcpdump -vvvv -ttt -i wlan0 icmp6 and 'ip6[40] = 134'`
replace `wlan0` with the interface for which you want to check.

If after executing this command (within a few minutes) you see no output, you are all good.
If OTOH, you see some output resembling what follows:

```sh
tcpdump: listening on wlan0, link-type EN10MB (Ethernet), snapshot length 262144 bytes
 00:00:00.000000 IP6 (hlim 255, next-header ICMPv6 (58) payload length: 64) fe80::xxxx:xxxx:xxxx:xxxx > ip6-allnodes: [icmp6 sum ok] ICMP6, router advertisement, length 64
        hop limit 64, Flags [other stateful], pref high, router lifetime 30s, reachable time 0ms, retrans timer 0ms
          prefix info option (3), length 32 (4): 2xxx:xxxx:xxxx:xxxx::/64, Flags [onlink, auto], valid time 300s, pref. time 120s
            0x0000:  40c0 0000 012c 0000 0078 0000 0000 2401
            0x0010:  4900 5814 5fd2 0000 0000 0000 0000
          mtu option (5), length 8 (1):  1500
            0x0000:  0000 0000 05dc
          source link-address option (1), length 8 (1): ax:xx:xx:xx:xx:xa
            0x0000:  a42b b02e c0aa
```

your modem is sending RA packets. You can also validate additionally by comparing MAC address mentioned as `source link-address` with the one from your modem.

##### Blocking RA packets from reaching your LAN

To block RA packets sent out from modem from reaching your LAN, simply make your own Linux router sit in front of the modem and drop all packets coming from modem. This can require purchasing of additional hardware to add additional network interfaces in Linux router.

### How does it work?

#### Network Configuration

`IFS` creates some network devices/configuration as per the input configuration file (`/etc/goifs/ifs.yaml`).

1. Create a network namespace for each of the internet connections that are present in configuration file.

    ```sh
    # ip netns list
    fourg (id: 1)
    fiber (id: 0)
    ```

2. Create a pair of virtual devices for each of the connections one in global namespace and other in each connection's network namespace created in step 1. Along with this the `lo` network is also brought up.

    ```sh
    # ip link list
    1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN mode DEFAULT group default qlen 1000
        link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP mode DEFAULT group default qlen 1000
        link/ether xx:xx:xx:xx:xx:xx brd ff:ff:ff:ff:ff:ff
    3: wlan0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP mode DEFAULT group default qlen 1000
        link/ether xx:xx:xx:xx:xx:xx brd ff:ff:ff:ff:ff:ff
    9: fibera@if8: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default qlen 1000
        link/ether xx:xx:xx:xx:xx:xx brd ff:ff:ff:ff:ff:ff link-netns fiber
    11: fourga@if10: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default qlen 1000
        link/ether xx:xx:xx:xx:xx:xx brd ff:ff:ff:ff:ff:ff link-netns fourg
    # ip netns exec fiber ip link list
    1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN mode DEFAULT group default qlen 1000
        link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    8: fiberb@if9: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default qlen 1000
        link/ether xx:xx:xx:xx:xx:xx brd ff:ff:ff:ff:ff:ff link-netnsid 0
    # ip netns exec fourg ip link list
    1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN mode DEFAULT group default qlen 1000
        link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    10: fourgb@if11: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default qlen 1000
        link/ether xx:xx:xx:xx:xx:xx brd ff:ff:ff:ff:ff:ff link-netnsid 0
    ```

3. Configure IP addresses of virtual devices in created in global and network namespaces.

    ```sh
    # ip -4 addr list dev fibera
    9: fibera@if8: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000 link-netns fiber
        inet 192.168.2.1/24 scope global fibera
        valid_lft forever preferred_lft forever

    # ip netns exec fiber ip -4 addr list
    8: fiberb@if9: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000 link-netnsid 0
        inet 192.168.2.2/24 scope global fiberb
        valid_lft forever preferred_lft forever
    # ip -4 addr list dev fourga
    11: fourga@if10: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000 link-netns fourg
        inet 192.168.4.1/24 scope global fourga
        valid_lft forever preferred_lft forever
    # ip netns exec fourg ip -4 addr list
    10: fourgb@if11: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000 link-netnsid 0
        inet 192.168.4.2/24 scope global fourgb
        valid_lft forever preferred_lft forever
    ```

4. Create routes for each of the connections in default and corresponding network namespaces.

    ```sh
    # ip -4 route list
    # ip route list
    default via 192.168.1.2 dev eth0 src 192.168.1.66 metric 202 
    default via 192.168.3.1 dev wlan0 proto dhcp src 192.168.3.100 metric 303 
    192.168.1.0/24 dev eth0 proto dhcp scope link src 192.168.1.66 metric 202 
    192.168.2.0/24 dev fibera proto kernel scope link src 192.168.2.1 
    192.168.3.0/24 dev wlan0 proto dhcp scope link src 192.168.3.100 metric 303 
    192.168.4.0/24 dev fourga proto kernel scope link src 192.168.4.1 
    # ip netns exec fiber ip -4 route list
    default via 192.168.2.1 dev fiberb 
    192.168.2.0/24 dev fiberb proto kernel scope link src 192.168.2.2 
    # ip netns exec fourg ip -4 route list
    default via 192.168.4.1 dev fourgb 
    192.168.4.0/24 dev fourgb proto kernel scope link src 192.168.4.2
    ```

5. Create additional routing table and rule for default route for each of the connections.

    ```sh
    # ip rule list
    0: from all lookup local
    32764: from all fwmark 0x66 lookup 102
    32765: from all fwmark 0x65 lookup 101
    32766: from all lookup main
    32767: from all lookup default
    # ip route list table 101
    default via 192.168.1.2 dev eth0 
    # ip route list table 102
    default via 192.168.3.1 dev wlan0
    ```

6. Create iptable rules to forward, nat and mangle packets as required.

    ```sh
    # iptables -L -v -n
    Chain INPUT (policy ACCEPT 5309K packets, 5759M bytes)
    pkts bytes target     prot opt in     out     source               destination         

    Chain FORWARD (policy ACCEPT 0 packets, 0 bytes)
    pkts bytes target     prot opt in     out     source               destination         
    15M 8255M ACCEPT     all  --  *      *       0.0.0.0/0            0.0.0.0/0           
        0     0 ACCEPT     all  --  fibera eth0    0.0.0.0/0            0.0.0.0/0           
        0     0 ACCEPT     all  --  eth0   fibera  0.0.0.0/0            0.0.0.0/0           
        0     0 ACCEPT     all  --  fourga wlan0   0.0.0.0/0            0.0.0.0/0           
        0     0 ACCEPT     all  --  wlan0  fourga  0.0.0.0/0            0.0.0.0/0           

    Chain OUTPUT (policy ACCEPT 3190K packets, 5550M bytes)
    pkts bytes target     prot opt in     out     source               destination         
    # iptables -L -v -n -t nat
    Chain PREROUTING (policy ACCEPT 535K packets, 51M bytes)
    pkts bytes target     prot opt in     out     source               destination         

    Chain INPUT (policy ACCEPT 24221 packets, 2243K bytes)
    pkts bytes target     prot opt in     out     source               destination         

    Chain OUTPUT (policy ACCEPT 66606 packets, 4545K bytes)
    pkts bytes target     prot opt in     out     source               destination         

    Chain POSTROUTING (policy ACCEPT 7476 packets, 573K bytes)
    pkts bytes target     prot opt in     out     source               destination         
    328K   33M MASQUERADE  all  --  *      *       192.168.1.0/24       0.0.0.0/0           
    261 21924 MASQUERADE  all  --  *      eth0    192.168.2.0/24       0.0.0.0/0           
    126 10584 MASQUERADE  all  --  *      wlan0   192.168.4.0/24       0.0.0.0/0           
    # iptables -L -v -n -t mangle
    Chain PREROUTING (policy ACCEPT 0 packets, 0 bytes)
    pkts bytes target     prot opt in     out     source               destination         
    301 26858 MARK       all  --  fibera *       0.0.0.0/0            0.0.0.0/0            MARK set 0x65
    154 13857 MARK       all  --  fourga *       0.0.0.0/0            0.0.0.0/0            MARK set 0x66

    Chain INPUT (policy ACCEPT 0 packets, 0 bytes)
    pkts bytes target     prot opt in     out     source               destination         

    Chain FORWARD (policy ACCEPT 0 packets, 0 bytes)
    pkts bytes target     prot opt in     out     source               destination         

    Chain OUTPUT (policy ACCEPT 0 packets, 0 bytes)
    pkts bytes target     prot opt in     out     source               destination         

    Chain POSTROUTING (policy ACCEPT 0 packets, 0 bytes)
    pkts bytes target     prot opt in     out     source               destination    
    ```

With this we are all set.

#### How does it work

The `IFS` service keeps pinging configured `Ping` IP using each of the connections periodically to determine if the connection is up. To ping from within the network execute the command: `ip netns exec <namespace> ping -c 1 <ip>`. If the status of a connection changes, the default routes in global namespace are modified to reflect best connection available.
