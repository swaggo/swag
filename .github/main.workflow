workflow "DangerPullRequest" {
  on = "pull_request"
  resolves = ["Danger"]
}

action "Danger" {
  uses = "./.github/actions/danger"
  secrets = ["GITHUB_TOKEN"]
}
