# Tesla Solar fetcher

A tool to fetch your solar energy use from Tesla portal. The current
output is formatted to be uploaded to [InfluxDB]

## Limitations

- Only collects the current days data (tesla doesn't provide historical
  hourly data).
- Assumes you only have a solar setup and no other tesla products (
  fixable but don't have a system to test it on)
- Only outputs solar data but could be made to output battery or grid
  usage.

## Usage

```sh
> EMAIL=<your username> PASSWORD=<your password> go run main.go 2>/dev/null >data
> curl -XPOST http://influxdb:8086/write?db=whatever --data-binary @data
```

[InfluxDB]: https://www.influxdata.com/
