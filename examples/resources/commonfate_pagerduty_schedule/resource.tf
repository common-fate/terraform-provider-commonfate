
resource "commonfate_pagerduty_schedule" "schedules" {
  id = data.commonfate_pagerduty_schedule.on-call-users.id
  name = data.commonfate_pagerduty_schedule.on-call-users.name
}