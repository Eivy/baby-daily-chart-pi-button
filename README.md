# Baby Daily Chart Pi Button

Send request for [BabyDailyChart](https://babydailychart.firebaseapp.com) using RaspberryPi GPIO.

## Usage

Copy `babydailychartbutton.service` to `/etc/systemd/system/` and `baby-daily-chart-pi-button` to `/usr/local/bin`.
Then set Environment Value `BABY_USER_ID` for service.

After all, just start service `babydailychartbutton`.

### Pin Layout

| GPIO Pin | BabyDailyChart button number |
|---|---|
| 4 | 1 |
| 17 | 2 |
| 24 | 3 |

And `14` GPIO is used for showing error status.
