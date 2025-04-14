# smtp2go_exporter

A Prometheus esporter for [smtp2go](https://www.smtp2go.com/) statistics.

## Usage

You should get an [API key](https://developers.smtp2go.com/docs/getting-started).

Exemple usage:

```
./smtp2go_exporter -api-url https://eu-api.smtp2go.com/v3/ -api-key <your API key>
```

Exemple metrics:

```
# HELP smtp2go_email_cycle_max Maximum number of emails by cycle
# TYPE smtp2go_email_cycle_max gauge
smtp2go_email_cycle_max 1000
# HELP smtp2go_email_cycle_remaining Remaining number of emails for this cycle
# TYPE smtp2go_email_cycle_remaining gauge
smtp2go_email_cycle_remaining 525
# HELP smtp2go_email_cycle_remaining_seconds Seconds remaining until end of this cycle
# TYPE smtp2go_email_cycle_remaining_seconds gauge
smtp2go_email_cycle_remaining_seconds 812311.400687645
# HELP smtp2go_email_cycle_used Used number of email for this cycle
# TYPE smtp2go_email_cycle_used gauge
smtp2go_email_cycle_used 475
```

## TODO

* Implement other API methods
