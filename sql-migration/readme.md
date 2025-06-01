### Readme

This go module/script is used to test sql migration with ephemeral sqlite3, the purpose is to make sure the DDL changes sequences can be properly used, replicated by other contributors dev environment and encourage them to keep sql DDL to be always well documented.

Since it uses sqlite3 as ephememeral SQL db, this should be keep in mind when sql migrations used on actual sql based db that is not sqlite (e.g. postgres, mysql).