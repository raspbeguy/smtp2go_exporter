# smtp2go_exporter

A Prometheus esporter for [smtp2go](https://www.smtp2go.com/) statistics.

## Usage

You should get an [API key](https://developers.smtp2go.com/docs/getting-started).

Example usage:

```
./smtp2go_exporter -api-url https://eu-api.smtp2go.com/v3/ -api-key <your API key>
```

Example metrics:

```
# HELP smtp2go_email_bounces_bounce_percent Percentage of bounced emails
# TYPE smtp2go_email_bounces_bounce_percent gauge
smtp2go_email_bounces_bounce_percent 0
# HELP smtp2go_email_bounces_emails Number of emails processed
# TYPE smtp2go_email_bounces_emails gauge
smtp2go_email_bounces_emails 414
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
smtp2go_email_cycle_remaining 478
# HELP smtp2go_email_cycle_remaining_seconds Seconds remaining until the end of the current cycle
# TYPE smtp2go_email_cycle_remaining_seconds gauge
smtp2go_email_cycle_remaining_seconds 747321.76931955
# HELP smtp2go_email_cycle_used Number of emails used in the current cycle
# TYPE smtp2go_email_cycle_used gauge
smtp2go_email_cycle_used 522
# HELP smtp2go_email_history_avgsize Average size of emails per email address
# TYPE smtp2go_email_history_avgsize gauge
smtp2go_email_history_avgsize{email_address="alice@example.tld""} 7374.04914004914
smtp2go_email_history_avgsize{email_address="bob@example.tld""} 20483.428571428572
# HELP smtp2go_email_history_bounces Number of bounces per email address
# TYPE smtp2go_email_history_bounces gauge
smtp2go_email_history_bounces{email_address="alice@example.tld""} 0
smtp2go_email_history_bounces{email_address="bob@example.tld""} 0
# HELP smtp2go_email_history_bytecount Total size in bytes of emails sent per email address
# TYPE smtp2go_email_history_bytecount gauge
smtp2go_email_history_bytecount{email_address="alice@example.tld""} 3.001238e+06
smtp2go_email_history_bytecount{email_address="bob@example.tld""} 143384
# HELP smtp2go_email_history_clicks Number of clicks per email address
# TYPE smtp2go_email_history_clicks gauge
smtp2go_email_history_clicks{email_address="alice@example.tld""} 0
smtp2go_email_history_clicks{email_address="bob@example.tld""} 0
# HELP smtp2go_email_history_opens Number of opens per email address
# TYPE smtp2go_email_history_opens gauge
smtp2go_email_history_opens{email_address="alice@example.tld""} 0
smtp2go_email_history_opens{email_address="bob@example.tld""} 0
# HELP smtp2go_email_history_rejects Number of rejected emails per email address
# TYPE smtp2go_email_history_rejects gauge
smtp2go_email_history_rejects{email_address="alice@example.tld""} 0
smtp2go_email_history_rejects{email_address="bob@example.tld""} 0
# HELP smtp2go_email_history_spam Number of spam reports per email address
# TYPE smtp2go_email_history_spam gauge
smtp2go_email_history_spam{email_address="alice@example.tld""} 0
smtp2go_email_history_spam{email_address="bob@example.tld""} 0
# HELP smtp2go_email_history_unsubscribes Number of unsubscribes per email address
# TYPE smtp2go_email_history_unsubscribes gauge
smtp2go_email_history_unsubscribes{email_address="alice@example.tld""} 0
smtp2go_email_history_unsubscribes{email_address="bob@example.tld""} 0
# HELP smtp2go_email_history_used Number of emails used per email address
# TYPE smtp2go_email_history_used gauge
smtp2go_email_history_used{email_address="alice@example.tld""} 407
smtp2go_email_history_used{email_address="bob@example.tld""} 7
```

## TODO

* Implement other API methods
