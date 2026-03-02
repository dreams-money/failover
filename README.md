# High Availability Failovers

We achieve high availability through a cluster of database nodes.  A virutal IP (VIP) always points at the primary node which is the main write point of a cluster.  DNS records also point toward the write and read nodes of the cluster.

In the event of a cluster change, this program:

1) Reflects leader election to local routing tables to correctly direct traffic to VIP over VPN.
2) Remote DNS records are also updated to reflect current cluster state.
3) There's an experimental health check feature where it checks other instances of this program to ensure they're all up

Cluster change detection is both on demand and periodic to ensure routing tables and DNS is correct.

## Configuration
Copy config.example.json to config.json to get started

## OPNSense setup
In order to get the correct configuration keys, you need to make a route to the VIP, disable it, and get the route's ID via the developer tools in the browser.

You also have to make a gateway specifically to the host machine for each VIP.

You will also need to make peers and get their IDs via the developer tools like you got the ID for the route you made.

## Omada Setup
- Network Config > Transmission > Routing
- Add the VIP

## DNS Setup
We support Route53 as our DNS store.

To access AWS's SDKv2, it requires API secrets to be held in the system's environment variables.

### Windows PowerShell

```powershell
$env:AWS_ACCESS_KEY_ID="AKsecretZDsecretTNse"
$env:AWS_SECRET_ACCESS_KEY="SdsecretVvsecretSjsecretC9secretJxsecretxv"
$env:AWS_PRIVATE_HOSTED_ZONE_ID="Z0secret3GsecretJKsec"
```

### Linux Shell

```shell
set AWS_ACCESS_KEY_ID="AKsecretZDsecretTNse"
set AWS_SECRET_ACCESS_KEY="SdsecretVvsecretSjsecretC9secretJxsecretxv"
set AWS_PRIVATE_HOSTED_ZONE_ID="Z0secret3GsecretJKsec"
```