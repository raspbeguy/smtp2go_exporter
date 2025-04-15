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
# HELP smtp2go_email_bounces_bounce_percent Percentage of bounced emails
# TYPE smtp2go_email_bounces_bounce_percent gauge
smtp2go_email_bounces_bounce_percent 0
# HELP smtp2go_email_bounces_emails Number of emails processed
# TYPE smtp2go_email_bounces_emails gauge
smtp2go_email_bounces_emails 412
# HELP smtp2go_email_bounces_hardbounces Number of hard bounces
# TYPE smtp2go_email_bounces_hardbounces gauge
smtp2go_email_bounces_hardbounces 0
# HELP smtp2go_email_bounces_rejects Number of rejected emails
# TYPE smtp2go_email_bounces_rejects gauge
smtp2go_email_bounces_rejects 108
# HELP smtp2go_email_bounces_softbounces Number of soft bounces
# TYPE smtp2go_email_bounces_softbounces gauge
smtp2go_email_bounces_softbounces 0
# HELP smtp2go_email_cycle_max Maximum number of emails allowed in the current cycle
# TYPE smtp2go_email_cycle_max gauge
smtp2go_email_cycle_max 1000
# HELP smtp2go_email_cycle_remaining Number of emails remaining in the current cycle
# TYPE smtp2go_email_cycle_remaining gauge
smtp2go_email_cycle_remaining 480
# HELP smtp2go_email_cycle_remaining_seconds Seconds remaining until the end of the current cycle
# TYPE smtp2go_email_cycle_remaining_seconds gauge
smtp2go_email_cycle_remaining_seconds 748713.54958255
# HELP smtp2go_email_cycle_used Number of emails used in the current cycle
# TYPE smtp2go_email_cycle_used gauge
smtp2go_email_cycle_used 520
```

## TODO

* Implement other API methods
