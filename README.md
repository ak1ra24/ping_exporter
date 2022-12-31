# ping exporter (WIP)
Prometheus exporter for ICMP echo requests using https://github.com/ak1ra24/pro-bing forked from https://github.com/prometheus-community/pro-bing

**Creating to support unicast and broadcast**

## Getting Started
### Config file
[example config file](example/config_example.yaml)


## Metrics
| Name            | Description                       |
| --------------- | --------------------------------- |
| ping_sent_count | The number of send ping packet    |
| ping_recv_count | The number of receive ping packet |
| ping_loss_count | The number of loss packet         |
| ping_last_rtt   | last rtt                          |

### Label
| Name        | Description                                               |
| ----------- | --------------------------------------------------------- |
| broadcast   | Broadcast or not                                          |
| description | Description in the configuration file                     |
| ip          | IP addresses actually sent and received                   |
| target      | IP in the configuration file                              |
| name        | name in the configuration file                            |
| loss_reason | The reason of loss packets only for the loss_count metric |

## Running non-root user
Refer to https://github.com/prometheus-community/pro-bing#linux