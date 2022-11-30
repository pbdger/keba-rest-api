# Keba-Rest-Api image
Provides Rest-API for a Keba Wallbox as container image.

### How to start ?
Start your container binding the external port 8080.

```
docker run -d --name=kebarestapi -p 8080:8080 \
            -e wallboxName=<your wallbox ip> \
            -e apiPort=<the port this process listens, e.g. 8080> \
            pbdger/keba-rest-api
```
Try it out.

### How to get state via rest api ?

Open in a browser the URL with your servername and the metric port.

> http://localhost:8080/state

This request returns an output like this

```
{
    "active_power_mW": 11380,
    "cable_state": 7,
    "charged_energy_Wh": 1040,
    "charging_state": 1,
    "error_code": 0,
    "firmware_version": 4711,
    "max_charging_current_mAh": 18900,
    "max_supported_current_mAh": 19000,
    "power_factor_percent": 98,
    "product_type_and_features": 456491,
    "serial_number": 46589548,
    "total_energy_counter_Wh": 938872,
    "phase": [
        {
            "charging_mAh": 10,
            "voltage_V": 228
        },
        {
            "charging_mAh": 10,
            "voltage_V": 227
        },
        {
            "charging_mAh": 10,
            "voltage_V": 228
        }
    ],
    "Timestamp": "2022-11-30T16:36:44+01:00"
}
```


### Additional optional environment parameters
> debug: true | false

> wallboxPort: number, default is 502 

## Core binary usage
###Prerequisites
#### Mandatory
Set the mandatory environment variable (This example is Linux based);

```
export wallboxPort=<IP or servername of your wallbox, e.g. 192.168.08.15>
```

#### Optional
Set the optional environment variables (This example is Linux based);

```
export wallboxPort=<Port on which your TCP/modbus listens. Default is 502>
```

#### Call on console
```
keba-rest-api
```



### How to build your own version ?

```
GOOS=windows GOARCH=amd64 go build -o ./bin/keba-rest-api.exe keba-rest-api.go
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/keba-rest-api.linux keba-rest-api.go
```
