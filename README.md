```
 dP""b8 88  88    db    888888 88   88 88 
dP   `" 88  88   dPYb     88   88   88 88 
Yb  "88 888888  dP__Yb    88   Y8   8P 88 
 YboodP 88  88 dP""""Yb   88   `YbodP' 88 
```

Simple TUI chat, written from simple CLI chat, written in 10 days to learn GO and prepare for a backend Go developer interview

Server usage:
- systemctl start postgresql
- createdb chat
- psql -d chat -a -f ./init-db.sql
- ./server [port]

Client usage:
- ./client [ip:port]

