# Test project to play with fuse

## Links
 * https://pkg.go.dev/bazil.org/fuse?tab=importedby
 * https://github.com/0xmohit/rclone/blob/master/cmd/mount/file.go
 * https://github.com/bazil/fuse/blob/master/examples/clockfs/clockfs.go
 * https://blog.gopheracademy.com/advent-2014/fuse-zipfs/

## Prepare

```bash
docker run --rm --name db -p 3306:3306 -e MYSQL_USER=root -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=db mysql/mysql-server
docker container exec -it db bash
mysql -u root -p
```

```sql
SELECT host, user FROM mysql.user;
UPDATE mysql.user SET Host='%' WHERE Host='localhost' AND User='root';
FLUSH PRIVILEGES;
```

## Run

```bash
mkdir /tmp/test
go run . /tmp/test

tree /tmp/test
umount /tmp/test
```

## Example Usage
```bash
$ ll
insgesamt 0
dr-xr-xr-x 1 root root 0 Okt 17 21:34 db/
dr-xr-xr-x 1 root root 0 Okt 17 21:34 information_schema/
dr-xr-xr-x 1 root root 0 Okt 17 21:34 mysql/
dr-xr-xr-x 1 root root 0 Okt 17 21:34 performance_schema/
dr-xr-xr-x 1 root root 0 Okt 17 21:34 sys/

$ ll mysql/
insgesamt 0
dr-xr-xr-x 1 root root 0 Okt 17 21:34 columns_priv/
dr-xr-xr-x 1 root root 0 Okt 17 21:34 component/
dr-xr-xr-x 1 root root 0 Okt 17 21:34 db/
...
dr-xr-xr-x 1 root root 0 Okt 17 21:34 time_zone_transition/
dr-xr-xr-x 1 root root 0 Okt 17 21:34 time_zone_transition_type/
dr-xr-xr-x 1 root root 0 Okt 17 21:34 user/

$ ll mysql/user/
insgesamt 0
-r--r--r-- 1 root root 1,9K Okt 17 21:34 all
dr-xr-xr-x 1 root root    0 Okt 17 21:34 by/

$ ll mysql/user/by/
insgesamt 0
dr-xr-xr-x 1 root root 0 Okt 17 21:34 account_locked/
...
dr-xr-xr-x 1 root root 0 Okt 17 21:34 Host/
...
dr-xr-xr-x 1 root root 0 Okt 17 21:34 User_attributes/
dr-xr-xr-x 1 root root 0 Okt 17 21:34 x509_issuer/
dr-xr-xr-x 1 root root 0 Okt 17 21:34 x509_subject/

$ ll mysql/user/by/Host/
insgesamt 0
-r--r--r-- 1 root root  970 Okt 17 21:34 %
-r--r--r-- 1 root root 1,7K Okt 17 21:34 localhost

$ cat mysql/user/by/Host/localhost 
Host,User,Select_priv,Insert_priv,Update_priv, ...
localhost,healthchecker,N,N,N, ...
localhost,mysql.infoschema,Y,N,N, ...
localhost,mysql.session,N,N,N, ...
localhost,mysql.sys,N,N,N, ...
```