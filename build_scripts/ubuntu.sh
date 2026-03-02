cd ..
mkdir tmp
go build -o tmp/ha_network_failover main.go
cp config.example.json tmp/
mkdir tmp/storage
cd tmp
tar -czvf ../ubuntu-$0.tar.gz .
cd ..
rm -rf tmp/