workflow "DangerPullRequest" {
  on = "pull_request"
  resolves = ["Danger"]
}

action "Danger" {
  uses = "pei0804/GithubActions/danger@master"
  secrets = ["GITHUB_TOKEN"]
}
