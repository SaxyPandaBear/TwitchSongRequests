Managing User Access
====================

> If you believe that I revoked your access in error, please feel free to open a GitHub issue to appeal it, otherwise you'll want to submit a new onboarding request.

### Connectiong to the Postgres database
This is mostly documenting for myself, especially since Railway removed their
web UI for running queries on the connected database so I have to do everything
from my terminal now.
```bash
# Assumes psql is installed, and the below variables mirror names of variables provided by Railway
psql -h $PGHOST -p $PGPORT -d railway -U $PGUSER
# type in password when prompted
```

### Getting Twitch IDs by user login
```bash
twitch api get users -q login=$LOGIN
```

### Getting the user's info from the db
```sql
select * from users where id='123456';
```

### Queries ran on the database to get Twitch IDs
Getting previously "active" users
```sql
select distinct cast(broadcaster_id as integer) from messages 
  where broadcaster_id != '' 
  and success = 1 
  and age(messages.created_at) > 30 * INTERVAL '1 day' 
except select distinct cast(broadcaster_id as integer) from messages 
  where broadcaster_id != '' 
  and age(messages.created_at) <= 30 * INTERVAL '1 day'
order by broadcaster_id asc;
```

So the criteria are:
1. Had at least 1 successful song request redeemed more than 30 days ago
1. Has not had ANY redeems in the past 30 days

This allows for errors such as issues with the API, credentials, etc in the past 30 days, because
at least it was attempted.

Getting users who signed up but never used it, because I need to be more cutthroat now.
If you are having issues using it, please cut a GitHub issue so I know you're at least trying
```sql
select distinct cast(broadcaster_id as integer) from messages
  where broadcaster_id != '' 
  and success = 0 
except select distinct cast(broadcaster_id as integer) from messages 
  where success = 1;
```

### API call to get usernames
```bash
# for each user ID
twitch token
twitch api get users -q id=$ID
```

### Full flow
1. Run the SQL query to get the list of IDs
1. Copy and save the output list of IDs to a text file: `/tmp/ids`
  1. `rm /tmp/ids`
  1. `vim /tmp/ids`
1. Run the bash script below to get the Twitch ID mapped to their username
1. Copy the onboarded users from the Spotify dashboard into a file: `/tmp/onboarded_users`
1. Run the python script below to figure out who needs to get their access revoked

```bash
twitch token
rm /tmp/twitch_users
cat /tmp/ids | while read line || [[ -n $line ]];
do
  twitch api get users -q id=$line | jq '.data | .[] | (.id, .display_name)' >> /tmp/twitch_users
done
```

```python
users = set()
with open("/tmp/twitch_users") as usersfile:
    cnt = 0
    for line in usersfile.readlines():
        if cnt % 2 == 1:
            users.add(line.replace('"', '').strip().lower())
        cnt += 1

onboarded = set()
with open("/tmp/onboarded_users") as onboardedfile:
    # every 2nd line of a grouping of 4 contains the username
    cnt = 0
    for line in onboardedfile.readlines():
        if cnt > 3:
            cnt = 0
        if cnt == 1:
            onboarded.add(line.strip().lower())
        cnt += 1

for username in users.intersection(onboarded):
    print(username)
```
