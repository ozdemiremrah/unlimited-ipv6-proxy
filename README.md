# unlimited-ipv6-proxy
[for educational purposes] This simple proxy server can generate an infinite number of proxies while adhering to the subnet limit.

## About this project
This project has been tested on Ubuntu installed server.

The logic of the project is quite simple, it instantly identifies the requested ip address to the server in the **x-proxy-ip** parameter sent with the header during the **http/https request** and uses it at that moment. Then it deletes the unused IP address.
