# pubsubclip

publish / subscribe clipboard via redis

## install

### Arch Linux

```bash
git clone https://github.com/mohemohe/pubsubclip.git
cd pubsubclip
makepkg -sif
# edit /etc/default/pubsubclip
sudo systemctl start pubsubclip@$(whoami) 
```

### from release

Download from https://github.com/mohemohe/pubsubclip/releases and unarchive tar.gz.

### go install

Require: go 1.22+

```bash
go install github.com/mohemohe/pubsubclip@latest
```

## usage

```bash
./pubsubclip watch --addr 172.16.34.200:6379
```
