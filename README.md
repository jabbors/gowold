# gowold

A simple wake-on-lan (wol) service to wake up a single device in a network. Once the initial magic packet has been sent, it will be re-transmitted periodically to keep the device awake as some devices supporting wol will shut down after a few minutes unless the magic packet is re-transmitted.

The service provides a WebUI with a single button to start/stop the wol agent. Additionally it offers an API endpoint to pull the current status.

## API Documentation

### /status

Returns the current status:

- An 204 No Content response is returned when the agent is stopped
- An 200 OK response is returned when the agent is active along with a small JSON payload

```
{
    "started_at": "1632244377",
    "last_broadcast": "1632244377"
}
```

- `started_at` specifies the timestamp when the wol agent was started
- `last_broadcast` specifies the timestamp when the last wol packet was sent

## Usage

Run this with docker using host network mode in order for the wol packets to be sent on the host network

`docker run --rm -d --network host -e TARGET_MAC=00:11:22:33:44:55 jabbors/gowold`
