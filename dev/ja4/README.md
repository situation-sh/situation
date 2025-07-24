# JA4 tests generator

Getting JA4 ground truth examples is not so easy to get. Here is a python script that generates golang test cases based from pcap files.

The JA4+ suite provides some testing pcap files. In addition, their author release a witeshark/tshark plugin to automatically compute JA4 hashes chen relevent packets are captured. The plugin output is our ground truth.

The script use `pyshark` to reads the pcap files. As the plugin is installed, JA4 hashes are automatically computed. The script generates some tuples `(relevant_bytes, ja4_hash)` for JA4S and JA4X for the moment.

## Pre-requisites

First you must install tshark/wireshark.

```shell
sudo apt install wireshark
```

Then you must download the latest JA4 wireshark plugin from [JA4+ releases](github.com/FoxIO-LLC/ja4/releases) and install it. 
To find the proper install location, look at [their guide](https://github.com/FoxIO-LLC/ja4/tree/main/wireshark#installing-the-plugin). You may use a command like the following to find the plugin path on your system:

```shell
find /usr -type d -name plugins -path '*/wireshark/plugins' 2>/dev/null
```

## Generate tests

You can clone the JA4 repo to get the pcap files
```shell
git clone https://github.com/FoxIO-LLC/ja4.git /tmp/ja4
```

Then run the script (beware of the final location).
```shell
./generate_tests.py --ja4s=../../modules/ja4/ja4s_test.go --ja4x=../../modules/ja4/ja4x_test.go /tmp/ja4/pcap/*
```