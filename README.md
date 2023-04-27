# slack-backup

A simple script to back up slack channels.

## Usage
```bash
# clone this repo
git clone https://github.com/beewee22/slack-backup.git
# Install dependencies
go get -u
# Build
go build
# set slack token to environment variable
export SLACK_BACKUP_TOKEN=<<xoxb-your-slack-app-token>>

# run
./slack-backup <<slack-channel-id>> # messages.json, threads.json will be created
```

